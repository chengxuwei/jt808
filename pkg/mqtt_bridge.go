package pkg

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

// RecvTopic MQTT 上行主题格式 recv/{terminalNo}/{msgIdHex}；注释原因：统一上行 Topic 命名；注释人：Cursor
func RecvTopic(terminalNo string, msgID MsgId) string {
	return fmt.Sprintf("recv/%s/%04X", terminalNo, uint16(msgID))
}

// MQTTBridge 基于内嵌 MQTT Broker 的桥接实现；注释原因：实现 MessageBridge，可替换为 Kafka/RabbitMQ/HTTP 等；注释人：Cursor
type MQTTBridge struct {
	srv *MQTTServer
	cfg MQTTCfg
}

// NewMQTTBridge 创建 MQTT 桥接器；注释原因：cfg 由 main 从 AppConfig.MQTT 注入，地址/开关均可配置；注释人：Cursor
func NewMQTTBridge(cfg MQTTCfg) *MQTTBridge {
	return &MQTTBridge{srv: &MQTTServer{}, cfg: cfg}
}

// Start 启动 MQTT Broker（非阻塞）；注释原因：StartMqttBroker 内部调用 server.Serve() 阻塞，需 goroutine 运行；注释人：Cursor
func (b *MQTTBridge) Start() error {
	go b.srv.StartMqttBroker(b.cfg.Addr)
	return nil
}

// Stop 停止 MQTT Broker；注释原因：预留优雅关闭接口；注释人：Cursor
func (b *MQTTBridge) Stop() error {
	// TODO: 调用 mochi-mqtt server.Close() 优雅关闭
	return nil
}

// PublishUplink 将上行事件序列化为 JSON 并发布到 MQTT；注释原因：实现 MessageBridge.PublishUplink；注释人：Cursor
func (b *MQTTBridge) PublishUplink(event *UplinkEvent) error {
	if b.srv == nil {
		return fmt.Errorf("MQTTBridge: MQTTServer 未初始化")
	}
	out, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("MQTTBridge: 序列化 UplinkEvent 失败: %w", err)
	}
	topic := RecvTopic(event.TerminalNo, MsgId(event.MsgID))
	if err := b.srv.Publish(topic, out); err != nil {
		return fmt.Errorf("MQTTBridge: MQTT Publish 失败 topic=%s: %w", topic, err)
	}
	slog.Info("[上行-5] MQTT 上行发布成功",
		slog.String("topic", topic),
		slog.String("terminalNo", event.TerminalNo),
		slog.String("msgId", event.MsgIDHex),
		slog.String("msgName", event.MsgName),
		slog.Int("payloadLen", len(out)),
	)
	return nil
}

// mqttDownEnvelope MQTT 下行消息包装格式；注释原因：兼容平台发布完整 envelope 或裸 JSON 两种格式；注释人：Cursor
type mqttDownEnvelope struct {
	TerminalNo string          `json:"terminalNo"`
	MsgID      uint16          `json:"msgId"`
	MsgIDHex   string          `json:"msgIdHex"`
	Data       json.RawMessage `json:"data"`
}

func parseTopicMsgIDHex(s string) (MsgId, error) {
	s = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(s)), "0x")
	v, err := strconv.ParseUint(s, 16, 16)
	if err != nil {
		return 0, err
	}
	return MsgId(v), nil
}

// HandleMQTTSendTopic MQTT 下行适配器；注释原因：仅负责解析 MQTT 特有的 topic 和 envelope 格式，解耦传输与业务；注释人：Cursor
// 主题格式：send/{terminalNo}/{msgIdHex}（十六进制，如 8300）
// 解析成功后构建 DownlinkCmd，交由传输无关的 DispatchDownlink 执行
func HandleMQTTSendTopic(topic string, payload []byte) {
	// [下行-2] 解析 MQTT 主题，提取 terminalNo 和 msgId
	parts := strings.Split(strings.Trim(topic, "/"), "/")
	if len(parts) != 3 || parts[0] != "send" {
		slog.Warn("[下行-2] 主题格式无效，期望 send/{terminalNo}/{msgIdHex}",
			slog.String("topic", topic),
		)
		return
	}
	terminalNo := parts[1]
	msgID, err := parseTopicMsgIDHex(parts[2])
	if err != nil {
		slog.Warn("[下行-2] 主题中 msgId 无法解析为十六进制",
			slog.String("topic", topic),
			slog.String("msgIdRaw", parts[2]),
			slog.Any("err", err),
		)
		return
	}
	msgIdHex := fmt.Sprintf("%04X", uint16(msgID))
	slog.Info("[下行-2] MQTT 主题解析成功，收到原始 JSON",
		slog.String("topic", topic),
		slog.String("terminalNo", terminalNo),
		slog.String("msgId", msgIdHex),
		slog.String("msgName", msgID.String()),
		slog.Int("payloadLen", len(payload)),
		slog.String("rawJSON", string(payload)),
	)

	// 解析 MQTT envelope，优先取 data 字段；注释原因：平台可发 {data:{...}} 或裸 JSON 两种格式；注释人：Cursor
	var env mqttDownEnvelope
	_ = json.Unmarshal(payload, &env)
	data := json.RawMessage(payload)
	if len(env.Data) > 0 && string(env.Data) != "null" {
		data = env.Data
	}
	slog.Info("[下行-2] 提取到编码器 data 字段",
		slog.String("terminalNo", terminalNo),
		slog.String("msgId", msgIdHex),
		slog.String("data", string(data)),
	)
	if env.MsgID != 0 && MsgId(env.MsgID) != msgID {
		slog.Warn("[下行-2] JSON 中 msgId 与主题不一致，以主题为准",
			slog.String("topicMsgId", msgIdHex),
			slog.String("jsonMsgId", fmt.Sprintf("%04X", env.MsgID)),
		)
	}

	// 构建标准下行指令，交由传输无关的核心分发
	DispatchDownlink(&DownlinkCmd{
		TerminalNo: terminalNo,
		MsgID:      msgID,
		Data:       data,
	})
}
