package pkg

import (
	"encoding/binary"
	"log/slog"
)

/**
*
	位置上报
*/

type T0200 struct {
	JTMessage
	Alarm  uint32  `json:"Alarm"`
	Status uint32  `json:"Status"`
	Lat    float32 `json:"Lat"`
	Lon    float32 `json:"Lon"`
}

func init() {
	RegisterDecoder(P0200, &T0200{})
}

func (h *T0200) Parse(jtMsg *JTMessage) error {
	// 基础信息
	h.JTMessage = *jtMsg
	// 假设 body 已经在 jtMsg.Body
	body := jtMsg.Body
	h.Alarm = binary.BigEndian.Uint32(body[0:5])
	h.Status = binary.BigEndian.Uint32(body[5:9])
	h.Lat = float32(binary.BigEndian.Uint32(body[9:13]) / 1000000)
	h.Lon = float32(binary.BigEndian.Uint32(body[13:17]) / 1000000)
	slog.Info("位置上报:",
		slog.Any("告警", h.Alarm),
		slog.Any("状态", h.Status),
		slog.Any("纬度", h.Lat),
		slog.Any("经度", h.Lon),
	)
	return nil
}

func (h *T0200) Encode() []byte {
	//发送时先编码信息
	return []byte{0x02, 0x00}
}
func (h *T0200) Name() string {
	return h.MsgID.String()
}
