package pkg

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
)

// T0200BaseLen 标准消息体最小长度；注释原因：JT/T 808 位置汇报基本数据单元；注释人：Cursor
const T0200BaseLen = 28

// T0200Extra 附加信息项；注释原因：JT808 TLV 扩展；注释人：Cursor
type T0200Extra struct {
	ID    uint8  `json:"id"`
	Len   uint8  `json:"len"`
	Value []byte `json:"value"`
}

// T0200OverspeedExtra 附加信息 0x11 超速报警；注释原因：表25 固定5字节；注释人：Cursor
type T0200OverspeedExtra struct {
	LocationType  uint8  `json:"locationType"`
	AreaOrRouteID uint32 `json:"areaOrRouteId"`
}

// T0200StatusBits 状态 DWORD 常用位释义；注释人：Cursor
type T0200StatusBits struct {
	ACCOn          bool  `json:"accOn"`
	Located        bool  `json:"located"`
	SouthLatitude  bool  `json:"southLatitude"`
	WestLongitude  bool  `json:"westLongitude"`
	Operating      bool  `json:"operating"`
	Encrypted      bool  `json:"encrypted"`
	Overload       uint8 `json:"overload"`
	FuelCutoff     bool  `json:"fuelCutoff"`
	CircuitCutoff  bool  `json:"circuitCutoff"`
	DoorLocked     bool  `json:"doorLocked"`
	FrontDoorOpen  bool  `json:"frontDoorOpen"`
	MidDoorOpen    bool  `json:"midDoorOpen"`
	RearDoorOpen   bool  `json:"rearDoorOpen"`
	DriverDoorOpen bool  `json:"driverDoorOpen"`
	CustomDoorOpen bool  `json:"customDoorOpen"`
	GPSUsed        bool  `json:"gpsUsed"`
	BeidouUsed     bool  `json:"beidouUsed"`
	GLONASSUsed    bool  `json:"glonassUsed"`
	GalileoUsed    bool  `json:"galileoUsed"`
}

// T0200 位置信息汇报 0x0200
type T0200 struct {
	JTMessage
	Alarm        uint32          `json:"alarm"`
	Status       uint32          `json:"status"`
	StatusBits   T0200StatusBits `json:"statusBits"`
	LatitudeRaw  uint32          `json:"latitudeRaw"`
	LongitudeRaw uint32          `json:"longitudeRaw"`
	Latitude     float64         `json:"latitude"`
	Longitude    float64         `json:"longitude"`
	Altitude     uint16          `json:"altitude"`
	SpeedRaw     uint16          `json:"speedRaw"`
	SpeedKmh     float64         `json:"speedKmh"`
	Direction    uint16          `json:"direction"`
	DateTime     string          `json:"dateTime"`
	Extras       []T0200Extra    `json:"extras,omitempty"`

	MileageKm        *float64             `json:"mileageKm,omitempty"`
	FuelLiters       *float64             `json:"fuelLiters,omitempty"`
	RecordSpeedKmh   *float64             `json:"recordSpeedKmh,omitempty"`
	AlarmConfirmNo   *uint16              `json:"alarmConfirmEventId,omitempty"`
	SignalStrength   *uint8               `json:"signalStrength,omitempty"`
	GNSSSatelliteNum *uint8               `json:"gnssSatelliteNum,omitempty"`
	ExtendedVehicle  *uint32              `json:"extendedVehicleSignals,omitempty"`
	IOStatus         *uint16              `json:"ioStatus,omitempty"`
	AnalogInput      *uint32              `json:"analogInput,omitempty"`
	ExtendedVehicle2 *uint32              `json:"extendedVehicleSignals2,omitempty"`
	Overspeed        *T0200OverspeedExtra `json:"overspeedExtra,omitempty"`
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

func parseT0200BCDTime6(b []byte) (string, error) {
	if len(b) < 6 {
		return "", fmt.Errorf("BCD time length insufficient")
	}
	bcd := func(x byte) int { return int(x>>4)*10 + int(x&0x0f) }
	yy, mm, dd := bcd(b[0]), bcd(b[1]), bcd(b[2])
	hh, mi, ss := bcd(b[3]), bcd(b[4]), bcd(b[5])
	if mm < 1 || mm > 12 || dd < 1 || dd > 31 || hh > 23 || mi > 59 || ss > 59 {
		return "", fmt.Errorf("BCD time value invalid")
	}
	return fmt.Sprintf("20%02d-%02d-%02d %02d:%02d:%02d", yy, mm, dd, hh, mi, ss), nil
}

func (h *T0200) parseExtras(body []byte, off int) error {
	h.Extras = nil
	for off < len(body) {
		if off+2 > len(body) {
			return fmt.Errorf("extra item header incomplete, offset %d", off)
		}
		id := body[off]
		elen := int(body[off+1])
		off += 2
		if off+elen > len(body) {
			return fmt.Errorf("extra item length invalid id=0x%02X len=%d offset=%d", id, elen, off)
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
	case 0x01:
		if len(val) >= 4 {
			v := binary.BigEndian.Uint32(val[:4])
			f := float64(v) / 10.0
			h.MileageKm = &f
		}
	case 0x02:
		if len(val) >= 2 {
			v := binary.BigEndian.Uint16(val[:2])
			f := float64(v) / 10.0
			h.FuelLiters = &f
		}
	case 0x03:
		if len(val) >= 2 {
			v := binary.BigEndian.Uint16(val[:2])
			f := float64(v) / 10.0
			h.RecordSpeedKmh = &f
		}
	case 0x04:
		if len(val) >= 2 {
			v := binary.BigEndian.Uint16(val[:2])
			h.AlarmConfirmNo = &v
		}
	case 0x11:
		if len(val) >= 5 {
			h.Overspeed = &T0200OverspeedExtra{
				LocationType:  val[0],
				AreaOrRouteID: binary.BigEndian.Uint32(val[1:5]),
			}
		}
	case 0x14:
		if len(val) >= 4 {
			v := binary.BigEndian.Uint32(val[:4])
			h.ExtendedVehicle = &v
		}
	case 0x15:
		if len(val) >= 2 {
			v := binary.BigEndian.Uint16(val[:2])
			h.IOStatus = &v
		}
	case 0x16:
		if len(val) >= 4 {
			v := binary.BigEndian.Uint32(val[:4])
			h.AnalogInput = &v
		}
	case 0x17:
		if len(val) >= 1 {
			b := val[0]
			h.SignalStrength = &b
		}
	case 0x18:
		if len(val) >= 1 {
			b := val[0]
			h.GNSSSatelliteNum = &b
		}
	case 0x25:
		if len(val) >= 4 {
			v := binary.BigEndian.Uint32(val[:4])
			h.ExtendedVehicle2 = &v
		}
	}
}

func (h *T0200) OnMsg(conn net.Conn) {
	logJT808DecodedJSON("T0200 position report", h)
	frame1, seqNo := Get8001Buf(h.JTMessage, 0)
	slog.Info("T0200 reply T8001",
		slog.String("terminalNo", h.TerminalNo),
		slog.String("recvMsgId", h.MsgID.String()),
		slog.Uint64("recvSeqNo", uint64(h.SeqNo)),
		slog.Uint64("seqNo", uint64(seqNo)),
		slog.String("frameHex", hex.EncodeToString(frame1)),
	)
	if _, err := conn.Write(frame1); err != nil {
		slog.Error("T0200 reply T8001 failed", slog.String("terminalNo", h.TerminalNo), slog.Any("err", err))
	}
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
		return fmt.Errorf("0x0200 body too short, need %d bytes, got %d", T0200BaseLen, len(body))
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
		return fmt.Errorf("0x0200 time field: %w", err)
	}
	h.DateTime = dt

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

	slog.Info("T0200 position report parsed",
		slog.String("terminalNo", h.TerminalNo),
		slog.Uint64("recvSeqNo", uint64(h.SeqNo)),
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