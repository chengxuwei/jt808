package pkg

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
)

// globalMQTTBroker 供 JT808 与 MQTT 回调共用；原因：解析结果需异步发布且避免循环依赖；注释人：Cursor
var globalMQTTBroker atomic.Pointer[MQTTServer]

// SetMQTTBroker 在启动 MQTT 前由 main 注入；原因：保证 Publish 能拿到同一 Server 实例；注释人：Cursor
func SetMQTTBroker(m *MQTTServer) {
	if m != nil {
		globalMQTTBroker.Store(m)
	}
}

// RecvTopic MQTT 上行主题 recv/{terminalNo}/{msgIdHex}；原因：与约定 Topic 格式一致；注释人：Cursor
func RecvTopic(terminalNo string, msgID MsgId) string {
	return fmt.Sprintf("recv/%s/%04X", terminalNo, uint16(msgID))
}

// PublishRecvDecoded 将解码后的结构体序列化为 JSON 发布；原因：业务侧订阅统一格式；注释人：Cursor
func PublishRecvDecoded(terminalNo string, msgID MsgId, decoder JT808Decoder) {
	m := globalMQTTBroker.Load()
	if m == nil || m.server == nil {
		return
	}
	data, err := json.Marshal(decoder)
	if err != nil {
		slog.Error("MQTT 序列化解码结果失败", slog.Any("err", err))
		return
	}
	env := struct {
		TerminalNo string          `json:"terminalNo"`
		MsgID      uint16          `json:"msgId"`
		MsgIDHex   string          `json:"msgIdHex"`
		MsgName    string          `json:"msgName,omitempty"`
		Data       json.RawMessage `json:"data"`
	}{
		TerminalNo: terminalNo,
		MsgID:      uint16(msgID),
		MsgIDHex:   fmt.Sprintf("%04X", uint16(msgID)),
		MsgName:    msgID.String(),
		Data:       data,
	}
	out, err := json.Marshal(env)
	if err != nil {
		slog.Error("MQTT 封装 envelope 失败", slog.Any("err", err))
		return
	}
	topic := RecvTopic(terminalNo, msgID)
	if err := m.server.Publish(topic, out, false, 1); err != nil {
		slog.Error("MQTT Publish 失败", slog.String("topic", topic), slog.Any("err", err))
	}
}

// mqttDownEnvelope 与上行 recv 一致：data 内为对应 MsgId 的结构体 JSON；原因：下行约定对齐；注释人：Cursor
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

// applyEncoderTerminalAndMsgID 用主题中的终端号与消息 ID 覆盖结构体，避免 JSON 漏填；原因：会话以 topic 终端为准；注释人：Cursor
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
}

// HandleMQTTSendTopic 处理 send/{terminalNo}/{msgIdHex}：JSON→对应 Struct→注册 Encoder→发完整 JT808 帧；原因：平台只发业务 JSON；注释人：Cursor
func HandleMQTTSendTopic(topic string, payload []byte) {
	parts := strings.Split(strings.Trim(topic, "/"), "/")
	if len(parts) != 3 || parts[0] != "send" {
		slog.Warn("send 主题格式无效，期望 send/{terminalNo}/{msgId}", slog.String("topic", topic))
		return
	}
	terminalNo := parts[1]
	msgID, err := parseTopicMsgIDHex(parts[2])
	if err != nil {
		slog.Warn("send 主题中 msgId 无法解析为十六进制", slog.String("topic", topic), slog.Any("err", err))
		return
	}

	v, ok := sessionMap.Load(terminalNo)
	if !ok {
		slog.Warn("MQTT 下行无 JT808 会话", slog.String("terminalNo", terminalNo), slog.String("topic", topic))
		return
	}
	session := v.(*Session)
	if session.Conn == nil {
		return
	}

	encoder := GetEncoder(msgID)
	if encoder == nil {
		slog.Warn("未注册该消息 ID 的编码器", slog.String("terminalNo", terminalNo), slog.Uint64("msgId", uint64(msgID)))
		return
	}

	var env mqttDownEnvelope
	_ = json.Unmarshal(payload, &env)

	jsonBody := payload
	if len(env.Data) > 0 && string(env.Data) != "null" {
		jsonBody = env.Data
	}

	if err := json.Unmarshal(jsonBody, encoder); err != nil {
		slog.Error("MQTT 下行 JSON 反序列化到编码器失败",
			slog.String("terminalNo", terminalNo),
			slog.Uint64("msgId", uint64(msgID)),
			slog.Any("err", err),
		)
		return
	}

	if env.MsgID != 0 && MsgId(env.MsgID) != msgID {
		slog.Warn("MQTT 下行 JSON msgId 与主题不一致，以主题为准",
			slog.Uint64("topicMsgId", uint64(msgID)),
			slog.Uint64("jsonMsgId", uint64(env.MsgID)),
		)
	}

	applyEncoderTerminalAndMsgID(encoder, terminalNo, msgID)

	frame := encoder.Encode()
	if len(frame) == 0 {
		slog.Warn("Encoder 返回空帧", slog.String("terminalNo", terminalNo), slog.Uint64("msgId", uint64(msgID)))
	}

	if _, err := session.Conn.Write(frame); err != nil {
		slog.Error("MQTT 下行写入终端连接失败", slog.String("terminalNo", terminalNo), slog.Any("err", err))
	}
}
