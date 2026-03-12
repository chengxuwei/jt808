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
func StringToBCD(s string) []byte {

	// 如果是奇数长度，前补0
	if len(s)%2 == 1 {
		s = "0" + s
	}

	b := make([]byte, len(s)/2)

	for i := 0; i < len(s); i += 2 {

		high := s[i] - '0'
		low := s[i+1] - '0'

		b[i/2] = (high << 4) | low
	}

	return b
}

// 校验（从第2个字节到倒数第2个字节异或）
//func Check(frame []byte) bool {
//	sum := CRC(frame)
//	return sum == frame[len(frame)-2]
//}

func XOR(data []byte) byte {
	var cs byte
	for _, b := range data {
		cs ^= b
	}
	return cs
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

	copy(header[4:], StringToBCD(msg.TerminalNo))

	binary.BigEndian.PutUint16(header[10:], msg.SeqNo)

	buf = append(buf, header...)
	buf = append(buf, msg.Body...)

	cs := XOR(buf)

	buf = append(buf, cs)

	buf = Escape(buf)

	frame := make([]byte, 0, len(buf)+2)

	frame = append(frame, 0x7E)
	frame = append(frame, buf...)
	frame = append(frame, 0x7E)

	return frame
}

func Get8001Buf(message JTMessage, result uint8) []byte {
	//包回昨
	t8001 := T8001{
		JTMessage: JTMessage{
			MsgID:      P8001,
			TerminalNo: message.TerminalNo,
		},
		//回复序号
		ResponseSeqNo: message.SeqNo,
		//回复消息ID
		ResponseMsgID: uint16(message.MsgID),
		//结果
		Result: result,
	}
	frame1 := t8001.Encode()
	return frame1
}
