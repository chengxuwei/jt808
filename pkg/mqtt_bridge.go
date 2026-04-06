package pkg

import (
	"encoding/json"
	"fmt"
	"log/slog"
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

// HandleMQTTSendTopic 处理 send/{terminalNo}/{msgId} 下行：按终端查 Session 并写入 payload；原因：平台经 MQTT 下发 JT808 原始帧；注释人：Cursor
func HandleMQTTSendTopic(topic string, payload []byte) {
	parts := strings.Split(strings.Trim(topic, "/"), "/")
	if len(parts) != 3 || parts[0] != "send" {
		slog.Warn("send 主题格式无效，期望 send/{terminalNo}/{msgId}", slog.String("topic", topic))
		return
	}
	terminalNo := parts[1]
	v, ok := sessionMap.Load(terminalNo)
	if !ok {
		slog.Warn("MQTT 下行无 JT808 会话", slog.String("terminalNo", terminalNo), slog.String("topic", topic))
		return
	}
	session := v.(*Session)
	if session.Conn == nil {
		return
	}
	if _, err := session.Conn.Write(payload); err != nil {
		slog.Error("MQTT 下行写入终端连接失败", slog.String("terminalNo", terminalNo), slog.Any("err", err))
	}
}
