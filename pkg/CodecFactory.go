package pkg

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
)

type MsgId uint16

var (
	//消息ID定义
	P0001 = MsgId(0x0001)
	P0002 = MsgId(0x0002)
	P8001 = MsgId(0x08001)
	P0100 = MsgId(0x0100)
	P8100 = MsgId(0x8100)
	P0003 = MsgId(0x0003)
	P0102 = MsgId(0x0102)

	P0200 = MsgId(0x0200)
	P0704 = MsgId(0x0704)
	P8103 = MsgId(0x8103)
	P8105 = MsgId(0x8105)
	P8104 = MsgId(0x8104)
	P0104 = MsgId(0x0104)
	P8201 = MsgId(0x8201)
	P0201 = MsgId(0x0201)
	P8300 = MsgId(0x08300)
	P6006 = MsgId(0x06006)

	//解码Map
	decodeFuncMap = make(map[MsgId]JT808Codec)
	//处理
	//会话Map
	sessionMap sync.Map
)

func (msgId MsgId) String() string {

	switch msgId {
	case P0001:
		return "终端-通用应答"
	case P0002:
		return "终端-心跳"
	case P8001:
		return "平台-通用应答"
	case P0100:
		return "终端-注册"
	case P8100:
		return "平台-终端注册应答"
	case P0003:
		return "终端-注销"
	case P0102:
		return "终端-鉴权"
	case P0200:
		return "终端-位置信息上报"
	case P0704:
		return "终端-位置信息批量上传"
	case P8103:
		return "平台-设置终端参数"
	case P8105:
		return "平台-终端控制"
	case P8104:
		return "平台-查询终端参数"
	case P0104:
		return "终端-查询终端参数应答"
	case P8201:
		return "平台-位置信息查询"
	case P0201:
		return "终端-位置信息查询应答"

	case P8300:
		return "终端-文本信息下发"
	case P6006:
		return "终端-批量位置信息上报"
	}

	return "平台-暂未实现的命令"
}

type (
	JT808Codec interface {
		Parse(jtMsg *JTMessage) error // 解析终端上传的body数据
		Encode() []byte
		Name() string //返回名称
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
func RegisterDecoder(msgId MsgId, handler JT808Codec) {
	slog.Info("注册一个解码器", slog.Any("msgId", fmt.Sprintf("%04x", msgId)))
	decodeFuncMap[msgId] = handler
}

/*
* 发布消息
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
func GetDecoder(msgId MsgId) JT808Codec {
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
