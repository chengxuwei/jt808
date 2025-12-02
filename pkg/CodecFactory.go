package pkg

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
)

var (
	//消息ID定义
	P0100 = uint16(0x0100)
	P0200 = uint16(0x0200)
	P8300 = uint16(0x08300)
	P6006 = uint16(0x06006)
	P0102 = uint16(0x0102)
	//解码Map
	decodeFuncMap = make(map[uint16]JT808Codec)
	//处理
	//会话Map
	sessionMap sync.Map
)

type (
	JT808Codec interface {
		Parse(jtMsg *JTMessage) error // 解析终端上传的body数据
		Encode() []byte
	}
	Session struct {
		Conn net.Conn
		IMEI string
	}
)

/*
		*
	  注册Handler
*/
func RegisterDecoder(msgId uint16, handler JT808Codec) {
	slog.Info("注册一个解码器", slog.Any("msgId", msgId))
	decodeFuncMap[msgId] = handler
}

/*
*
 */
func SendMsgToDevice(imei string, msg []byte) {
	storeObj, ok := sessionMap.Load(imei)
	if ok {
		session := storeObj.(*Session)
		session.Conn.Write(msg)
	}
}

func saveSession(imei string, conn net.Conn) {
	sessionMap.Store(imei, &Session{
		Conn: conn,
		IMEI: imei,
	})
}
func idelSession(imei string, conn net.Conn) {

}

/*
*
获取Handler
*/
func GetDecoder(msgId uint16) JT808Codec {
	return decodeFuncMap[msgId]
}

/*
*
*
编码消息帧
*/
func EncodeFrame(message *JTMessage) ([]byte, error) {
	return nil, nil
}

/*
*
w
  - 处理消息
    conn 注册Session,或发布
*/
func OnMsg(decoder JT808Codec, conn net.Conn) {
	switch decoder.(type) {
	case *T0200:
		OnT0200(decoder.(*T0200))
		fmt.Println("T0200分发处理")
	case *T0100:
		OnT0100(decoder.(*T0100))
		fmt.Println("T0100分发处理")
	}
}

func OnT0100(t0100 *T0100) {
	//TODO 1. 推送消息到三方 2.注册Session，保存会话
}

func OnT0200(t0200 *T0200) {

}
