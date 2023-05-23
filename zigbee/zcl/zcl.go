/*
GSB, 2023
gbatanov@yandex.ru
*/
package zcl

import "fmt"

type FrameType uint8

const (
	FrameType_GLOBAL   FrameType = 0
	FrameType_SPECIFIC FrameType = 1
)

type FrameDirection uint8

const (
	FROM_CLIENT_TO_SERVER FrameDirection = 0
	FROM_SERVER_TO_CLIENT FrameDirection = 8
)

type FrameControl struct {
	Ftype                  FrameType
	ManufacturerSpecific   uint8
	Direction              FrameDirection
	DisableDefaultResponse uint8
}

type Frame struct {
	Frame_control             FrameControl
	ManufacturerCode          uint16
	TransactionSequenceNumber uint8
	Command                   uint8
	Payload                   []byte
}

type Cluster uint16

const (
	BASIC                            Cluster = 0x0000
	POWER_CONFIGURATION              Cluster = 0x0001
	DEVICE_TEMPERATURE_CONFIGURATION Cluster = 0x0002
	IDENTIFY                         Cluster = 0x0003
	GROUPS                           Cluster = 0x0004
	SCENES                           Cluster = 0x0005
	ON_OFF                           Cluster = 0x0006
	ON_OFF_SWITCH_CONFIGURATION      Cluster = 0x0007
	LEVEL_CONTROL                    Cluster = 0x0008
	ALARMS                           Cluster = 0x0009
	TIME                             Cluster = 0x000a // Attributes and commands that provide an interface to a real-time clock. (С реле aquara идет каждую минуту с одним и тем же значением)
	RSSI                             Cluster = 0x000b // Attributes and commands for exchanging location information and channel parameters among devices, and (optionally) reporting data to a centralized device that collects data from devices in the network and calculates their positions from the set of collected data.
	ANALOG_INPUT                     Cluster = 0x000c // у реле от Aquara похоже на передачу потребляемой мощности, значения идут только при включенной нагрузке, чередуются значение и нуль. При выключенной нагрузке ничего не передается.
	// передается на endpoint=15 По значению похоже на показатель потребляемой мощности.
	ANALOG_OUTPUT                    Cluster = 0x000d
	ANALOG_VALUE                     Cluster = 0x000e
	BINARY_INPUT                     Cluster = 0x000f
	BINARY_OUTPUT                    Cluster = 0x0010
	BINARY_VALUE                     Cluster = 0x0011
	MULTISTATE_INPUT                 Cluster = 0x0012
	MULTISTATE_OUTPUT                Cluster = 0x0013
	MULTISTATE_VALUE                 Cluster = 0x0014
	OTA                              Cluster = 0x0019
	POWER_PROFILE                    Cluster = 0x001a // Attributes and commands that provide an interface to the power profile of a device
	PULSE_WIDTH_MODULATION           Cluster = 0x001c //
	POLL_CONTROL                     Cluster = 0x0020 // Attributes and commands that provide an interface to control the polling of sleeping end device
	XIAOMI_SWITCH_OUTPUT             Cluster = 0x0021 // выходной кластер на реле aqara
	KEEP_ALIVE                       Cluster = 0x0025
	WINDOW_COVERING                  Cluster = 0x0102
	TEMPERATURE_MEASUREMENT          Cluster = 0x0402
	HUMIDITY_MEASUREMENT             Cluster = 0x0405
	ILLUMINANCE_MEASUREMENT          Cluster = 0x0400
	IAS_ZONE                         Cluster = 0x0500
	SIMPLE_METERING                  Cluster = 0x0702 // умная розетка
	METER_IDENTIFICATION             Cluster = 0x0b01 // Attributes and commands that provide an interface to meter identification
	ELECTRICAL_MEASUREMENTS          Cluster = 0x0b04 //
	DIAGNOSTICS                      Cluster = 0x0b05 // Attributes and commands that provide an interface to diagnostics of the ZigBee stack
	TOUCHLINK_COMISSIONING           Cluster = 0x1000 // Для устройств со светом, в другом варианте LIGHT_LINK
	TUYA_SWITCH_MODE_0               Cluster = 0xe000 // кран
	TUYA_ELECTRICIAN_PRIVATE_CLUSTER Cluster = 0xe001 // имеется у умной розетки и крана Voltage - ??
	IKEA_BUTTON_ATTR_UNKNOWN2        Cluster = 0xfc7c // Имеется у кнопки IKEA
	XIAOMI_SWITCH                    Cluster = 0xfcc0 // проприетарный кластер (Lumi) на реле Aquara, присутствует код производителя и длинная строка payload
	SMOKE_SENSOR                     Cluster = 0xfe00 // Датчик дыма, TUYA-совместимый
	UNKNOWN_CLUSTER                  Cluster = 0xffff
)

func Cluster_to_string(cl Cluster) string {
	var result string
	switch cl {
	case BASIC: // 0x0000
		result = "BASIC"
	case POWER_CONFIGURATION: //0x0001
		result = "POWER_CONFIGURATION"
	case DEVICE_TEMPERATURE_CONFIGURATION: //0x0002
		result = "DEVICE_TEMPERATURE_CONFIGURATION"
	case IDENTIFY: // 0x0003
		result = "IDENTIFY"
	case GROUPS: // 0x0004
		result = "GROUPS"
	case SCENES: // 0x0005
		result = "SCENES"
	case ON_OFF: // 0x0006
		result = "ON_OFF"
	case ON_OFF_SWITCH_CONFIGURATION: // 0x0007
		result = "ON_OFF_SWITCH_CONFIGURATION"
	case LEVEL_CONTROL: // 0x0008
		result = "LEVEL_CONTROL"
	case ALARMS: //0x0009
		result = "ALARMS"
	case TIME: // 0x000a
		result = "TIME"
	case RSSI: // 0x000b
		result = "RSSI"
	case ANALOG_INPUT: //0x000c
		result = "ANALOG_INPUT"
	case ANALOG_OUTPUT: //0x000d
		result = "ANALOG_OUTPUT"
	case ANALOG_VALUE: // 0x000e
		result = "ANALOG_VALUE"
	case BINARY_INPUT: // 0x000f
		result = "BINARY_INPUT"
	case BINARY_OUTPUT: // 0x0010
		result = "BINARY_OUTPUT"
	case BINARY_VALUE: // 0x0011
		result = "BINARY_VALUE"
	case MULTISTATE_INPUT: // 0x0012
		result = "MULTISTATE_INPUT"
	case MULTISTATE_OUTPUT: // 0x0013
		result = "MULTISTATE_OUTPUT"
	case MULTISTATE_VALUE: // 0x0014
		result = "MULTISTATE_VALUE"
	case OTA: //0x0019
		result = "OTA"
	case POWER_PROFILE: // 0x001a //
		result = "POWER_PROFILE"
	case PULSE_WIDTH_MODULATION: // 0x001c
		result = "PULSE_WIDTH_MODULATION"
	case POLL_CONTROL: // 0x0020 //
		result = "POLL_CONTROL"
	case XIAOMI_SWITCH_OUTPUT: // 0x0021
		result = "XIAOMI_SWITCH_OUTPUT"
	case KEEP_ALIVE: // 0x0025
		result = "KEEP_ALIVE"
	case WINDOW_COVERING: // 0x0102
		result = "WINDOW_COVERING"
	case TEMPERATURE_MEASUREMENT: //0x0402
		result = "TEMPERATURE_MEASUREMENT"
	case HUMIDITY_MEASUREMENT: // 0x0405
		result = "HUMIDITY_MEASUREMENT"
	case ILLUMINANCE_MEASUREMENT: // 0x0400
		result = "ILLUMINANCE_MEASUREMENT"
	case IAS_ZONE: // 0x0500
		result = "IAS_ZONE"
	case SIMPLE_METERING: // 0x0702
		result = "SIMPLE_METERING"
	case METER_IDENTIFICATION: // 0x0b01
		result = "METER_IDENTIFICATION"
	case ELECTRICAL_MEASUREMENTS: //0x0b04
		result = "ELECTRICAL_MEASUREMENTS"
	case DIAGNOSTICS: // 0x0b05
		result = "DIAGNOSTICS"
	case TOUCHLINK_COMISSIONING: // 0x1000
		result = "TOUCHLINK_COMISSIONING"
	case TUYA_SWITCH_MODE_0: // 0xe000
		result = "TUYA_SWITCH_MODE_0"
	case TUYA_ELECTRICIAN_PRIVATE_CLUSTER: // 0xe001
		result = "TUYA_ELECTRICIAN_PRIVATE_CLUSTER"
	case IKEA_BUTTON_ATTR_UNKNOWN2: // 0xfc7c
		result = "IKEA_BUTTON_ATTR_UNKNOWN2"
	case XIAOMI_SWITCH: //0xfcc0
		result = "XIAOMI_SWITCH"
	case SMOKE_SENSOR: // 0xfe00
		result = "SMOKE_SENSOR"
	default:
		result = "UNKNOWN_CLUSTER"
	}
	return result
}

func init() {
	fmt.Println("Init in zcl")
}
