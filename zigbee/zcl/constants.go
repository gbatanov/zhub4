/*
GSB, 2023
gbatanov@yandex.ru
*/
package zcl

type StartupStatus byte

const (
	RESTORED_STATE StartupStatus = 0 //  Restored network state
	NEW_STATE      StartupStatus = 1 //  New network state
	NOT_STARTED    StartupStatus = 2 // Leave and not Starte
)

type LogicalType byte

const (
	LogicalType_COORDINATOR LogicalType = 0
	LogicalType_ROUTER      LogicalType = 1
	LogicalType_END_DEVICE  LogicalType = 2
)

// Reset type
type ResetType byte

const (
	RESET_TYPE_HARD ResetType = 0
	RESET_TYPE_SOFT ResetType = 1
)

// Statuses
type Statuses uint8

const (
	SUCCESS                   Statuses = 0x00
	FAILURE                   Statuses = 0x01
	INVALID_PARAMETR          Statuses = 0x02
	NV_ITEM_UNINIT            Statuses = 0x09
	NV_OPER_FAILED            Statuses = 0x0a
	NV_BAD_ITEM_LENGTH        Statuses = 0x0c
	MEMORY_ERROR              Statuses = 0x10
	BUFFER_FULL               Statuses = 0x11
	UNSUPPORTED_MODE          Statuses = 0x12
	MAC_MEMMORY_ERROR         Statuses = 0x13
	ZDO_INVALID_REQUEST_TYPE  Statuses = 0x80
	ZDO_INVALID_ENDPOINT      Statuses = 0x82
	ZDO_UNSUPPORTED           Statuses = 0x84
	ZDO_TIMEOUT               Statuses = 0x85
	ZDO_NO_MUTCH              Statuses = 0x86
	ZDO_TABLE_FULL            Statuses = 0x87
	ZDO_NO_BIND_ENTRY         Statuses = 0x88
	SEC_NO_KEY                Statuses = 0xa1
	SEC_MAX_FRM_COUNT         Statuses = 0xa3
	APS_FAIL                  Statuses = 0xb1
	APS_TABLE_FULL            Statuses = 0xb2
	APS_ILLEGAL_REQEST        Statuses = 0xb3
	APS_INVALID_BINDING       Statuses = 0xb4
	APS_UNSUPPORTED_ATTRIBUTE Statuses = 0xb5
	APS_NOT_SUPPORTED         Statuses = 0xb6
	APS_NO_ACK                Statuses = 0xb7
	APS_DUPLICATE_ENTRY       Statuses = 0xb8
	APS_NO_BOUND_DEVICE       Statuses = 0xb9
	NWK_INVALID_PARAM         Statuses = 0xc1
	NWK_INVALID_REQUEST       Statuses = 0xc2
	NWK_NOT_PERMITTED         Statuses = 0xc3
	NWK_STARTUP_FAILURE       Statuses = 0xc4
	NWK_TABLE_FULL            Statuses = 0xc7
	NWK_UNKNOWN_DEVICE        Statuses = 0xc8
	NWK_UNSUPPORTED_ATTRIBUTE Statuses = 0xc9
	NWK_NO_NETWORK            Statuses = 0xca
	NWK_LEAVE_UNCONFIRMED     Statuses = 0xcb
	NWK_NO_ACK                Statuses = 0xcc
	NWK_NO_ROUTE              Statuses = 0xcd
	MAC_NO_ACK                Statuses = 0xe9
	MAC_TRANSACTION_EXPIRED   Statuses = 0xf0
)

// NvItems
type NvItems uint16

const (
	STARTUP_OPTION     NvItems = 0x0003 // 1 : default - 0, bit 0 - STARTOPT_CLEAR_CONFIG, bit 1 - STARTOPT_CLEAR_STATE (need reset).
	PAN_ID             NvItems = 0x0083 // 2 : 0xFFFF to indicate dont care (auto).
	EXTENDED_PAN_ID    NvItems = 0x002D // 8: (0xDD, 0xDD, 0xDD, 0xDD, 0xDD, 0xDD, 0xDD, 0xDD) - from zigbee2mqtt
	CHANNEL_LIST       NvItems = 0x0084 // 4 : channel bit mask. Little endian. Default is 0x00000800 for CH11;  Ex: value: [ 0x00, 0x00, 0x00, 0x04 ] for CH26, [ 0x00, 0x00, 0x20, 0x00 ] for CH15.
	LOGICAL_TYPE       NvItems = 0x0087 // 1 : COORDINATOR: 0, ROUTER : 1, END_DEVICE : 2
	PRECFG_KEY         NvItems = 0x0062 // 16 : (0x01, 0x03, 0x05, 0x07, 0x09, 0x0B, 0x0D, 0x0F, 0x00, 0x02, 0x04, 0x06, 0x08, 0x0A, 0x0C, 0x0D) - from zigbee2mqtt
	PRECFG_KEYS_ENABLE NvItems = 0x0063 // 1 : defalt - 1
	ZDO_DIRECT_CB      NvItems = 0x008F // 1 : defaul - 0, need for callbacks - 1
	ZNP_HAS_CONFIGURED NvItems = 0x0F00 // 1 : 0x55
	MY_DEVICES_MAP     NvItems = 0x0FF3 // 150 : для хранения своих данных ( не более 150 байт)
)

type PowerSource uint8

const (
	PowerSource_UNKNOWN                                  PowerSource = 0x00
	PowerSource_SINGLE_PHASE                             PowerSource = 0x01
	PowerSource_THREE_PHASE                              PowerSource = 0x02
	PowerSource_BATTERY                                  PowerSource = 0x03
	PowerSource_DC                                       PowerSource = 0x04
	PowerSource_EMERGENCY_CONSTANTLY                     PowerSource = 0x05
	PowerSource_EMERGENCY_MAINS_AND_TRANSFER_SWITCH      PowerSource = 0x06
	PowerSource_UNKNOWN_PLUS                             PowerSource = 0x80
	PowerSource_SINGLE_PHASE_PLUS                        PowerSource = 0x81
	PowerSource_THREE_PHASE_PLUS                         PowerSource = 0x82
	PowerSource_BATTERY_PLUS                             PowerSource = 0x83
	PowerSource_DC_PLUS                                  PowerSource = 0x84
	PowerSource_EMERGENCY_CONSTANTLY_PLUS                PowerSource = 0x85
	PowerSource_EMERGENCY_MAINS_AND_TRANSFER_SWITCH_PLUS PowerSource = 0x86
)
