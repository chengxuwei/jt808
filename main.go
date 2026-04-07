package main

import (
	"jt808/pkg"
	"log"
)

func main() {
	// 1. 加载配置（优先读 ./config.yaml，缺失时使用默认值）
	cfg, err := pkg.LoadConfig()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 2. 初始化日志（按 config.yaml log.level 设置，需在所有服务启动前调用）
	pkg.InitLogger(cfg.Log)

	// 3. 启动消息桥接器（MQTT Broker）；注释原因：替换为 KafkaBridge 时只改此处；注释人：Cursor
	if cfg.MQTT.Enabled {
		bridge := pkg.NewMQTTBridge(cfg.MQTT)
		pkg.SetBridge(bridge)
		if err := bridge.Start(); err != nil {
			log.Fatal("MQTT 桥接器启动失败:", err)
		}
	}

	// 4. 启动协议网关服务（根据 config.yaml servers.* 按需启动）
	servers := pkg.BuildServers(cfg)
	if len(servers) == 0 {
		log.Fatal("未启用任何协议服务，请检查 config.yaml servers 配置")
	}
	for _, srv := range servers {
		s := srv
		go func() {
			if err := s.Start(); err != nil {
				log.Printf("[%s] 服务退出: %v", s.Name(), err)
			}
		}()
	}

	// 5. 阻塞主 goroutine
	select {}
}
