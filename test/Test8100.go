package main

import (
	"encoding/hex"
	"jt808/pkg"
	"log/slog"
	"strings"
)

func main() {
	t8100 := pkg.T8100{
		JTMessage: pkg.JTMessage{
			MsgID:      pkg.P8100,
			TerminalNo: "061000210052",
			SeqNo:      44,
		},
		ResponseSeqNo: 26,
		Status:        0,
		Token:         "htzj-1264579279",
	}
	frame := t8100.Encode()
	slog.Info("转义编码发送帧", slog.String("terminalNo", t8100.TerminalNo), slog.String("msgId", t8100.MsgID.String()), slog.Any("frame", strings.ToUpper(hex.EncodeToString(frame)))) //pkg.StartJT808()
}
