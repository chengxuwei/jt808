package pkg

// GatewayServer 协议网关服务抽象接口；
// 注释原因：支持 JT808/JT1078/GB32960 等多协议在不同端口同时运行，新增协议只需实现此接口并在 BuildServers 注册；注释人：Cursor
//
// 使用方式：
//
//	type MyServer struct { cfg ServerCfg }
//	func (s *MyServer) Name() string    { return "MyProtocol" }
//	func (s *MyServer) Start() error    { /* 监听+处理循环 */ }
type GatewayServer interface {
	// Name 返回服务名称，用于日志标识和配置 Key
	Name() string
	// Start 启动服务（阻塞），调用方应在 goroutine 中运行；返回 error 表示启动失败
	Start() error
}
