package pkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	ESC   = 0x7D
	FRAME = 0x7E
)

func Escape(input []byte) []byte {
	if len(input) == 0 {
		return input
	}
	out := bytes.NewBuffer(make([]byte, 0, len(input)*2))
	for _, b := range input {
		switch b {
		case ESC:
			out.WriteByte(ESC)
			out.WriteByte(0x01)
		case FRAME:
			out.WriteByte(ESC)
			out.WriteByte(0x02)
		default:
			out.WriteByte(b)
		}
	}

	return out.Bytes()
}

func Unescape(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return input, nil
	}
	out := bytes.NewBuffer(make([]byte, 0, len(input)))
	for i := 0; i < len(input); i++ {
		b := input[i]
		if b == ESC {
			if i+1 >= len(input) {
				return nil, fmt.Errorf("truncated escape sequence")
			}
			i++
			switch input[i] {
			case 0x01:
				out.WriteByte(ESC)
			case 0x02:
				out.WriteByte(FRAME)
			default:
				return nil, fmt.Errorf("invalid escape code 0x%02X", input[i])
			}
		} else {
			out.WriteByte(b)
		}
	}
	return out.Bytes(), nil
}

// 校验（从第2个字节到倒数第2个字节异或）
func Check(frame []byte) bool {
	//if len(frame) < 3 {
	//	return false
	//}
	//sum := byte(0)
	//for _, b := range frame[1 : len(frame)-2] {
	//	sum ^= b
	//}
	sum := CRC(frame)
	return sum == frame[len(frame)-2]
}

func CRC(frame []byte) byte {
	sum := byte(0)
	for _, b := range frame[1 : len(frame)-2] {
		sum ^= b
	}
	return sum
}

// BCDToString 把BCD码转为字符串
func BCDToString(b []byte) string {
	res := make([]byte, 0, len(b)*2)
	for _, v := range b {
		hi := v >> 4
		lo := v & 0x0F
		res = append(res, '0'+hi, '0'+lo)
	}
	// 去掉可能的前导0
	return string(res)
}

func ParseFrame(frame []byte) JTMessage {
	msg := JTMessage{
		MsgID:      MsgId(binary.BigEndian.Uint16(frame[1:3])), //后面字节是[1)
		Prop:       binary.BigEndian.Uint16(frame[3:5]),        //后面字节是[3)
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
		msg.Body = frame[13 : len(frame)-2]
	}
	return msg
}

func PackFrame(msg JTMessage) []byte {

	bodyLen := len(msg.Body)

	attr := uint16(bodyLen) // 不分包 不加密

	buf := make([]byte, 0, 1024)

	header := make([]byte, 12)

	binary.BigEndian.PutUint16(header[0:], uint16(msg.MsgID))
	binary.BigEndian.PutUint16(header[2:], attr)

	copy(header[4:], msg.TerminalNo)

	binary.BigEndian.PutUint16(header[10:], msg.SeqNo)

	buf = append(buf, header...)
	buf = append(buf, msg.Body...)

	cs := CRC(buf)

	buf = append(buf, cs)

	buf = Escape(buf)

	frame := make([]byte, 0, len(buf)+2)

	frame = append(frame, 0x7E)
	frame = append(frame, buf...)
	frame = append(frame, 0x7E)

	return frame
}
