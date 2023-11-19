/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2023 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
*/
package zdo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

const RELAY_1 = uint64(0x54ef44100019335b)
const RELAY_2_WASH = uint64(0x54ef441000193352)
const RELAY_3_CAB_LIGHT = uint64(0x54ef44100018b523)
const RELAY_4_CORIDOR_LIGHT = uint64(0x54ef4410001933d3)
const RELAY_5_TOILET = uint64(0x54ef4410005b2639)
const RELAY_6_ROOM_LIGHT = uint64(0x54ef441000609dcc)
const RELAY_7_KITCHEN = uint64(0x00158d0009414d7e)

const PLUG_1 = uint64(0x70b3d52b6001b4a4)
const PLUG_2_CHARGER = uint64(0x70b3d52b6001b5d9)
const PLUG_3_NURSERY_LIGHT = uint64(0x70b3d52b60022ac9)
const PLUG_4_SOLDER = uint64(0x70b3d52b60022cfd)

const VALVE_HOT_WATER = uint64(0xa4c138d9758e1dcd)
const VALVE_COLD_WATER = uint64(0xa4c138373e89d731)

const MOTION_1_CORIDOR = uint64(0x00124b0025137475)
const MOTION_2_ROOM = uint64(0x00124b0024455048)
const MOTION_3_CORIDOR = uint64(0x00124b002444d159)
const MOTION_4_NURSERY = uint64(0x00124b002a535b66)
const MOTION_5_KITCHEN = uint64(0x00124b002a507fe2)
const PRESENCE_1_KITCHEN = uint64(0x00124b0009451438)
const MOTION_LIGHT_CORIDOR = uint64(0x00124b0014db2724)
const MOTION_IKEA = uint64(0x0c4314fffe17d8a8)
const MOTION_LIGHT_NURSERY = uint64(0x00124b0007246963)

const DOOR_1_TOILET = uint64(0x00124b0025485ee6)
const DOOR_2_CAB = uint64(0x00124b002512a60b)
const DOOR_3_BOX = uint64(0x00124b00250bba63)

const BUTTON_IKEA = uint64(0x8cf681fffe0656ef)
const BUTTON_SONOFF_1 = uint64(0x00124b0028928e8a)
const BUTTON_SONOFF_2 = uint64(0x00124b00253ba75f)

const CLIMAT_BALCON = uint64(0x00124b000b1bb401)

const WATER_LEAK_1 = uint64(0x00158d0006e469a4)
const WATER_LEAK_2 = uint64(0x00158d0006f8fc61)
const WATER_LEAK_3 = uint64(0x00158d0006b86b79)
const WATER_LEAK_4 = uint64(0x00158d0006ea99db)

type DeviceInfo struct {
	deviceType   uint8
	manufacturer string
	productCode  string // model
	engName      string // name for Grafana
	humanName    string
	PowerSource  zcl.PowerSource
	Available    uint8 // include in prod configuration
	Test         uint8 //include in test configuration
}

// MAC Address,Type, Vendor,Model, GrafanaName, Human name, Power source,available,test
var KNOWN_DEVICES map[uint64]DeviceInfo = map[uint64]DeviceInfo{
	// Датчики протечки воды
	WATER_LEAK_1: {5, "Aqara", "SJCGQ11LM", "Протечка1", "Датчик протечки 1 (туалет)", zcl.PowerSource_BATTERY, 1, 0},
	WATER_LEAK_2: {5, "Aqara", "SJCGQ11LM", "Протечка2", "Датчик протечки 2 (кухня)", zcl.PowerSource_BATTERY, 1, 0},
	WATER_LEAK_3: {5, "Aqara", "SJCGQ11LM", "Протечка3", "Датчик протечки 3 (ванна)", zcl.PowerSource_BATTERY, 1, 0},
	WATER_LEAK_4: {5, "Aqara", "SJCGQ11LM", "Протечка4", "Датчик протечки 4 (кухня)", zcl.PowerSource_BATTERY, 1, 0},
	// реле
	//	RELAY_1:               {9, "Aqara", "SSM-U01", "Реле1", "Реле1", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	RELAY_2_WASH: {9, "Aqara", "SSM-U01", "Стиралка", "Реле2(Стиральная машина)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	//	RELAY_3_CAB_LIGHT:     {9, "Aqara", "SSM-U01", "ШкафСвет", "Реле3(Шкаф, подсветка)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	RELAY_4_CORIDOR_LIGHT: {9, "Aqara", "SSM-U01", "КоридорСвет", "Реле4(Свет в коридоре)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	RELAY_5_TOILET:        {9, "Aqara", "SSM-U01", "ТулетЗанят", "Реле5(Туалет занят)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	RELAY_6_ROOM_LIGHT:    {9, "Aqara", "SSM-U01", "Реле6", "Реле6 (Свет комната)", zcl.PowerSource_SINGLE_PHASE, 1, 1},
	RELAY_7_KITCHEN:       {11, "Aqara", "Double", "КухняСвет/КухняВент", "Реле 7(Свет/Вентилятор кухня)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	// Умные розетки
	//	PLUG_1:               {10, "Girier", "TS011F", "Розетка1", "Розетка 1", zcl.PowerSource_SINGLE_PHASE, 1, 1},
	PLUG_2_CHARGER:       {10, "Girier", "TS011F", "Розетка2", "Розетка 2(Зарядники)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	PLUG_3_NURSERY_LIGHT: {10, "Girier", "TS011F", "Розетка3", "Розетка 3(Лампы в детской)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	//	PLUG_4_SOLDER:        {10, "Girier", "TS011F", "Розетка4", "Розетка 4(Паяльник)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	// краны
	VALVE_HOT_WATER:  {6, "TUYA", "Valve", "КранГВ", "Кран1 ГВ", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	VALVE_COLD_WATER: {6, "TUYA", "Valve", "КранХВ", "Кран2 ХВ", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	// датчики движения и/или освещения
	MOTION_1_CORIDOR: {2, "Sonoff", "SNZB-03", "КоридорДвижение", "Датчик движения 1 (коридор)", zcl.PowerSource_BATTERY, 1, 0},
	MOTION_2_ROOM:    {2, "Sonoff", "SNZB-03", "КомнатаДвижение", "Датчик движения 2 (комната)", zcl.PowerSource_BATTERY, 1, 0},
	MOTION_3_CORIDOR: {2, "Sonoff", "SNZB-03", "Движение3", "Датчик движения 3(коридор) ", zcl.PowerSource_BATTERY, 1, 0},
	MOTION_4_NURSERY: {2, "Sonoff", "SNZB-03", "ДетскаяДвижение4", "Датчик движения 4 (детская)", zcl.PowerSource_BATTERY, 1, 0},
	//	MOTION_5_KITCHEN:     {2, "Sonoff", "SNZB-03", "КухняДвижение", "Датчик движения 5 (кухня)", zcl.PowerSource_BATTERY, 1, 0},
	PRESENCE_1_KITCHEN: {4, "Custom", "CC2530", "КухняПрисутствие", "Датчик присутствия 1 (кухня)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	//	MOTION_LIGHT_CORIDOR: {4, "Custom", "CC2530", "ПрихожаяДвижение", "Датчик движение + освещение (прихожая)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	//	MOTION_IKEA:          {8, "IKEA", "E1745", "ИкеаДвижение", "Датчик движения IKEA", zcl.PowerSource_BATTERY, 1, 1},
	MOTION_LIGHT_NURSERY: {4, "Custom", "CC2530", "ДетскаяДвижение", "Датчик движение + освещение (детская)", zcl.PowerSource_SINGLE_PHASE, 1, 0},
	// Датчики открытия дверей
	DOOR_1_TOILET: {3, "Sonoff", "SNZB-04", "ТуалетДатчик", "Датчик открытия 1 (туалет)", zcl.PowerSource_BATTERY, 1, 0},
	//	DOOR_2_CAB:    {3, "Sonoff", "SNZB-04", "ШкафДатчик", "Датчик открытия 2 (шкаф, подсветка)", zcl.PowerSource_BATTERY, 1, 0},
	//	DOOR_3_BOX:    {3, "Sonoff", "SNZB-04", "ЯщикДатчик", "Датчик открытия 3 (ящик)", zcl.PowerSource_BATTERY, 1, 0},
	// Кнопки
	//	BUTTON_IKEA:     {7, "IKEA", "E1743", "КнопкаИкеа", "Кнопка ИКЕА", zcl.PowerSource_BATTERY, 1, 1},
	//	BUTTON_SONOFF_1: {1, "Sonoff", "SNZB-01", "Кнопка1", "Кнопка Sonoff 1", zcl.PowerSource_BATTERY, 1, 0},
	BUTTON_SONOFF_2: {1, "Sonoff", "SNZB-01", "Кнопка2", "Кнопка Sonoff 2", zcl.PowerSource_BATTERY, 1, 0},
	// Датчики климата
	CLIMAT_BALCON: {4, "GSB", "CC2530", "КлиматБалкон", "Датчик климата (балкон)", zcl.PowerSource_BATTERY, 1, 0},
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
	9: "RelayAqara",       // 5 endpoint
	// endpoint  1 - (8 input cluster -  0x0000, 0x0002, 0x0003,0x0004,0x0005,0x0009,0x000a,0xfcc0
	//  	          3 output cluster - 0x000a, 0x0019, 0xffff)
	// endpoint 21, 31 - (1 input cluster -  0x0012 )
	// endpoint 41 -     (1 input cluster -  0x000c )
	// endpoint 242 -    (1 output cluster - 0x0021)
	10: "SmartPlug",        // 1 endpoint (8 input cluster - 0x0000, 0x0003, 0x0004, 0x0005, 0x0006, 0x0702,0x0b04, 0xe001,  )
	11: "RelayAqaraDouble", // 2 endpoint 1 - {11 input cluster - 0x0000, 0x0001,0x0002, 0x0003,0x0004,0x0005,0x0006,0x000a,0x0010,0x0b04,0x000c 2 output cluster - 0x000a,0x0019}
} //            2 -?

// List of devices that are turned off by long pressing the Sonoff1 button
// I use the same list for forced shutdown in the mode "No one is at home"
var OFF_LIST []uint64 = []uint64{
	RELAY_4_CORIDOR_LIGHT, // light in coridor
	RELAY_7_KITCHEN,       // light and ventilation in kitchen
	RELAY_3_CAB_LIGHT,     // cabinet in room(backlighting)
	RELAY_5_TOILET,        // toilet is busy
	PLUG_1,                // SmartPlug 1
	PLUG_3_NURSERY_LIGHT,  // SmartPlug 3
	PLUG_4_SOLDER,         // SmartPlug 4
	RELAY_6_ROOM_LIGHT,    // Relay 6 room light
}

// List of devices to display in Grafana

var PROM_MOTION_LIST []uint64 = []uint64{
	MOTION_1_CORIDOR,     // coridor
	MOTION_LIGHT_CORIDOR, // hallway
	PRESENCE_1_KITCHEN,   // kitchen presence sensor
	MOTION_5_KITCHEN,     // kitchen onoff sensor
	MOTION_2_ROOM,        // room
	MOTION_3_CORIDOR,     // coridor
	MOTION_4_NURSERY,     // nursery
	MOTION_LIGHT_NURSERY, // nursery
	MOTION_IKEA,          // IKEA motion sensor
}

var PROM_DOOR_LIST []uint64 = []uint64{
	DOOR_1_TOILET, // toilet
}

var PROM_RELAY_LIST []uint64 = []uint64{
	RELAY_7_KITCHEN,       // light/ventilation in kitchen
	RELAY_4_CORIDOR_LIGHT, // light in coridor
	RELAY_5_TOILET}        // toilet is busy

type BatteryParams struct {
	level   uint8
	voltage float64
}
type ElectricParams struct {
	mainVoltage float64 // high voltage instant|RMS value
	current     float64 //
	power       float64 // instant power
	energy      float64 // energy
}

type EndDevice struct {
	MacAddress      uint64
	ShortAddress    uint16
	Di              DeviceInfo
	modelIdentifier string
	linkQuality     uint8
	lastSeen        time.Time
	lastAction      time.Time
	state           string // status string value, dependent on device type
	state2          string // channel2 status string value, dependent on device type
	battery         BatteryParams
	electric        ElectricParams
	temperature     int8
	humidity        int8
	luminocity      int8 // high/low 1/0
	pressure        float64
	motionState     int8
	ChargerOn       bool
}

func EndDeviceCreate(macAddress uint64, shortAddress uint16) *EndDevice {
	ed := EndDevice{MacAddress: macAddress, ShortAddress: shortAddress}
	ed.Di = KNOWN_DEVICES[macAddress]
	ed.modelIdentifier = ""
	ed.linkQuality = 0
	ed.lastSeen = time.Time{} // time.isZero - time is not initialized
	ed.lastAction = time.Time{}
	ed.state = "Unknown"
	ed.state2 = "Unknown"
	ed.electric = ElectricParams{
		mainVoltage: -100.0,
		current:     -100.0,
		power:       -100.0,
		energy:      -100.0}
	ed.battery = BatteryParams{level: 0, voltage: -100.0}
	ed.temperature = -100
	ed.humidity = -100
	ed.luminocity = -100
	ed.pressure = -100
	ed.motionState = -1
	ed.ChargerOn = false

	return &ed
}

func (ed EndDevice) GetPowerSource() uint8 {
	return uint8(ed.Di.PowerSource)
}
func (ed EndDevice) Get_mac_address() uint64 {
	return ed.MacAddress
}
func (ed *EndDevice) Set_linkquality(quality uint8) {
	ed.linkQuality = quality
}
func (ed *EndDevice) Get_linkquality() uint8 {
	return ed.linkQuality
}
func (ed *EndDevice) Set_last_seen(tm time.Time) {
	ed.lastSeen = tm
}
func (ed *EndDevice) Get_last_seen() time.Time {
	return ed.lastSeen
}
func (ed *EndDevice) SetLastAction(tm time.Time) {
	ed.lastAction = tm
}
func (ed *EndDevice) Get_last_action() time.Time {
	return ed.lastAction
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
func (ed *EndDevice) SetPowerSource(value uint8) {
	ed.Di.PowerSource = zcl.PowerSource(value)
}
func (ed *EndDevice) SetMainsVoltage(value float64) {
	if value > 0 {
		ed.electric.mainVoltage = value
	}
}
func (ed *EndDevice) GetMainsVoltage() float64 {
	if ed.electric.mainVoltage > 0 {
		return ed.electric.mainVoltage
	} else {
		return -100.0
	}
}
func (ed *EndDevice) SetCurrent(value float64) {
	ed.electric.current = value
}
func (ed *EndDevice) Get_current() float64 {
	return ed.electric.current
}

func (ed *EndDevice) Set_power(value float64) {
	ed.electric.power = value
}
func (ed *EndDevice) Get_power() float64 {
	return ed.electric.power
}

func (ed *EndDevice) Set_energy(value float64) {
	ed.electric.energy = value
}
func (ed *EndDevice) Get_energy() float64 {
	return ed.electric.energy
}

// charge level, battery voltage
func (ed *EndDevice) SetBatteryParams(value1 uint8, value2 float64) {
	if value1 > 0 {
		ed.battery.level = value1
	}
	if value2 > 0 {
		ed.battery.voltage = value2
	}
}
func (ed *EndDevice) Get_battery_level() uint8 {
	return ed.battery.level
}
func (ed *EndDevice) Get_battery_voltage() float64 {
	return ed.battery.voltage
}
func (ed *EndDevice) Set_temperature(value int8) {
	ed.temperature = value
}
func (ed *EndDevice) Get_temperature() int8 {
	return ed.temperature
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
func (ed *EndDevice) Get_humidity() int8 {
	return ed.humidity
}
func (ed *EndDevice) Set_pressure(value float64) {
	ed.pressure = value * 0.00750063755419211

}
func (ed *EndDevice) Get_pressure() float64 {
	return ed.pressure
}

func (ed *EndDevice) GetPromPressure() string {
	pressure := ed.Get_pressure()
	if pressure > 0 {
		return "zhub2_metrics{device=\"" + ed.Di.engName + "\",type=\"pressure\"} " + strconv.FormatFloat(pressure, 'f', 3, 64) + "\n"
	} else {
		return ""
	}
}

func (ed EndDevice) GetHumanName() string {
	return ed.Di.humanName
}

func (ed EndDevice) GetDeviceType() uint8 {
	return ed.Di.deviceType
}

func (ed *EndDevice) SetCurrentState(state string, channel uint8) {
	if channel == 1 {
		ed.state = state
	} else if channel == 2 {
		ed.state2 = state
	}
}
func (ed EndDevice) GetCurrentState(channel uint8) string {
	if channel == 1 {
		return ed.state
	} else if channel == 2 {
		return ed.state2
	}
	return "Unknown"
}
func (ed EndDevice) GetMotionState() int8 {
	return ed.motionState
}
func (ed *EndDevice) SetMotionState(state uint8) {
	if state == 0 || state == 1 {
		ed.motionState = int8(state)
	}
}
func (ed EndDevice) GetPromMotionString() string {
	state := ed.GetMotionState()
	if state < 0 || len(ed.Di.engName) == 0 {
		return ""
	}
	return "zhub2_metrics{device=\"" + ed.Di.engName + "\",type=\"motion\"} " + strconv.Itoa(int(state)) + "\n"
}

// / @brief Возвращает строку для Prometheus
// / Чтобы линия реле не сливалась с линией датчика,
// / искуственно отступаю на 0,1
// / @return
func (ed EndDevice) GetPromRelayString() string {
	state2 := ""
	state := ed.GetCurrentState(1)
	if ed.GetDeviceType() == 11 {
		state2 = ed.GetCurrentState(2)
	}
	strState := ""
	strState2 := ""
	if state == "On" {
		strState = "0.9"
	} else if state == "Off" {
		strState = "0.1"
	}
	if state2 == "On" {
		strState2 = "0.95"
	} else if state2 == "Off" {
		strState2 = "0.15"
	}
	names := strings.Split(ed.Di.engName, "/")
	name1 := names[0]
	name2 := ""
	if len(names) > 1 {
		name2 = names[1]
	}
	if len(strState) > 0 && len(name1) > 0 {
		strState = "zhub2_metrics{device=\"" + name1 + "\",type=\"relay\"} " + strState + "\n"
	}
	if len(strState2) > 0 && len(name2) > 0 {
		strState2 = "zhub2_metrics{device=\"" + name2 + "\",type=\"relay\"} " + strState2 + "\n"
	}
	return strState + strState2
}

// / @brief Возвращает строку для Prometheus
// / @return
func (ed *EndDevice) GetPromDoorString() string {
	state := ed.GetCurrentState(1)
	strState := ""
	if state == "Opened" {
		strState = "0.95"
	} else if state == "Closed" {
		strState = "0.05"
	} else {
		return ""
	}
	return "zhub2_metrics{device=\"" + ed.Di.engName + "\",type=\"door\"} " + strState + "\n"
}

func (ed *EndDevice) Bytes_to_float64(src []byte) (float64, error) {

	if len(src) != 4 {
		return 0.0, errors.New("bad source slice")
	}
	var value float32
	buff := bytes.NewReader(src)
	err := binary.Read(buff, binary.LittleEndian, &value)
	if err != nil {
		return 0.0, err
	}
	return float64(value), nil
}

func GetDevicesByType(tp uint8) []uint64 {
	var devicesList []uint64 = make([]uint64, 0)

	for adr, di := range KNOWN_DEVICES {
		if di.deviceType == tp {
			devicesList = append(devicesList, adr)
		}
	}

	return devicesList
}
