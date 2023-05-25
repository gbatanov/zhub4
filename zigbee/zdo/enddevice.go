/*
GSB, 2023
gbatanov@yandex.ru
*/
package zdo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
	"zhub4/zigbee/zdo/zcl"
)

type DeviceInfo struct {
	deviceType   uint8
	manufacturer string
	productCode  string // model
	engName      string // name for Grafana
	humanName    string
	powerSource  zcl.PowerSource
	Available    uint8 // include in prod configuration
	test         uint8 //include in test configuration
}

// MAC Address,Type, Vendor,Model, GrafanaName, Human name, Power source,available,test
var KNOWN_DEVICES map[uint64]DeviceInfo = map[uint64]DeviceInfo{
	// Датчики протечки воды
	0x00158d0006e469a4: {5, "Aqara", "SJCGQ11LM", "Протечка1", "Датчик протечки 1 (туалет)", zcl.PowerSource_BATTERY, 1, 0},
	0x00158d0006f8fc61: {5, "Aqara", "SJCGQ11LM", "Протечка2", "Датчик протечки 2 (кухня)", zcl.PowerSource_BATTERY, 1, 0},
	0x00158d0006b86b79: {5, "Aqara", "SJCGQ11LM", "Протечка3", "Датчик протечки 3 (ванна)", zcl.PowerSource_BATTERY, 1, 0},
	0x00158d0006ea99db: {5, "Aqara", "SJCGQ11LM", "Протечка4", "Датчик протечки 4 (кухня)", zcl.PowerSource_BATTERY, 1, 0},
	// реле
	0x54ef44100019335b: {9, "Aqara", "SSM-U01", "Реле1", "Реле1", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	0x54ef441000193352: {9, "Aqara", "SSM-U01", "Стиралка", "Реле2(Стиральная машина)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	0x54ef44100018b523: {9, "Aqara", "SSM-U01", "ШкафСвет", "Реле3(Шкаф, подсветка)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	0x54ef4410001933d3: {9, "Aqara", "SSM-U01", "КоридорСвет", "Реле4(Свет в коридоре)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	0x54ef4410005b2639: {9, "Aqara", "SSM-U01", "ТулетЗанят", "Реле5(Туалет занят)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	0x54ef441000609dcc: {9, "Aqara", "SSM-U01", "Реле6", "Реле6", zcl.PowerSource_SINGLE_PHASE, 1, 1},
	0x00158d0009414d7e: {11, "Aqara", "Double", "КухняСвет/КухняВент", "Реле 7(Свет/Вентилятор кухня)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	// Умные розетки
	0x70b3d52b6001b4a4: {10, "Girier", "TS011F", "Розетка1", "Розетка 1", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	0x70b3d52b6001b5d9: {10, "Girier", "TS011F", "Розетка2", "Розетка 2(Зарядники)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	0x70b3d52b60022ac9: {10, "Girier", "TS011F", "Розетка3", "Розетка 3(Лампы в десткой)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	0x70b3d52b60022cfd: {10, "Girier", "TS011F", "Розетка3", "Розетка 4(Паяльник)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	// краны
	0xa4c138d9758e1dcd: {6, "TUYA", "Valve", "КранГВ", "Кран1 ГВ", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	0xa4c138373e89d731: {6, "TUYA", "Valve", "КранХВ", "Кран2 ХВ", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	// датчики движения и/или освещения
	0x00124b0025137475: {2, "Sonoff", "SNZB-03", "КоридорДвижение", "Датчик движения 1 (коридор)", zcl.PowerSource_BATTERY, 1, 0},
	0x00124b0024455048: {2, "Sonoff", "SNZB-03", "КомнатаДвижение", "Датчик движения 2 (комната)", zcl.PowerSource_BATTERY, 1, 0},
	0x00124b002444d159: {2, "Sonoff", "SNZB-03", "Движение3", "Датчик движения 3 ", zcl.PowerSource_BATTERY, 1, 0},
	0x00124b0009451438: {4, "Custom", "CC2530", "КухняДвижение", "Датчик присутствия 1 (кухня)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	0x00124b0014db2724: {4, "Custom", "CC2530", "ПрихожаяДвижение", "Датчик движение + освещение (прихожая)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	0x0c4314fffe17d8a8: {8, "IKEA", "E1745", "ИкеаДвижение", "Датчик движения IKEA", zcl.PowerSource_BATTERY, 0, 1},
	0x00124b0007246963: {4, "Custom", "CC2530", "ДетскаяДвижение", "Датчик движение + освещение (детская)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	// Датчики открытия дверей
	0x00124b0025485ee6: {3, "Sonoff", "SNZB-04", "ТуалетДатчик", "Датчик открытия 1 (туалет)", zcl.PowerSource_BATTERY, 1, 0},
	0x00124b002512a60b: {3, "Sonoff", "SNZB-04", "ШкафДатчик", "Датчик открытия 2 (шкаф, подсветка)", zcl.PowerSource_BATTERY, 1, 0},
	0x00124b00250bba63: {3, "Sonoff", "SNZB-04", "ЯщикДатчик", "Датчик открытия 3 (ящик)", zcl.PowerSource_BATTERY, 1, 0},
	// Кнопки
	0x8cf681fffe0656ef: {7, "IKEA", "E1743", "КнопкаИкеа", "Кнопка ИКЕА", zcl.PowerSource_BATTERY, 0, 1},
	0x00124b0028928e8a: {1, "Sonoff", "SNZB-01", "Кнопка1", "Кнопка Sonoff 1", zcl.PowerSource_BATTERY, 1, 0},
	0x00124b00253ba75f: {1, "Sonoff", "SNZB-01", "Кнопка2", "Кнопка Sonoff 2", zcl.PowerSource_BATTERY, 1, 0},
	// Датчики климата
	0x00124b000b1bb401: {4, "GSB", "CC2530", "КлиматБалкон", "Датчик климата (балкон)", zcl.PowerSource_BATTERY, 1, 0},
}

// Input clusters can have commands sent to them to perform actions, where as output clusters instead send these commands to a bound device.
var DEVICE_TYPES map[uint8]string = map[uint8]string{
	1: "SonoffButton",       // 1 endpoint (3 input cluster - 0x0000, 0x0001, 0x0003,        2 output cluster - 0x0003, 0x0006 )
	2: "SonoffMotionSensor", // 1 endpoint (4 input cluster - 0x0000, 0x0001, 0x0003,0x0500, 1 output cluster - 0x0003,        )
	3: "SonoffDoorSensor",   // 1 endpoint (4 input cluster - 0x0000, 0x0001, 0x0003,0x0500, 1 output cluster - 0x0003,       )
	4: "Custom",
	5: "WaterSensor", // 1 endpoint (3 input cluster - 0x0000, 0x0001, 0x0003,        1 output cluster - 0x0019        )
	6: "WaterValve",
	7: "IkeaButton",       // 1 endpoint (7 input cluster - 0x0000, 0x0001, 0x0003,0x0009,0x0020,0x1000,0xfc7c 7 output cluster - 0x0003,0x0004,0x0006,0x0008,0x0019,0x0102, 0x1000)
	8: "IkeaMotionSensor", // 1 endpoint (7 input cluster - 0x0000, 0x0001, 0x0003,0x0009,0x0020,0x1000,0xfc7c 6 output cluster - 0x0003,0x0004,0x0006,0x0008,0x0019,0x1000 )
	9: "RelayAqara",       // 5 endpoint 1 - (8 input cluster - 0x0000, 0x0002, 0x0003,0x0004,0x0005,0x0009,0x000a,0xfcc0 3 output cluster - 0x000a,0x0019,0xffff)
	//   endpoint 21, 31 - (1 input cluster - 0x0012 )
	//   endpoint 41 - (1 input cluster - 0x000c )
	//   endpoint 242 - (1 output cluster 0x0021)
	10: "SmartPlug",        // 1 endpoint (8 input cluster - 0x0000, 0x0003, 0x0004, 0x0005, 0x0006, 0x0702,0x0b04, 0xe001,  )
	11: "RelayAqaraDouble", // 2 endpoint 1 - {11 input cluster - 0x0000, 0x0001,0x0002, 0x0003,0x0004,0x0005,0x0006,0x000a,0x0010,0x0b04,0x000c 2 output cluster - 0x000a,0x0019}
} //            2 -?

// List of devices that are turned off by long pressing the Sonoff1 button
// I use the same list for forced shutdown in the mode "No one is at home"
var OFF_LIST []uint64 = []uint64{
	0x54ef4410001933d3, // light in coridor
	0x00158d0009414d7e, // light and ventilation in kitchen
	0x54ef44100018b523, // cabinet in room(backlighting)
	0x54ef4410005b2639, // toilet is busy
	0x70b3d52b6001b4a4, // SmartPlug 1
	0x70b3d52b60022ac9, // SmartPlug 3
	0x70b3d52b60022cfd, // SmartPlug 4
}

/*
	doesn't use in this version

// List of devices to display in Grafana

	var PROM_MOTION_LIST []uint64 = []uint64{
		0x00124b0025137475, // coridor
		0x00124b0014db2724, // hallway
		0x00124b0009451438, // kitchen
		0x00124b0024455048, // room
		0x00124b002444d159, // children's room
		0x00124b0007246963, // balconen
	}

	var PROM_DOOR_LIST []uint64 = []uint64{
		0x00124b0025485ee6, // toilet
	}

	var PROM_RELAY_LIST []uint64 = []uint64{
		0x00158d0009414d7e, // light/ventilation in kitchen
		0x54ef4410001933d3, // light in coridor
		0x54ef4410005b2639} // toilet is busy
*/
type BatteryParams struct {
	level   uint8
	voltage float32
}
type EndDevice struct {
	MacAddress      uint64
	ShortAddress    uint16
	Di              DeviceInfo
	modelIdentifier string
	linkQuality     uint8
	lastSeen        time.Time
	lastAction      time.Time
	state           string  // status string value, dependent on device type
	state2          string  // channel2 status string value, dependent on device type
	mainVoltage     float32 // high voltage instant|RMS value
	current         float32 //
	battery         BatteryParams
	temperature     int8
	humidity        int8
	luminocity      int8 // high/low 1/0
	pressure        float32
	motionState     int8
}

func End_device_create(macAddress uint64, shortAddress uint16) *EndDevice {
	ed := EndDevice{MacAddress: macAddress, ShortAddress: shortAddress}
	ed.Di = KNOWN_DEVICES[macAddress]
	ed.modelIdentifier = ""
	ed.linkQuality = 0
	ed.lastSeen = time.Time{} // time.isZero - time is not initialized
	ed.lastAction = time.Time{}
	ed.state = "Unknown"
	ed.state2 = "Unknown"
	ed.mainVoltage = -100.0
	ed.current = -100.0
	ed.battery = BatteryParams{level: 0, voltage: -100.0}
	ed.temperature = -100
	ed.humidity = -100
	ed.luminocity = -100
	ed.pressure = -100
	ed.motionState = -1

	return &ed
}

func (ed EndDevice) Get_mac_address() uint64 {
	return ed.MacAddress
}
func (ed *EndDevice) Set_linkquality(quality uint8) {
	ed.linkQuality = quality
}
func (ed *EndDevice) Set_last_seen(tm time.Time) {
	ed.lastSeen = tm
}
func (ed *EndDevice) Set_last_action(tm time.Time) {
	ed.lastAction = tm
}
func (ed *EndDevice) Set_manufacturer(value string) {
	ed.Di.manufacturer = value
}
func (ed *EndDevice) Set_model_identifier(value string) {
	ed.modelIdentifier = value
}
func (ed *EndDevice) Set_product_code(value string) {
	ed.Di.productCode = value
}
func (ed *EndDevice) Set_power_source(value uint8) {
	ed.Di.powerSource = zcl.PowerSource(value)
}
func (ed *EndDevice) Set_mains_voltage(value float32) {
	ed.mainVoltage = value
}
func (ed *EndDevice) Get_mains_voltage() float32 {
	return ed.mainVoltage
}
func (ed *EndDevice) Set_current(value float32) {
	ed.current = value
}
func (ed *EndDevice) Get_current() float32 {
	return ed.current
}

// charge level, battery voltage
func (ed *EndDevice) Set_battery_params(value1 uint8, value2 float32) {
	if value1 > 0 {
		ed.battery.level = value1
	}
	if value2 > 0 {
		ed.battery.voltage = value2
	}
}
func (ed *EndDevice) Set_temperature(value int8) {
	ed.temperature = value
}
func (ed *EndDevice) Set_luminocity(value int8) {
	ed.luminocity = value
}
func (ed *EndDevice) Get_luminocity() int8 {
	return ed.luminocity
}
func (ed *EndDevice) Set_humidity(value int8) {
	ed.humidity = value
}
func (ed *EndDevice) Set_pressure(value float32) {
	ed.pressure = value
}

func (ed EndDevice) Get_human_name() string {
	return ed.Di.humanName
}

func (ed EndDevice) Get_device_type() uint8 {
	return ed.Di.deviceType
}

func (ed *EndDevice) Set_current_state(state string, channel uint8) {
	if channel == 1 {
		ed.state = state
	} else if channel == 2 {
		ed.state2 = state
	}
}
func (ed EndDevice) Get_current_state(channel uint8) string {
	if channel == 1 {
		return ed.state
	} else if channel == 2 {
		return ed.state2
	}
	return "Unknown"
}
func (ed EndDevice) Get_motion_state() int8 {
	return ed.motionState
}
func (ed EndDevice) Set_motion_state(state uint8) {
	if state == 0 || state == 1 {
		ed.motionState = int8(state)
	}
}

func (ed *EndDevice) Bytes_to_float32(src []byte) (float32, error) {

	if len(src) != 4 {
		return 0.0, errors.New("bad source slice")
	}
	var value float32
	buff := bytes.NewReader(src)
	err := binary.Read(buff, binary.LittleEndian, &value)
	if err != nil {
		return 0.0, err
	}
	return value, nil
}
