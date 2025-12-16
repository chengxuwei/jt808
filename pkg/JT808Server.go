package pkg

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"log"
	"log/slog"
	"net"
)

func StartJT808() {
	ln, err := net.Listen("tcp", ":1808")
	log.Println("JT808 服务器临听 :1808")
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
	defer conn.Close()
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
		//校验
		if !Check(frame) {
			slog.Error("帧校验失败")
		}
		slog.Info("转义解码后的帧", slog.Any("frame", hex.EncodeToString(frame)))
		//解析头,默认不分包
		msg := JTMessage{
			MsgID:      binary.BigEndian.Uint16(frame[1:3]), //后面字节是[1)
			Prop:       binary.BigEndian.Uint16(frame[3:5]), //后面字节是[3)
			TerminalNo: BCDToString(frame[5:11]),
			SeqNo:      binary.BigEndian.Uint16(frame[11:13]), //后面字节是[3),
			PkgCount:   0,
			PkgNo:      0,
		}
		//是否分包
		if (msg.Prop >> 13 & 1) == 1 {
			msg.PkgCount = binary.BigEndian.Uint16(frame[13:15])
			msg.PkgNo = binary.BigEndian.Uint16(frame[15:17])
			msg.Body = frame[17:]
		} else {
			msg.Body = frame[13:]
		}
		slog.Info("收到一个数据包", slog.Any("msg", hex.EncodeToString(msg.Body)))
		decoder := GetDecoder(msg.MsgID)
		if decoder != nil {
			//解析字段
			decoder.Parse(&msg)
			//分派处理
			OnMsg(decoder, conn)
		} else {
			slog.Info("没找到", slog.Any("msg", msg))
		}

	}
}
