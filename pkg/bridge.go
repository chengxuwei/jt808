package pkg

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"sync/atomic"
)

// DownlinkCmd 下行指令标准化结构；注释原因：屏蔽 MQTT/Kafka/HTTP 等传输层差异，统一下行调度入口；注释人：Cursor
type DownlinkCmd struct {
	TerminalNo string          // 目标终端号
	MsgID      MsgId           // 消息ID
	Data       json.RawMessage // 已提取的消息体 JSON，直接反序列化到 Encoder
}

// UplinkEvent 上行事件标准化结构；注释原因：屏蔽传输层差异，统一上行发布出口；注释人：Cursor
type UplinkEvent struct {
	TerminalNo string          `json:"terminalNo"`
	MsgID      uint16          `json:"msgId"`
	MsgIDHex   string          `json:"msgIdHex"`
	MsgName    string          `json:"msgName,omitempty"`
	Data       json.RawMessage `json:"data"`
}

// MessageBridge JT808 消息桥接器接口；注释原因：对接不同 MQ/API 时只需实现此接口，无需修改核心逻辑；注释人：Cursor
// 上行：JT808终端 → PublishUplink → MQTT/Kafka/HTTP 等
// 下行：MQTT/Kafka/HTTP → 解析为 DownlinkCmd → DispatchDownlink → JT808终端
type MessageBridge interface {
	// PublishUplink 将终端上行事件推送到外部系统（MQTT/Kafka/HTTP 等）
	PublishUplink(event *UplinkEvent) error
	// Start 启动桥接器（非阻塞，内部订阅/监听下行指令并调用 DispatchDownlink）
	Start() error
	// Stop 停止桥接器并释放资源
	Stop() error
}

// globalBridge 全局桥接器，启动时通过 SetBridge 注入；注释原因：避免包循环依赖，支持运行时切换；注释人：Cursor
var globalBridge atomic.Value // 存储 MessageBridge

// SetBridge 注入全局桥接器；注释原因：main 启动时调用一次，注入具体实现；注释人：Cursor
func SetBridge(b MessageBridge) {
	if b != nil {
		globalBridge.Store(b)
	}
}

func loadBridge() MessageBridge {
	v := globalBridge.Load()
	if v == nil {
		return nil
	}
	return v.(MessageBridge)
}

// NotifyUplink 将 JT808 解码结果序列化并通过全局桥接器发布；注释原因：统一上行入口，由 JT808Server 调用；注释人：Cursor
func NotifyUplink(terminalNo string, msgID MsgId, decoder JT808Decoder) {
	b := loadBridge()
	if b == nil {
		slog.Warn("[上行-5] 桥接器未注册，跳过发布",
			slog.String("terminalNo", terminalNo),
			slog.String("msgId", fmt.Sprintf("%04X", uint16(msgID))),
		)
		return
	}
	data, err := json.Marshal(decoder)
	if err != nil {
		slog.Error("[上行-5] 序列化解码结果失败",
			slog.String("terminalNo", terminalNo),
			slog.String("msgId", fmt.Sprintf("%04X", uint16(msgID))),
			slog.Any("err", err),
		)
		return
	}
	event := &UplinkEvent{
		TerminalNo: terminalNo,
		MsgID:      uint16(msgID),
		MsgIDHex:   fmt.Sprintf("%04X", uint16(msgID)),
		MsgName:    msgID.String(),
		Data:       data,
	}
	if err := b.PublishUplink(event); err != nil {
		slog.Error("[上行-5] 桥接器 PublishUplink 失败",
			slog.String("terminalNo", terminalNo),
			slog.String("msgId", event.MsgIDHex),
			slog.Any("err", err),
		)
	}
}

// applyEncoderTerminalAndMsgID 通过反射将 terminalNo、msgID、SeqNo 注入 Encoder 的嵌入 JTMessage 字段；
// 注释原因：统一在此处为所有经 bridge 的下行消息分配平台侧序号，避免各 Encoder 各自处理；注释人：Cursor
func applyEncoderTerminalAndMsgID(enc JT808Encoder, terminalNo string, msgID MsgId) {
	rv := reflect.ValueOf(enc)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return
	}
	rv = rv.Elem()
	jt := rv.FieldByName("JTMessage")
	if !jt.IsValid() || jt.Kind() != reflect.Struct {
		return
	}
	if f := jt.FieldByName("TerminalNo"); f.IsValid() && f.CanSet() && terminalNo != "" {
		f.SetString(terminalNo)
	}
	if f := jt.FieldByName("MsgID"); f.IsValid() && f.CanSet() {
		f.Set(reflect.ValueOf(msgID).Convert(f.Type()))
	}
	// 平台下行序号：每个终端独立递增，与终端上行序号互相独立；注释原因：JT808 规范要求平台侧自维护流水号；注释人：Cursor
	if f := jt.FieldByName("SeqNo"); f.IsValid() && f.CanSet() {
		f.SetUint(uint64(NextSeqNo(terminalNo)))
	}
}

// DispatchDownlink 传输无关的下行分发核心；注释原因：所有桥接实现（MQTT/Kafka/HTTP）共用此函数，不含任何传输层代码；注释人：Cursor
// 步骤：[下行-3] 会话 → [下行-4] 编码器 → [下行-5] JSON反序列化 → [下行-6] 帧编码 → [下行-7] 写入TCP连接
func DispatchDownlink(cmd *DownlinkCmd) {
	terminalNo := cmd.TerminalNo
	msgID := cmd.MsgID
	msgIdHex := fmt.Sprintf("%04X", uint16(msgID))

	// [下行-3] 查找终端会话
	v, ok := sessionMap.Load(terminalNo)
	if !ok {
		slog.Warn("[下行-3] 终端未在线，无 JT808 会话",
			slog.String("terminalNo", terminalNo),
			slog.String("msgId", msgIdHex),
		)
		return
	}
	session := v.(*Session)
	if session.Conn == nil {
		slog.Warn("[下行-3] 终端会话连接为空",
			slog.String("terminalNo", terminalNo),
			slog.String("msgId", msgIdHex),
		)
		return
	}
	slog.Info("[下行-3] 找到终端会话",
		slog.String("terminalNo", terminalNo),
		slog.String("remote", session.Conn.RemoteAddr().String()),
	)

	// [下行-4] 获取编码器（工厂模式，每次新建实例，线程安全）
	encoder := GetEncoder(msgID)
	if encoder == nil {
		slog.Warn("[下行-4] 未注册该消息 ID 的编码器",
			slog.String("terminalNo", terminalNo),
			slog.String("msgId", msgIdHex),
		)
		return
	}
	slog.Info("[下行-4] 编码器就绪",
		slog.String("terminalNo", terminalNo),
		slog.String("msgId", msgIdHex),
		slog.String("encoder", fmt.Sprintf("%T", encoder)),
	)

	// [下行-5] JSON 反序列化到编码器结构体
	if err := json.Unmarshal(cmd.Data, encoder); err != nil {
		slog.Error("[下行-5] JSON 反序列化到编码器失败",
			slog.String("terminalNo", terminalNo),
			slog.String("msgId", msgIdHex),
			slog.String("data", string(cmd.Data)),
			slog.Any("err", err),
		)
		return
	}
	slog.Info("[下行-5] JSON 反序列化成功",
		slog.String("terminalNo", terminalNo),
		slog.String("msgId", msgIdHex),
		slog.Int("dataLen", len(cmd.Data)),
	)

	// [下行-6] 用 topic 中的 terminalNo/msgId 覆盖结构体，确保帧头正确；注释原因：JSON 中可能漏填；注释人：Cursor
	applyEncoderTerminalAndMsgID(encoder, terminalNo, msgID)
	frame := encoder.Encode()
	if len(frame) == 0 {
		slog.Warn("[下行-6] Encoder 返回空帧",
			slog.String("terminalNo", terminalNo),
			slog.String("msgId", msgIdHex),
		)
		return
	}
	frameHex := hex.EncodeToString(frame)
	slog.Info("[下行-6] JT808 帧编码完成",
		slog.String("terminalNo", terminalNo),
		slog.String("msgId", msgIdHex),
		slog.Int("frameLen", len(frame)),
		slog.String("frameHex", frameHex),
	)

	// [下行-7] 发送到终端 TCP 连接
	n, err := session.Conn.Write(frame)
	if err != nil {
		slog.Error("[下行-7] 发送到终端失败",
			slog.String("terminalNo", terminalNo),
			slog.String("msgId", msgIdHex),
			slog.String("remote", session.Conn.RemoteAddr().String()),
			slog.String("frameHex", frameHex),
			slog.Any("err", err),
		)
		return
	}
	slog.Info("[下行-7] 发送到终端成功",
		slog.String("terminalNo", terminalNo),
		slog.String("msgId", msgIdHex),
		slog.String("msgName", msgID.String()),
		slog.String("remote", session.Conn.RemoteAddr().String()),
		slog.Int("bytes", n),
		slog.String("frameHex", frameHex),
	)
}
