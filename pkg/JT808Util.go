package pkg

import (
	"bytes"
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
	if len(frame) < 3 {
		return false
	}
	sum := byte(0)
	for _, b := range frame[1 : len(frame)-2] {
		sum ^= b
	}
	return sum == frame[len(frame)-2]
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
