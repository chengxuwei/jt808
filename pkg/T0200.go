package pkg

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
)

// T0200BaseLen 标准消息体最小长度：报警+状态+经纬度+高程+速度+方向+时间(BCD6)；原因：JT/T 808 位置汇报基本数据单元；注释人：Cursor
const T0200BaseLen = 28

// T0200Extra 附加信息项（终端 ID + 长度 + 内容）；原因：JT808 表 23/24 TLV 扩展；注释人：Cursor
type T0200Extra struct {
	ID    uint8  `json:"id"`
	Len   uint8  `json:"len"`
	Value []byte `json:"value"`
}

// T0200OverspeedExtra 附加信息 0x11 超速报警；原因：表 25 固定 5 字节；注释人：Cursor
type T0200OverspeedExtra struct {
	LocationType  uint8  `json:"locationType"`  // 位置类型
	AreaOrRouteID uint32 `json:"areaOrRouteId"` // 区域或路段 ID
}

// T0200StatusBits 状态 DWORD 常用位释义（见 JT808 状态位定义）；原因：便于 MQTT/业务直接使用；注释人：Cursor
type T0200StatusBits struct {
	ACCOn         bool `json:"accOn"`         // bit0 0:ACC 关 1:ACC 开
	Located       bool `json:"located"`       // bit1 0:未定位 1:定位
	SouthLatitude bool `json:"southLatitude"` // bit2 0:北纬 1:南纬
	WestLongitude bool `json:"westLongitude"` // bit3 0:东经 1:西经
	Operating     bool `json:"operating"`     // bit4 运营状态
	Encrypted     bool `json:"encrypted"`     // bit5 经纬度加密（部分版本）
	// bit6-7 保留
	Overload uint8 `json:"overload"` // bit8-9 客货状态（取值依协议版本）
	// bit10 油路断开 bit11 电路断开 bit12 车门加锁 … 继续展开常用位
	FuelCutoff     bool `json:"fuelCutoff"`     // bit10
	CircuitCutoff  bool `json:"circuitCutoff"`  // bit11
	DoorLocked     bool `json:"doorLocked"`     // bit12
	FrontDoorOpen  bool `json:"frontDoorOpen"`  // bit13
	MidDoorOpen    bool `json:"midDoorOpen"`    // bit14
	RearDoorOpen   bool `json:"rearDoorOpen"`   // bit15
	DriverDoorOpen bool `json:"driverDoorOpen"` // bit16
	CustomDoorOpen bool `json:"customDoorOpen"` // bit17
	GPSUsed        bool `json:"gpsUsed"`        // bit18 0:使用北斗 1:使用 GPS（常见实现，以终端为准）
	BeidouUsed     bool `json:"beidouUsed"`     // bit19
	GLONASSUsed    bool `json:"glonassUsed"`    // bit20
	GalileoUsed    bool `json:"galileoUsed"`    // bit21
}

// T0200 位置信息汇报 0x0200（JT/T 808）
type T0200 struct {
	JTMessage
	Alarm        uint32          `json:"alarm"`
	Status       uint32          `json:"status"`
	StatusBits   T0200StatusBits `json:"statusBits"`
	LatitudeRaw  uint32          `json:"latitudeRaw"`  // 百万分之一度，无符号幅度
	LongitudeRaw uint32          `json:"longitudeRaw"` // 百万分之一度，无符号幅度
	Latitude     float64         `json:"latitude"`     // 度，已按状态位 N/S、E/W 取符号
	Longitude    float64         `json:"longitude"`    // 度
	Altitude     uint16          `json:"altitude"`     // 米
	SpeedRaw     uint16          `json:"speedRaw"`     // 0.1 km/h
	SpeedKmh     float64         `json:"speedKmh"`     // km/h
	Direction    uint16          `json:"direction"`    // 0–359，正北为 0
	DateTime     string          `json:"dateTime"`     // BCD 时间 本地协议时区（一般为北京时间）
	Extras       []T0200Extra    `json:"extras,omitempty"`
	// 常见附加信息解析（0x01 里程、0x02 油量、0x17 信号强度、0x18 卫星数等）
	MileageKm        *float64             `json:"mileageKm,omitempty"`               // 0x01，DWORD，0.1km
	FuelLiters       *float64             `json:"fuelLiters,omitempty"`              // 0x02，WORD，0.1L
	RecordSpeedKmh   *float64             `json:"recordSpeedKmh,omitempty"`          // 0x03，WORD，0.1km/h
	AlarmConfirmNo   *uint16              `json:"alarmConfirmEventId,omitempty"`     // 0x04，WORD
	SignalStrength   *uint8               `json:"signalStrength,omitempty"`          // 0x17，BYTE
	GNSSSatelliteNum *uint8               `json:"gnssSatelliteNum,omitempty"`        // 0x18，BYTE
	ExtendedVehicle  *uint32              `json:"extendedVehicleSignals,omitempty"`  // 0x14，DWORD
	IOStatus         *uint16              `json:"ioStatus,omitempty"`                // 0x15，WORD
	AnalogInput      *uint32              `json:"analogInput,omitempty"`             // 0x16，DWORD
	ExtendedVehicle2 *uint32              `json:"extendedVehicleSignals2,omitempty"` // 0x25，DWORD
	Overspeed        *T0200OverspeedExtra `json:"overspeedExtra,omitempty"`          // 0x11
}

func decodeT0200StatusBits(s uint32) T0200StatusBits {
	return T0200StatusBits{
		ACCOn:          s&1 != 0,
		Located:        s>>1&1 != 0,
		SouthLatitude:  s>>2&1 != 0,
		WestLongitude:  s>>3&1 != 0,
		Operating:      s>>4&1 != 0,
		Encrypted:      s>>5&1 != 0,
		Overload:       uint8(s>>8) & 0x3,
		FuelCutoff:     s>>10&1 != 0,
		CircuitCutoff:  s>>11&1 != 0,
		DoorLocked:     s>>12&1 != 0,
		FrontDoorOpen:  s>>13&1 != 0,
		MidDoorOpen:    s>>14&1 != 0,
		RearDoorOpen:   s>>15&1 != 0,
		DriverDoorOpen: s>>16&1 != 0,
		CustomDoorOpen: s>>17&1 != 0,
		GPSUsed:        s>>18&1 != 0,
		BeidouUsed:     s>>19&1 != 0,
		GLONASSUsed:    s>>20&1 != 0,
		GalileoUsed:    s>>21&1 != 0,
	}
}

// parseT0200BCDTime6 解析 6 字节 BCD：YY-MM-DD-hh-mm-ss；原因：与 JT808 时间字段一致；注释人：Cursor
func parseT0200BCDTime6(b []byte) (string, error) {
	if len(b) < 6 {
		return "", fmt.Errorf("BCD 时间长度不足")
	}
	bcd := func(x byte) int { return int(x>>4)*10 + int(x&0x0f) }
	yy := bcd(b[0])
	mm := bcd(b[1])
	dd := bcd(b[2])
	hh := bcd(b[3])
	mi := bcd(b[4])
	ss := bcd(b[5])
	if mm < 1 || mm > 12 || dd < 1 || dd > 31 || hh > 23 || mi > 59 || ss > 59 {
		return "", fmt.Errorf("BCD 时间非法")
	}
	return fmt.Sprintf("20%02d-%02d-%02d %02d:%02d:%02d", yy, mm, dd, hh, mi, ss), nil
}

func (h *T0200) parseExtras(body []byte, off int) error {
	h.Extras = nil
	for off < len(body) {
		if off+2 > len(body) {
			return fmt.Errorf("附加信息项头不完整，偏移 %d", off)
		}
		id := body[off]
		elen := int(body[off+1])
		off += 2
		if elen < 0 || off+elen > len(body) {
			return fmt.Errorf("附加信息长度非法 id=0x%02X len=%d 偏移=%d", id, elen, off)
		}
		val := append([]byte(nil), body[off:off+elen]...)
		off += elen
		h.Extras = append(h.Extras, T0200Extra{ID: id, Len: uint8(elen), Value: val})
		h.applyKnownExtra(id, val)
	}
	return nil
}

func (h *T0200) applyKnownExtra(id uint8, val []byte) {
	switch id {
	case 0x01: // 里程，DWORD，0.1 km
		if len(val) >= 4 {
			v := binary.BigEndian.Uint32(val[:4])
			f := float64(v) / 10.0
			h.MileageKm = &f
		}
	case 0x02: // 油量，WORD，0.1 L
		if len(val) >= 2 {
			v := binary.BigEndian.Uint16(val[:2])
			f := float64(v) / 10.0
			h.FuelLiters = &f
		}
	case 0x03: // 行驶记录速度，WORD，0.1 km/h
		if len(val) >= 2 {
			v := binary.BigEndian.Uint16(val[:2])
			f := float64(v) / 10.0
			h.RecordSpeedKmh = &f
		}
	case 0x04: // 人工确认报警事件编号，WORD
		if len(val) >= 2 {
			v := binary.BigEndian.Uint16(val[:2])
			h.AlarmConfirmNo = &v
		}
	case 0x11: // 超速报警附加信息：位置类型 1 + 区域/路段 ID 4
		if len(val) >= 5 {
			h.Overspeed = &T0200OverspeedExtra{
				LocationType:  val[0],
				AreaOrRouteID: binary.BigEndian.Uint32(val[1:5]),
			}
		}
	case 0x14: // 扩展车辆信号状态，DWORD
		if len(val) >= 4 {
			v := binary.BigEndian.Uint32(val[:4])
			h.ExtendedVehicle = &v
		}
	case 0x15: // IO 状态，WORD
		if len(val) >= 2 {
			v := binary.BigEndian.Uint16(val[:2])
			h.IOStatus = &v
		}
	case 0x16: // 模拟量，DWORD
		if len(val) >= 4 {
			v := binary.BigEndian.Uint32(val[:4])
			h.AnalogInput = &v
		}
	case 0x17: // 无线通信网络信号强度
		if len(val) >= 1 {
			b := val[0]
			h.SignalStrength = &b
		}
	case 0x18: // GNSS 定位卫星数
		if len(val) >= 1 {
			b := val[0]
			h.GNSSSatelliteNum = &b
		}
	case 0x25: // 扩展车辆信号状态（另一版本/扩展）
		if len(val) >= 4 {
			v := binary.BigEndian.Uint32(val[:4])
			h.ExtendedVehicle2 = &v
		}
	}
}

func (h *T0200) OnMsg(conn net.Conn) {
	logJT808DecodedJSON("收到终端位置汇报 0x0200（JSON）", h)
	frame1 := Get8001Buf(h.JTMessage, 0)
	slog.Info("【T0200】回复8001", slog.Any("frame", hex.EncodeToString(frame1)))
	_, _ = conn.Write(frame1)
}

func init() {
	codec := &T0200{}
	codec.MsgID = P0200
	RegisterDecode(codec)
	RegisterEncode(codec)
}

func (h *T0200) Parse(jtMsg *JTMessage) error {
	h.JTMessage = *jtMsg
	body := jtMsg.Body
	if len(body) < T0200BaseLen {
		return fmt.Errorf("0x0200 消息体过短: 至少 %d 字节，实际 %d", T0200BaseLen, len(body))
	}

	h.Alarm = binary.BigEndian.Uint32(body[0:4])
	h.Status = binary.BigEndian.Uint32(body[4:8])
	h.StatusBits = decodeT0200StatusBits(h.Status)

	h.LatitudeRaw = binary.BigEndian.Uint32(body[8:12])
	h.LongitudeRaw = binary.BigEndian.Uint32(body[12:16])
	lat := float64(h.LatitudeRaw) / 1e6
	lon := float64(h.LongitudeRaw) / 1e6
	if h.StatusBits.SouthLatitude {
		lat = -lat
	}
	if h.StatusBits.WestLongitude {
		lon = -lon
	}
	h.Latitude = lat
	h.Longitude = lon

	h.Altitude = binary.BigEndian.Uint16(body[16:18])
	h.SpeedRaw = binary.BigEndian.Uint16(body[18:20])
	h.SpeedKmh = float64(h.SpeedRaw) / 10.0
	h.Direction = binary.BigEndian.Uint16(body[20:22])

	dt, err := parseT0200BCDTime6(body[22:28])
	if err != nil {
		return fmt.Errorf("0x0200 时间字段: %w", err)
	}
	h.DateTime = dt

	// 清空可选解析字段（复用 decoder 实例时由工厂 New，仍显式清理）
	h.MileageKm = nil
	h.FuelLiters = nil
	h.RecordSpeedKmh = nil
	h.AlarmConfirmNo = nil
	h.SignalStrength = nil
	h.GNSSSatelliteNum = nil
	h.ExtendedVehicle = nil
	h.IOStatus = nil
	h.AnalogInput = nil
	h.ExtendedVehicle2 = nil
	h.Overspeed = nil

	if len(body) > T0200BaseLen {
		if err := h.parseExtras(body, T0200BaseLen); err != nil {
			return err
		}
	}

	slog.Info("位置上报 0x0200",
		slog.Uint64("alarm", uint64(h.Alarm)),
		slog.Uint64("status", uint64(h.Status)),
		slog.Float64("lat", h.Latitude),
		slog.Float64("lon", h.Longitude),
		slog.Uint64("alt", uint64(h.Altitude)),
		slog.Float64("speedKmh", h.SpeedKmh),
		slog.Uint64("dir", uint64(h.Direction)),
		slog.String("time", h.DateTime),
		slog.Int("extras", len(h.Extras)),
	)
	return nil
}

func (h *T0200) Encode() []byte {
	return []byte{0x02, 0x00}
}

func (h *T0200) GetMsgId() MsgId {
	return P0200
}
