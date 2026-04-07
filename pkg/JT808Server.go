package pkg

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"net"
)

// JT808Server JT/T808 协议 TCP 服务；注释原因：实现 GatewayServer 接口，支持配置化多实例启动；注释人：Cursor
type JT808Server struct {
	cfg ServerCfg
}

// NewJT808Server 创建 JT808 服务实例；注释原因：由 BuildServers 根据配置调用；注释人：Cursor
func NewJT808Server(cfg ServerCfg) *JT808Server {
	return &JT808Server{cfg: cfg}
}

// Name 实现 GatewayServer 接口；注释原因：用于日志标识和多服务区分；注释人：Cursor
func (s *JT808Server) Name() string { return "JT808" }

// Start 启动 JT808 TCP 监听（阻塞）；注释原因：实现 GatewayServer.Start，调用方需在 goroutine 中运行；注释人：Cursor
func (s *JT808Server) Start() error {
	addr := s.cfg.Addr
	if addr == "" {
		addr = ":1808"
	}
	bufSize := s.cfg.ReadBuffer
	if bufSize <= 0 {
		bufSize = 4096
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("[%s] 监听 %s 失败: %w", s.Name(), addr, err)
	}
	defer ln.Close()
	log.Printf("[%s] 服务启动，监听 %s，读缓冲 %d 字节", s.Name(), addr, bufSize)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[%s] accept error: %v", s.Name(), err)
			continue
		}
		slog.Info("收到一个连接",
			slog.String("server", s.Name()),
			slog.String("remote", conn.RemoteAddr().String()),
		)
		go parseFrame(conn, bufSize)
	}
}

// StartJT808 兼容旧版调用（使用默认配置）；注释原因：保留给不使用配置文件的场景；注释人：Cursor
func StartJT808() {
	srv := NewJT808Server(ServerCfg{Addr: ":1808", ReadBuffer: 4096, Enabled: true})
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}

// parseFrame 处理单个 TCP 连接的完整上行生命周期；注释原因：每条连接独立 goroutine，bufSize 来自配置；注释人：Cursor
func parseFrame(conn net.Conn, bufSize int) {
	lastTerminal := ""
	// sessionRegistered 避免每条消息重复写 sync.Map；注释原因：减少无效写操作；注释人：Cursor
	sessionRegistered := false
	defer func() {
		UnregisterSessionIfConn(lastTerminal, conn)
		conn.Close()
	}()
	reader := bufio.NewReaderSize(conn, bufSize)
	for {
		frame, err := reader.ReadBytes(0x7e)
		if err != nil {
			slog.Info("连接断开", slog.String("terminalNo", lastTerminal), slog.Any("err", err))
			return
		}
		// 单独的 0x7E 是上一帧的结束符，与下一帧头合并；注释原因：JT808 帧以 0x7E 为界符；注释人：Cursor
		if len(frame) == 1 {
			next, err := reader.ReadBytes(0x7e)
			if err != nil {
				slog.Info("连接断开", slog.String("terminalNo", lastTerminal), slog.Any("err", err))
				return
			}
			frame = append(frame, next...)
		}
		// 转义还原
		frame, err = Unescape(frame)
		if err != nil {
			slog.Error("帧转义解码失败", slog.Any("err", err))
			continue
		}
		// JT808最小帧：0x7E + 12字节头 + 1字节BCC + 0x7E = 15字节；注释原因：防止越界；注释人：Cursor
		if len(frame) < 15 {
			slog.Error("帧长度过短", slog.Int("len", len(frame)))
			continue
		}
		// 校验：MSGID~BODY做XOR，校验失败直接丢弃；注释原因：避免解析非法帧；注释人：Cursor
		if XOR(frame[1:len(frame)-2]) != frame[len(frame)-2] {
			slog.Error("帧校验失败，丢弃")
			continue
		}
		// [上行-1] 帧解析
		msg := ParseFrame(frame)
		msgIdHex := fmt.Sprintf("%04X", uint16(msg.MsgID))
		// terminalNo 变化时重新注册（例如同连接切换终端，实际少见）；注释原因：保持会话一致；注释人：Cursor
		if msg.TerminalNo != lastTerminal {
			lastTerminal = msg.TerminalNo
			sessionRegistered = false
		}
		// Debug 级别输出完整帧十六进制，避免正常运行时的 CPU 开销；注释人：Cursor
		if slog.Default().Handler().Enabled(context.Background(), slog.LevelDebug) {
			slog.Debug("[上行-1] 收到终端帧",
				slog.String("terminalNo", lastTerminal),
				slog.String("msgId", msgIdHex),
				slog.String("msgName", msg.MsgID.String()),
				slog.Int("frameLen", len(frame)),
				slog.Int("bodyLen", len(msg.Body)),
				slog.String("frameHex", hex.EncodeToString(frame)),
			)
		} else {
			slog.Info("[上行-1] 收到终端帧",
				slog.String("terminalNo", lastTerminal),
				slog.String("msgId", msgIdHex),
				slog.String("msgName", msg.MsgID.String()),
				slog.Int("frameLen", len(frame)),
				slog.Int("bodyLen", len(msg.Body)),
			)
		}

		// [上行-2] 查找解码器
		decoder := GetDecoder(msg.MsgID)
		if decoder == nil {
			slog.Warn("[上行-2] 未找到解码器，忽略该消息",
				slog.String("terminalNo", lastTerminal),
				slog.String("msgId", msgIdHex),
			)
			continue
		}

		// [上行-3] 解码消息体
		if err := decoder.Parse(&msg); err != nil {
			slog.Error("[上行-3] 消息体解码失败",
				slog.String("terminalNo", lastTerminal),
				slog.String("msgId", msgIdHex),
				slog.Any("err", err),
			)
			continue
		}
		slog.Info("[上行-3] 消息体解码成功",
			slog.String("terminalNo", lastTerminal),
			slog.String("msgId", msgIdHex),
			slog.String("msgName", msg.MsgID.String()),
		)

		// [上行-4] 注册/更新会话（首次或终端号变化）
		if !sessionRegistered {
			RegisterSession(&Session{Conn: conn, IMEI: msg.TerminalNo})
			sessionRegistered = true
			slog.Info("[上行-4] 注册终端会话",
				slog.String("terminalNo", lastTerminal),
				slog.String("remote", conn.RemoteAddr().String()),
			)
		}

		// [上行-5] 通过桥接器发布上行事件（在 NotifyUplink/PublishUplink 内部记录）
		NotifyUplink(msg.TerminalNo, msg.MsgID, decoder)

		// [上行-6] 业务回调（各 T*.go OnMsg 内部有具体日志）
		decoder.OnMsg(conn)
	}
}
