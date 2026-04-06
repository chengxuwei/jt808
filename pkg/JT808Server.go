package pkg

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"net"
)

func StartJT808() {
	ln, err := net.Listen("tcp", ":1808")
	log.Println("JT808 服务器监听 :1808")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	//handlerMap := make(map[uint16]handler.JT808Handler)
	////handlerMap[0x0001] =
	////	log.Println("Server listening on :1808")
	//handlerMap[0x0200] = &handler.JT0200Handler{}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		} else {
			slog.Info("收到一个连接")
		}
		go parseFrame(conn)
	}
}

func parseFrame(conn net.Conn) {
	lastTerminal := ""
	defer func() {
		UnregisterSessionIfConn(lastTerminal, conn)
		conn.Close()
	}()
	reader := bufio.NewReader(conn)
	for {
		frame, err := reader.ReadBytes(0x7e)
		if err != nil {
			//TODO 连接中断
			return
		}
		if len(frame) == 1 {
			next, err := reader.ReadBytes(0x7e)
			if err != nil {
				//TODO 连接中断
				return
			}
			frame = append(frame, next[0:len(next)]...)
		}
		//转义
		frame, _ = Unescape(frame)
		//校验:MSGID~BODY; BCC,7E 不算
		if XOR(frame[1:len(frame)-2]) != frame[len(frame)-2] {
			slog.Error("帧校验失败")
		}
		//解析头,默认不分包
		msg := ParseFrame(frame)
		lastTerminal = msg.TerminalNo

		slog.Info("转义解码后的帧", slog.String("terminalNo", lastTerminal), slog.String("msgId", msg.MsgID.String()), slog.Any("frame", hex.EncodeToString(frame)))

		//slog.Info("收到BODY", slog.String("终端号：", msg.TerminalNo), slog.Any("消息ID", msg.MsgID), slog.Any("msg", hex.EncodeToString(msg.Body)))
		decoder := GetDecoder(msg.MsgID)
		if decoder != nil {
			// INSERT_YOUR_CODE
			slog.Info("终端号和解码器信息", slog.String("终端号", msg.TerminalNo), slog.String("解码器", fmt.Sprintf("%T", decoder)))
			// slog.Info("调用解码器", slog.String("hex:", fmt.Sprintf("0x%04X", uint16(msg.MsgID))), slog.String("msgId", msg.MsgID.String())	)
			if err := decoder.Parse(&msg); err != nil {
				slog.Error("解码失败", slog.Any("err", err))
				continue
			}
			//会话处理
			RegisterSession(&Session{Conn: conn, IMEI: msg.TerminalNo})
			//发送
			PublishRecvDecoded(msg.TerminalNo, msg.MsgID, decoder)
			//业务处理
			decoder.OnMsg(conn)
		} else {
			slog.Info("没找到", slog.Any("msg", msg))
		}

	}
}
