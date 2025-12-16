package pkg

import (
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"log"
	"time"
)

func StartMqttBroker() {
	options := &mqtt.Options{
		InlineClient: true,
	}

	server := mqtt.New(options)
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
		Address: ":1883",
	})
	err = server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	// 4. 订阅示例
	callbackFn := func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
		server.Log.Info("收到一个MQTT消息，转发到JT808服务器", "client", cl.ID, "subscriptionId", sub.Identifier, "topic", pk.TopicName, "payload", string(pk.Payload))
	}
	server.Subscribe("direct/#", 1, callbackFn)

	// 5. 启动服务
	err1 := server.Serve()
	if err1 != nil {
		log.Fatal(err)
	}
	// 3. 发布示例
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			//todo 定义全局变量从JT808收到消息转发MQTT
			if errr := server.Publish("topic/test", []byte("test"), true, 1); errr != nil {
				log.Fatal(errr)
			}
		}
	}

}
