package pkg

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
)

/*
*

	终端注册
*/
type T0100 struct {
	JTMessage
	ProvinceId uint16 `json:"province_id"` //省份ID
	CityId     uint16 `json:"city_id"`     //城市ID

}

func init() {
	//注册 TODO ，保证协程安全，可在连接时动态创建
	//TODO 增加context上下文处理
	codec := &T0100{}
	codec.MsgID = P0100
	RegisterDecode(codec)
	RegisterEncode(codec)
}
func (h *T0100) Parse(msg *JTMessage) error {
	h.JTMessage = *msg
	// 假设 body 已经在 jtMsg.Body
	slog.Info("解析数据，然后Process处理", slog.Any("msg", msg))
	//len(msg.Body)
	// T0100注册消息body不能为空；注释原因：注册报文必须携带省市等注册信息；注释人：Cursor
	if len(msg.Body) == 0 {
		return fmt.Errorf("T0100 body为空，注册消息body长度必须大于0")
	}

	return nil
}

func (h *T0100) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0100) GetMsgId() MsgId {
	return P0100
}
func (h *T0100) OnMsg(conn net.Conn) {
	logJT808DecodedJSON("收到终端注册 0x0100（JSON）", h)
	//业务回复；注释原因：SeqNo 由 NextSeqNo 分配，与终端上行序号独立；注释人：Cursor
	t8100 := T8100{
		JTMessage: JTMessage{
			MsgID:      P8100,
			TerminalNo: h.TerminalNo,
			SeqNo:      NextSeqNo(h.TerminalNo), // 平台侧独立递增序号
		},
		ResponseSeqNo: h.SeqNo, // 回复终端上行序号
		Status:        0,
		Token:         "htzj-1264579279",
	}
	frame := t8100.Encode()
	slog.Info("T0100 回复 T8100",
		slog.String("terminalNo", h.TerminalNo),
		slog.String("recvMsgId", h.MsgID.String()),
		slog.Uint64("seqNo", uint64(t8100.SeqNo)),
		slog.String("frameHex", hex.EncodeToString(frame)),
	)
	if _, err := conn.Write(frame); err != nil {
		slog.Error("T0100 回复 T8100 失败", slog.String("terminalNo", h.TerminalNo), slog.Any("err", err))
	}

}
