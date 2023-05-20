package zigbee

import "log"

// Table 2-11 ZCL Specification
type DataType byte

const (
	DataType_NODATA                DataType = 0x00
	DataType_DATA8                 DataType = 0x08
	DataType_DATA16                DataType = 0x09
	DataType_DATA24                DataType = 0x0a
	DataType_DATA32                DataType = 0x0b
	DataType_DATA40                DataType = 0x0c
	DataType_DATA48                DataType = 0x0d
	DataType_DATA56                DataType = 0x0e
	DataType_DATA64                DataType = 0x0f
	DataType_BOOLEAN               DataType = 0x10
	DataType_BITMAP8               DataType = 0x18
	DataType_BITMAP16              DataType = 0x19
	DataType_BITMAP24              DataType = 0x1a
	DataType_BITMAP32              DataType = 0x1b
	DataType_BITMAP40              DataType = 0x1c
	DataType_BITMAP48              DataType = 0x1d
	DataType_BITMAP56              DataType = 0x1e
	DataType_BITMAP64              DataType = 0x1f
	DataType_UINT8                 DataType = 0x20
	DataType_UINT16                DataType = 0x21
	DataType_UINT24                DataType = 0x22
	DataType_UINT32                DataType = 0x23
	DataType_UINT40                DataType = 0x24
	DataType_UINT48                DataType = 0x25
	DataType_UINT56                DataType = 0x26
	DataType_UINT64                DataType = 0x27
	DataType_INT8                  DataType = 0x28
	DataType_INT16                 DataType = 0x29
	DataType_INT24                 DataType = 0x2a
	DataType_INT32                 DataType = 0x2b
	DataType_INT40                 DataType = 0x2c
	DataType_INT48                 DataType = 0x2d
	DataType_INT56                 DataType = 0x2e
	DataType_INT64                 DataType = 0x2f
	DataType_ENUM8                 DataType = 0x30
	DataType_ENUM16                DataType = 0x31
	DataType_SEMI_FLOAT            DataType = 0x38
	DataType_FLOAT                 DataType = 0x39
	DataType_DOUBLE                DataType = 0x3a
	DataType_OCT_STRING            DataType = 0x41
	DataType_CHARACTER_STRING      DataType = 0x42
	DataType_LONG_OCT_STRING       DataType = 0x43
	DataType_LONG_CHARACTER_STRING DataType = 0x44
	DataType_ARRAY                 DataType = 0x48
	DataType_STRUCTURE             DataType = 0x4c
	DataType_SET                   DataType = 0x50
	DataType_BAG                   DataType = 0x51
	DataType_TIME_OF_DAY           DataType = 0xe0
	DataType_DATE                  DataType = 0xe1
	DataType_UTC_TIME              DataType = 0xe2
	DataType_CLUSTER_ID            DataType = 0xe8
	DataType_ATTRIBUTE_ID          DataType = 0xe9
	DataType_BACNET_OID            DataType = 0xea
	DataType_IEEE_ADDRESS          DataType = 0xf0
	DataType_SECURITY_KEY_128      DataType = 0xf1
	DataType_UNK                   DataType = 0xff
)

type BasicAttribute uint16

const (
	Basic_ZCL_VERSION                  BasicAttribute = 0x0000 // Type: uint8, Range: 0x00 – 0xff, Access : Read Only, Default: 0x02.
	Basic_APPLICATION_VERSION          BasicAttribute = 0x0001 // Type: uint8, Range: 0x00 – 0xff, Access : Read Only, Default: 0x00.
	Basic_STACK_VERSION                BasicAttribute = 0x0002 // Type: uint8, Range: 0x00 – 0xff, Access : Read Only, Default: 0x00.
	Basic_HW_VERSION                   BasicAttribute = 0x0003 // Type: uint8, Range: 0x00 – 0xff, Access : Read Only, Default: 0x00.
	Basic_MANUFACTURER_NAME            BasicAttribute = 0x0004 // Type: string, Range: 0 – 32 bytes, Access : Read Only, Default: Empty string.
	Basic_MODEL_IDENTIFIER             BasicAttribute = 0x0005 // Type: string, Range: 0 – 32 bytes, Access : Read Only, Default: Empty string.
	Basic_DATA_CODE                    BasicAttribute = 0x0006 // Type: string, Range: 0 – 16 bytes, Access : Read Only, Default: Empty string.
	Basic_POWER_SOURCE                 BasicAttribute = 0x0007 // Type: enum8, Range: 0x00 – 0xff, Access : Read Only, Default: 0x00.
	Basic_GENERIC_DEVICE_CLASS         BasicAttribute = 0x0008 // Type: enum8, 0x00 - 0xff, RO, 0xff
	Basic_GENERIC_DEVICE_TYPE          BasicAttribute = 0x0009 // enum8, 0x00 - 0xff, RO, 0xff
	Basic_PRODUCT_CODE                 BasicAttribute = 0x000a // octstr, RO, empty string
	Basic_PRODUCT_URL                  BasicAttribute = 0x000b // string, RO, empty string
	Basic_MANUFACTURER_VERSION_DETAILS BasicAttribute = 0x000c // string, RO, empty string
	Basic_SERIAL_NUMBER                BasicAttribute = 0x000d // string, RO, empty string
	Basic_PRODUCT_LABEL                BasicAttribute = 0x000e // string, RO, empty string
	Basic_LOCATION_DESCRIPTION         BasicAttribute = 0x0010 // Type: string, Range: 0 – 16 bytes, Access : Read Write, Default: Empty string.
	Basic_PHYSICAL_ENVIRONMENT         BasicAttribute = 0x0011 // Type: enum8, Range: 0x00 – 0xff, Access : Read Write, Default: 0x00.
	Basic_DEVICE_ENABLED               BasicAttribute = 0x0012 // Type: uint8, Range: 0x00 – 0x01, Access : Read Write, Default: 0x01.
	Basic_ALARM_MASK                   BasicAttribute = 0x0013 // Type: map8, Range: 000000xx, Access : Read Write, Default: 0x00.
	Basic_DISABLE_LOCAL_CONFIG         BasicAttribute = 0x0014 // Type: map8, Range: 000000xx, Access : Read Write, Default: 0x00.
	Basic_SW_BUILD_ID                  BasicAttribute = 0x4000 // Type: string, Range: 0 – 16 bytes, Access : Read Only, Default: Empty string.
	Basic_FF01                         BasicAttribute = 0xff01 // Type: string, Датчик протечек Xiaomi. Двухканальное реле Aqara. Строка с набором аттрибутов.
	Basic_FFE2                         BasicAttribute = 0xffe2
	Basic_FFE4                         BasicAttribute = 0xffe4
	Basic_GLOBAL_CLUSTER_REVISION      BasicAttribute = 0xfffd //
)

type PowerConfigurationAttribute uint16

const (
	PowerConfiguration_MAINS_VOLTAGE   PowerConfigurationAttribute = 0x0000 // Type: uint16, Range: 0x0000 - 0xffff, Access: ReadOnly
	PowerConfiguration_MAINS_SETTINGS  PowerConfigurationAttribute = 0x0010
	PowerConfiguration_BATTERY_VOLTAGE PowerConfigurationAttribute = 0x0020 // uint8 0x00-0xff 0,1V step
	PowerConfiguration_BATTERY_REMAIN  PowerConfigurationAttribute = 0x0021 // 0x00 - 0%  0x32 - 25% 0x64 -50%  0x96 - 75% 0xc8 -100%
	PowerConfiguration_BATTERY_SIZE    PowerConfigurationAttribute = 0x0031
)

type OnOffAttribute uint16

const (
	OnOff_ON_OFF               OnOffAttribute = 0x0000 // Type: bool, Range: 0x00 – 0x01, Access : Read Only, Default: 0x00.
	OnOff_00F5                 OnOffAttribute = 0x00f5
	OnOff_00F7                 OnOffAttribute = 0x00f7
	OnOff_GLOBAL_SCENE_CONTROL OnOffAttribute = 0x4000 // Type: bool, Range: 0x00 – 0x01, Access : Read Only, Default: 0x01.
	OnOff_ON_TIME              OnOffAttribute = 0x4001 // Type: uint16, Range: 0x0000 – 0xffff, Access : Read Write, Default: 0x0000.
	OnOff_OFF_WAIT_TIME        OnOffAttribute = 0x4002 // Type: uint16, Range: 0x0000 – 0xffff, Access : Read Write, Default: 0x0000.
	OnOff_5000                 OnOffAttribute = 0x5000
	OnOff_8000                 OnOffAttribute = 0x8000
	OnOff_8001                 OnOffAttribute = 0x8001
	OnOff_8002                 OnOffAttribute = 0x8002
	OnOff_F000                 OnOffAttribute = 0xf000
	OnOff_F500                 OnOffAttribute = 0xf500
	OnOff_F501                 OnOffAttribute = 0xf501
)

type AnalogInputAttribute uint16

const (
	AnalogInput_0055 AnalogInputAttribute = 0x0055
	AnalogInput_006f AnalogInputAttribute = 0x006f
	AnalogInput_001c AnalogInputAttribute = 0x001c
)

type MultiStateInputAttribute uint16

const (
	MultiStateInput_000E MultiStateInputAttribute = 0x000E // state text - Array of character string
	MultiStateInput_001C MultiStateInputAttribute = 0x001C // description - string
	MultiStateInput_004A MultiStateInputAttribute = 0x004A // NumberOfStates - uint16
	MultiStateInput_0051 MultiStateInputAttribute = 0x0051 // OutOfService - bool
	MultiStateInput_0055 MultiStateInputAttribute = 0x0055 // PresentValue - uint16
	MultiStateInput_0067 MultiStateInputAttribute = 0x0067 // Reliability - enum8
	MultiStateInput_006F MultiStateInputAttribute = 0x006F // StatusFlags - map8 Bit 0 = IN ALARM, Bit 1 = FAULT, Bit 2 = OVERRIDDEN, Bit 3 = OUT OF SERVICE
	MultiStateInput_0100 MultiStateInputAttribute = 0x0100 // ApplicationType
)

type XiaomiAttribute uint16

const (
	Xiaomi_0x00F7 XiaomiAttribute = 0x00f7 // string for parsing
	Xiaomi_0xFF01 XiaomiAttribute = 0xFF01 // string for parsing
)

type ElectricalMeasurementAttribute uint16

const (
	ElectricalMeasurement_0505 ElectricalMeasurementAttribute = 0x0505 // RMS Voltage V
	ElectricalMeasurement_0508 ElectricalMeasurementAttribute = 0x0508 // RMS Current mA
)

// ------------------------------------------
type Attribute struct {
	id       uint16
	value    []byte
	dataType DataType
	size     uint16
}

// return an array of bytes with array size and attribute type
func parse_attributes_payload(payload []byte, wStatus bool) []Attribute {
	var attributes []Attribute = make([]Attribute, 0)

	i := uint16(0)
	valid := true
	maxI := uint16(len(payload))

	for i < maxI {
		var attribute Attribute
		valid = true

		lo := payload[i]
		i++
		hi := payload[i]
		i++
		attribute.id = UINT16_(lo, hi)

		if wStatus {
			valid = payload[i] == byte(SUCCESS)
			i++
			if !valid {
				continue
			}
		}
		attribute.dataType = DataType(payload[i])
		i++

		if attribute.dataType == 0 || attribute.dataType == 0xff {
			return attributes
		}

		size := uint16(0) //in this var will be current attribute size

		switch attribute.dataType {

		case DataType_ARRAY, // реализация условная, теоретически могут быть вложенные объекты
			DataType_STRUCTURE,
			DataType_SET,
			DataType_BAG:
			lo := payload[i]
			i++
			hi := payload[i]
			i++
			size = UINT16_(lo, hi)
			attribute.value = payload[i : i+size]
			attribute.size = size

		case DataType_OCT_STRING,
			DataType_CHARACTER_STRING:
			size = uint16(payload[i])
			i++
			attribute.value = payload[i : i+size]

		case DataType_LONG_OCT_STRING,
			DataType_LONG_CHARACTER_STRING:
			lo := payload[i]
			i++
			hi := payload[i]
			i++
			size = UINT16_(lo, hi)
			attribute.value = payload[i : i+size]

		case DataType_BOOLEAN:
			attribute.value = []byte{payload[i]}
			size = 1 // ???

		case DataType_ENUM8,
			DataType_BITMAP8,
			DataType_DATA8,
			DataType_UINT8:
			attribute.value = []byte{payload[i]}
			size = 1

		case DataType_ENUM16,
			DataType_BITMAP16,
			DataType_DATA16,
			DataType_UINT16:
			size = 2
			attribute.value = payload[i : i+size]

		case DataType_BITMAP32,
			DataType_DATA32,
			DataType_UINT32:
			size = 4
			attribute.value = payload[i : i+size]

		case DataType_BITMAP40,
			DataType_DATA40,
			DataType_UINT40:
			size = 5
			attribute.value = payload[i : i+size]

		case DataType_BITMAP48,
			DataType_DATA48,
			DataType_UINT48:
			size = 6
			attribute.value = payload[i : i+size]

		case DataType_BITMAP56,
			DataType_DATA56,
			DataType_UINT56:
			size = 7
			attribute.value = payload[i : i+size]

		case DataType_BITMAP64,
			DataType_DATA64,
			DataType_UINT64:
			size = 8
			attribute.value = payload[i : i+size]

		case DataType_INT8:
			attribute.value = []byte{payload[i]}
			size = 1

		case DataType_INT16:
			size = 2
			attribute.value = payload[i : i+size]

		case DataType_INT24:
			size = 3
			attribute.value = payload[i : i+size]

		case DataType_INT32:
			size = 4
			attribute.value = payload[i : i+size]

		case DataType_INT40:
			size = 5
			attribute.value = payload[i : i+size]

		case DataType_INT48:
			size = 6
			attribute.value = payload[i : i+size]

		case DataType_INT56:
			size = 7
			attribute.value = payload[i : i+size]

		case DataType_INT64:
			size = 8
			attribute.value = payload[i : i+size]

			// TODO: считать по формуле из спецификации, с.101
		case DataType_SEMI_FLOAT: // half precision
			size = 2
			attribute.value = payload[i : i+size]

		case DataType_FLOAT: // single precision
			size = 4
			attribute.value = payload[i : i+size]

		case DataType_DOUBLE: // double precision
			size = 8
			attribute.value = payload[i : i+size]

		case DataType_TIME_OF_DAY: // hours minutes seconds hundredths
		case DataType_DATE: // year-1900 month day_of_month day_of_week
		case DataType_UTC_TIME: // the number of seconds since 0 hours, 0 minutes, 0 seconds, on the 1st of January, 2000 UTC.
			size = 4
			attribute.value = payload[i : i+size]

		case DataType_CLUSTER_ID:
		case DataType_ATTRIBUTE_ID:
			size = 2
			attribute.value = payload[i : i+size]

		case DataType_BACNET_OID:
			size = 4
			attribute.value = payload[i : i+size]

		case DataType_IEEE_ADDRESS:
			attribute.value = []byte{payload[i+7], payload[i+6], payload[i+5], payload[i+4], payload[i+3], payload[i+2], payload[i+1], payload[i]}
			size = 8

		case DataType_SECURITY_KEY_128:
			size = 16
			attribute.value = payload[i : i+size]

		default:
			log.Printf("Unknown attribute data type: 0x%02x for attribute 0x%04x\n", uint8(attribute.dataType), attribute.id)
			valid = false
			return attributes
		} // switch

		if valid {
			attributes = append(attributes, attribute)
		}
		i += size
		if i >= maxI {
			break
		}
	} // while

	return attributes
}
