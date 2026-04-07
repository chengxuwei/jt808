package pkg

import (
	"fmt"
	"log"
	"log/slog"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

type MQTTServer struct {
	server *mqtt.Server
}

// Publish 将消息发布到内嵌 MQTT Broker；注释原因：供 MQTTBridge.PublishUplink 调用；注释人：Cursor
func (m *MQTTServer) Publish(topic string, payload []byte) error {
	if m == nil || m.server == nil {
		return fmt.Errorf("MQTTServer: server 未初始化")
	}
	return m.server.Publish(topic, payload, false, 1)
}

// PublishMsg 字符串发布（兼容保留）；注释原因：外部业务代码可能仍使用此方法；注释人：Cursor
func (m *MQTTServer) PublishMsg(topic string, msg string, qos byte) {
	if m == nil || m.server == nil {
		return
	}
	_ = m.server.Publish(topic, []byte(msg), false, qos)
}

// StartMqttBroker 启动内嵌 MQTT Broker（阻塞）；注释原因：addr 由 MQTTBridge 从配置注入，不再硬编码；注释人：Cursor
func (m *MQTTServer) StartMqttBroker(addr string) {
	if addr == "" {
		addr = ":1883"
	}
	options := &mqtt.Options{
		InlineClient: true,
	}
	server := mqtt.New(options)
	m.server = server

	// 创建基于用户名密码的认证
	//server.AddHook(new(auth.Hook), &auth.Options{
	//	Ledger: &auth.Ledger{
	//		Auth: auth.AuthRules{ // 认证规则
	//			{Username: "user1", Password: "pass1", Allow: true},
	//			{Username: "user2", Password: "pass2", Allow: true},
	//		},
	//		ACL: auth.ACLRules{ // 访问控制规则
	//			{Username: "user1", Filters: auth.Filters{
	//				"#": auth.ReadWrite, // user1可以读写所有主题
	//			}},
	//			{Username: "user2", Filters: auth.Filters{
	//				"public/#": auth.ReadOnly, // user2只能读取public/前缀的主题
	//			}},
	//		},
	//	},
	//})
	// For security reasons, the default implementation disallows all connections.
	// If you want to allow all connections, you must specifically allow it.
	err := server.AddHook(new(auth.AllowHook), nil)
	if err != nil {
		log.Fatal(err)
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "t1",
		Address: addr,
	})
	err = server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	// 订阅 send/{terminalNo}/{msgIdHex}，msgId 为十六进制，如 send/013912345678/8300
	// 同时订阅 /send/+/+ 兼容客户端携带前导斜杠的发布；注释原因：MQTT 协议中 /send/x/y 与 send/x/y 是不同主题；注释人：Cursor
	sendHandler := func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
		slog.Info("[下行-1] MQTT Broker 收到下行指令",
			slog.String("clientId", cl.ID),
			slog.String("topic", pk.TopicName),
			slog.Int("payloadLen", len(pk.Payload)),
		)
		HandleMQTTSendTopic(pk.TopicName, pk.Payload)
	}
	if err := server.Subscribe("send/+/+", 1, sendHandler); err != nil {
		log.Fatal(err)
	}
	// 兼容前导斜杠 /send/+/+；注释原因：部分MQTT客户端默认会加前导斜杠；注释人：Cursor
	if err := server.Subscribe("/send/+/+", 1, sendHandler); err != nil {
		log.Fatal(err)
	}

	slog.Info("MQTT Broker 启动",
		slog.String("addr", addr),
		slog.String("topics", "send/+/+, /send/+/+"),
	)
	// 5. 启动服务（阻塞）
	err1 := server.Serve()
	if err1 != nil {
		log.Fatal(err1)
	}
	// 3. 发布示例
	//ticker := time.NewTicker(5 * time.Second)
	//defer ticker.Stop()
	//
	//for {
	//	select {
	//	case <-ticker.C:
	//		//todo 定义全局变量从JT808收到消息转发MQTT
	//		if errr := server.Publish("topic/test", []byte("{\"a\": \"b\"}"), true, 1); errr != nil {
	//			log.Fatal(errr)
	//		}
	//	}
	//}

}
