package pkg

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
)

// T0200BaseLen 鏍囧噯娑堟伅浣撴渶灏忛暱搴︼細鎶ヨ+鐘舵€?缁忕含搴?楂樼▼+閫熷害+鏂瑰悜+鏃堕棿(BCD6)锛涘師鍥狅細JT/T 808 浣嶇疆姹囨姤鍩烘湰鏁版嵁鍗曞厓锛涙敞閲婁汉锛欳ursor
const T0200BaseLen = 28

// T0200Extra 闄勫姞淇℃伅椤癸紙缁堢 ID + 闀垮害 + 鍐呭锛夛紱鍘熷洜锛欽T808 琛?23/24 TLV 鎵╁睍锛涙敞閲婁汉锛欳ursor
type T0200Extra struct {
	ID    uint8  `json:"id"`
	Len   uint8  `json:"len"`
	Value []byte `json:"value"`
}

// T0200OverspeedExtra 闄勫姞淇℃伅 0x11 瓒呴€熸姤璀︼紱鍘熷洜锛氳〃 25 鍥哄畾 5 瀛楄妭锛涙敞閲婁汉锛欳ursor
type T0200OverspeedExtra struct {
	LocationType  uint8  `json:"locationType"`  // 浣嶇疆绫诲瀷
	AreaOrRouteID uint32 `json:"areaOrRouteId"` // 鍖哄煙鎴栬矾娈?ID
}

// T0200StatusBits 鐘舵€?DWORD 甯哥敤浣嶉噴涔夛紙瑙?JT808 鐘舵€佷綅瀹氫箟锛夛紱鍘熷洜锛氫究浜?MQTT/涓氬姟鐩存帴浣跨敤锛涙敞閲婁汉锛欳ursor
type T0200StatusBits struct {
	ACCOn         bool `json:"accOn"`         // bit0 0:ACC 鍏?1:ACC 寮€
	Located       bool `json:"located"`       // bit1 0:鏈畾浣?1:瀹氫綅
	SouthLatitude bool `json:"southLatitude"` // bit2 0:鍖楃含 1:鍗楃含
	WestLongitude bool `json:"westLongitude"` // bit3 0:涓滅粡 1:瑗跨粡
	Operating     bool `json:"operating"`     // bit4 杩愯惀鐘舵€?	Encrypted     bool `json:"encrypted"`     // bit5 缁忕含搴﹀姞瀵嗭紙閮ㄥ垎鐗堟湰锛?	// bit6-7 淇濈暀
	Overload uint8 `json:"overload"` // bit8-9 瀹㈣揣鐘舵€侊紙鍙栧€间緷鍗忚鐗堟湰锛?	// bit10 娌硅矾鏂紑 bit11 鐢佃矾鏂紑 bit12 杞﹂棬鍔犻攣 鈥?缁х画灞曞紑甯哥敤浣?	FuelCutoff     bool `json:"fuelCutoff"`     // bit10
	CircuitCutoff  bool `json:"circuitCutoff"`  // bit11
	DoorLocked     bool `json:"doorLocked"`     // bit12
	FrontDoorOpen  bool `json:"frontDoorOpen"`  // bit13
	MidDoorOpen    bool `json:"midDoorOpen"`    // bit14
	RearDoorOpen   bool `json:"rearDoorOpen"`   // bit15
	DriverDoorOpen bool `json:"driverDoorOpen"` // bit16
	CustomDoorOpen bool `json:"customDoorOpen"` // bit17
	GPSUsed        bool `json:"gpsUsed"`        // bit18 0:浣跨敤鍖楁枟 1:浣跨敤 GPS锛堝父瑙佸疄鐜帮紝浠ョ粓绔负鍑嗭級
	BeidouUsed     bool `json:"beidouUsed"`     // bit19
	GLONASSUsed    bool `json:"glonassUsed"`    // bit20
	GalileoUsed    bool `json:"galileoUsed"`    // bit21
}

// T0200 浣嶇疆淇℃伅姹囨姤 0x0200锛圝T/T 808锛?type T0200 struct {
	JTMessage
	Alarm        uint32          `json:"alarm"`
	Status       uint32          `json:"status"`
	StatusBits   T0200StatusBits `json:"statusBits"`
	LatitudeRaw  uint32          `json:"latitudeRaw"`  // 鐧句竾鍒嗕箣涓€搴︼紝鏃犵鍙峰箙搴?	LongitudeRaw uint32          `json:"longitudeRaw"` // 鐧句竾鍒嗕箣涓€搴︼紝鏃犵鍙峰箙搴?	Latitude     float64         `json:"latitude"`     // 搴︼紝宸叉寜鐘舵€佷綅 N/S銆丒/W 鍙栫鍙?	Longitude    float64         `json:"longitude"`    // 搴?	Altitude     uint16          `json:"altitude"`     // 绫?	SpeedRaw     uint16          `json:"speedRaw"`     // 0.1 km/h
	SpeedKmh     float64         `json:"speedKmh"`     // km/h
	Direction    uint16          `json:"direction"`    // 0鈥?59锛屾鍖椾负 0
	DateTime     string          `json:"dateTime"`     // BCD 鏃堕棿 鏈湴鍗忚鏃跺尯锛堜竴鑸负鍖椾含鏃堕棿锛?	Extras       []T0200Extra    `json:"extras,omitempty"`
	// 甯歌闄勫姞淇℃伅瑙ｆ瀽锛?x01 閲岀▼銆?x02 娌归噺銆?x17 淇″彿寮哄害銆?x18 鍗槦鏁扮瓑锛?	MileageKm        *float64             `json:"mileageKm,omitempty"`               // 0x01锛孌WORD锛?.1km
	FuelLiters       *float64             `json:"fuelLiters,omitempty"`              // 0x02锛學ORD锛?.1L
	RecordSpeedKmh   *float64             `json:"recordSpeedKmh,omitempty"`          // 0x03锛學ORD锛?.1km/h
	AlarmConfirmNo   *uint16              `json:"alarmConfirmEventId,omitempty"`     // 0x04锛學ORD
	SignalStrength   *uint8               `json:"signalStrength,omitempty"`          // 0x17锛孊YTE
	GNSSSatelliteNum *uint8               `json:"gnssSatelliteNum,omitempty"`        // 0x18锛孊YTE
	ExtendedVehicle  *uint32              `json:"extendedVehicleSignals,omitempty"`  // 0x14锛孌WORD
	IOStatus         *uint16              `json:"ioStatus,omitempty"`                // 0x15锛學ORD
	AnalogInput      *uint32              `json:"analogInput,omitempty"`             // 0x16锛孌WORD
	ExtendedVehicle2 *uint32              `json:"extendedVehicleSignals2,omitempty"` // 0x25锛孌WORD
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

// parseT0200BCDTime6 瑙ｆ瀽 6 瀛楄妭 BCD锛歒Y-MM-DD-hh-mm-ss锛涘師鍥狅細涓?JT808 鏃堕棿瀛楁涓€鑷达紱娉ㄩ噴浜猴細Cursor
func parseT0200BCDTime6(b []byte) (string, error) {
	if len(b) < 6 {
		return "", fmt.Errorf("BCD 鏃堕棿闀垮害涓嶈冻")
	}
	bcd := func(x byte) int { return int(x>>4)*10 + int(x&0x0f) }
	yy := bcd(b[0])
	mm := bcd(b[1])
	dd := bcd(b[2])
	hh := bcd(b[3])
	mi := bcd(b[4])
	ss := bcd(b[5])
	if mm < 1 || mm > 12 || dd < 1 || dd > 31 || hh > 23 || mi > 59 || ss > 59 {
		return "", fmt.Errorf("BCD 鏃堕棿闈炴硶")
	}
	return fmt.Sprintf("20%02d-%02d-%02d %02d:%02d:%02d", yy, mm, dd, hh, mi, ss), nil
}

func (h *T0200) parseExtras(body []byte, off int) error {
	h.Extras = nil
	for off < len(body) {
		if off+2 > len(body) {
			return fmt.Errorf("闄勫姞淇℃伅椤瑰ご涓嶅畬鏁达紝鍋忕Щ %d", off)
		}
		id := body[off]
		elen := int(body[off+1])
		off += 2
		if elen < 0 || off+elen > len(body) {
			return fmt.Errorf("闄勫姞淇℃伅闀垮害闈炴硶 id=0x%02X len=%d 鍋忕Щ=%d", id, elen, off)
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
	case 0x01: // 閲岀▼锛孌WORD锛?.1 km
		if len(val) >= 4 {
			v := binary.BigEndian.Uint32(val[:4])
			f := float64(v) / 10.0
			h.MileageKm = &f
		}
	case 0x02: // 娌归噺锛學ORD锛?.1 L
		if len(val) >= 2 {
			v := binary.BigEndian.Uint16(val[:2])
			f := float64(v) / 10.0
			h.FuelLiters = &f
		}
	case 0x03: // 琛岄┒璁板綍閫熷害锛學ORD锛?.1 km/h
		if len(val) >= 2 {
			v := binary.BigEndian.Uint16(val[:2])
			f := float64(v) / 10.0
			h.RecordSpeedKmh = &f
		}
	case 0x04: // 浜哄伐纭鎶ヨ浜嬩欢缂栧彿锛學ORD
		if len(val) >= 2 {
			v := binary.BigEndian.Uint16(val[:2])
			h.AlarmConfirmNo = &v
		}
	case 0x11: // 瓒呴€熸姤璀﹂檮鍔犱俊鎭細浣嶇疆绫诲瀷 1 + 鍖哄煙/璺 ID 4
		if len(val) >= 5 {
			h.Overspeed = &T0200OverspeedExtra{
				LocationType:  val[0],
				AreaOrRouteID: binary.BigEndian.Uint32(val[1:5]),
			}
		}
	case 0x14: // 鎵╁睍杞﹁締淇″彿鐘舵€侊紝DWORD
		if len(val) >= 4 {
			v := binary.BigEndian.Uint32(val[:4])
			h.ExtendedVehicle = &v
		}
	case 0x15: // IO 鐘舵€侊紝WORD
		if len(val) >= 2 {
			v := binary.BigEndian.Uint16(val[:2])
			h.IOStatus = &v
		}
	case 0x16: // 妯℃嫙閲忥紝DWORD
		if len(val) >= 4 {
			v := binary.BigEndian.Uint32(val[:4])
			h.AnalogInput = &v
		}
	case 0x17: // 鏃犵嚎閫氫俊缃戠粶淇″彿寮哄害
		if len(val) >= 1 {
			b := val[0]
			h.SignalStrength = &b
		}
	case 0x18: // GNSS 瀹氫綅鍗槦鏁?		if len(val) >= 1 {
			b := val[0]
			h.GNSSSatelliteNum = &b
		}
	case 0x25: // 鎵╁睍杞﹁締淇″彿鐘舵€侊紙鍙︿竴鐗堟湰/鎵╁睍锛?		if len(val) >= 4 {
			v := binary.BigEndian.Uint32(val[:4])
			h.ExtendedVehicle2 = &v
		}
	}
}

func (h *T0200) OnMsg(conn net.Conn) {
	logJT808DecodedJSON("鏀跺埌缁堢浣嶇疆姹囨姤 0x0200锛圝SON锛?, h)
	frame1, seqNo := Get8001Buf(h.JTMessage, 0)
	slog.Info("銆怲0200銆戝洖澶?001", slog.Any("frame", hex.EncodeToString(frame1)))
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
		return fmt.Errorf("0x0200 娑堟伅浣撹繃鐭? 鑷冲皯 %d 瀛楄妭锛屽疄闄?%d", T0200BaseLen, len(body))
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
		return fmt.Errorf("0x0200 鏃堕棿瀛楁: %w", err)
	}
	h.DateTime = dt

	// 娓呯┖鍙€夎В鏋愬瓧娈碉紙澶嶇敤 decoder 瀹炰緥鏃剁敱宸ュ巶 New锛屼粛鏄惧紡娓呯悊锛?	h.MileageKm = nil
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

	slog.Info("浣嶇疆涓婃姤 0x0200",
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
