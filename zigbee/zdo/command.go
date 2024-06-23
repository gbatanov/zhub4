/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2024 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
*/

package zdo

import (
	"fmt"

	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

type COMMAND_TYPE int

const (
	POLL COMMAND_TYPE = iota
	SREQ
	AREQ
	SRSP
)

type COMMAND_SUBSYSTEM int

const (
	RPC_ERROR COMMAND_SUBSYSTEM = iota
	SYS
	MAC
	NWK
	AF
	ZDO
	SAPI
	UTIL
	APP
)
const (
	APP_CNF    COMMAND_SUBSYSTEM = 15
	GREENPOWER COMMAND_SUBSYSTEM = 21
)

type Command struct {
	// идентификатор команды состоит из старшего и младшего байтов команды (CMD1 CMD0)
	// 3 старших бита в cmd1 - это тип команды, 5 младших - подсистема команд

	Id      CommandId // command ID
	Ts      int64     // timestamp
	Dir     bool      // направление команды 0-in, 1-out
	Payload []byte    // payload field
}

func (c Command) Subsystem() byte {
	return zcl.HIGHBYTE(uint16(c.Id)) & 0b00011111
}
func (c Command) Type() byte {
	return zcl.HIGHBYTE(uint16(c.Id)) >> 5
}

// control summ
func (c Command) Fcs() byte {
	var _fcs byte = byte(c.Payload_size()) ^ zcl.HIGHBYTE(uint16(c.Id))
	_fcs = _fcs ^ zcl.LOWBYTE(uint16(c.Id))
	for _, b := range c.Payload {
		_fcs = _fcs ^ b
	}
	return _fcs
}

// get payload size (size of field "Payload")
func (c Command) Payload_size() byte {
	return byte(len(c.Payload))
}
func (c *Command) Set_data(data []byte) {
	c.Payload = make([]byte, len(data))
	copy(c.Payload, data)
}

func (c *Command) String() string {
	return Command_to_string(c.Id)
}

// empty command
func NewCommand(cmnd CommandId) *Command {
	data := make([]byte, 0)
	cmd := Command{Id: cmnd, Payload: data}
	return &cmd
}

// command with payload
func New2(cmnd CommandId, payload_length uint8) *Command {
	data := make([]byte, payload_length)
	cmd := Command{Id: cmnd, Payload: data}
	return &cmd
}

func Command_to_string(c CommandId) string {
	var cmd_str string

	switch c {
	case SYS_RESET_REQ: //       0x4100,
		cmd_str = "SYS_RESET_REQ"
	case SYS_RESET_IND: //        0x4180,
		cmd_str = "SYS_RESET_IND"
	case ZDO_STARTUP_FROM_APP: //    0x2540,
		cmd_str = "ZDO_STARTUP_FROM_APP"
	case ZDO_STARTUP_FROM_APP_SRSP: //  0x6540
		cmd_str = "ZDO_STARTUP_FROM_APP_SRSP"
	case ZDO_STATE_CHANGE_IND: //     0x45c0,
		cmd_str = "ZDO_STATE_CHANGE_IND"
	case AF_REGISTER: //       0x2400,
		cmd_str = "AF_REGISTER"
	case AF_REGISTER_SRSP: //      0x6400,
		cmd_str = "AF_REGISTER_SRSP"
	case AF_INCOMING_MSG: //        0x4481,
		cmd_str = "AF_INCOMING_MSG"
	case ZDO_MGMT_PERMIT_JOIN_REQ: //  0x2536,
		cmd_str = "ZDO_MGMT_PERMIT_JOIN_REQ"
	case ZDO_MGMT_PERMIT_JOIN_SRSP: //  0x6536,
		cmd_str = "ZDO_MGMT_PERMIT_JOIN_SRSP"
	case ZDO_MGMT_PERMIT_JOIN_RSP: //  0x45b6,
		cmd_str = "ZDO_MGMT_PERMIT_JOIN_RSP"
	case ZDO_PERMIT_JOIN_IND: //     0x45cb,
		cmd_str = "ZDO_PERMIT_JOIN_IND"
	case ZDO_TC_DEV_IND: //      0x45ca,
		cmd_str = "ZDO_TC_DEV_IND"
	case ZDO_LEAVE_IND: //      0x45c9,
		cmd_str = "ZDO_LEAVE_IND"
	case ZDO_END_DEVICE_ANNCE_IND: //  0x45c1,
		cmd_str = "ZDO_END_DEVICE_ANNCE_IND"
	case ZDO_ACTIVE_EP_REQ: //     0x2505,
		cmd_str = "ZDO_ACTIVE_EP_REQ"
	case ZDO_ACTIVE_EP_SRSP: //      0x6505,
		cmd_str = "ZDO_ACTIVE_EP_SRSP"
	case ZDO_ACTIVE_EP_RSP: //      0x4585,
		cmd_str = "ZDO_ACTIVE_EP_RSP"
	case ZDO_SIMPLE_DESC_REQ: //      0x2504,
		cmd_str = "ZDO_SIMPLE_DESC_REQ"
	case ZDO_SIMPLE_DESC_SRSP: //    0x6504,
		cmd_str = "ZDO_SIMPLE_DESC_SRSP"
	case ZDO_SIMPLE_DESC_RSP: //     0x4584,
		cmd_str = "ZDO_SIMPLE_DESC_RSP"
	case ZDO_POWER_DESC_REQ: //      0x2503,
		cmd_str = "ZDO_POWER_DESC_REQ"
	case ZDO_POWER_DESC_SRSP: //    0x6503,
		cmd_str = "ZDO_POWER_DESC_SRSP"
	case ZDO_POWER_DESC_RSP: //     0x4583,
		cmd_str = "ZDO_POWER_DESC_RSP"
	case ZDO_SRC_RTG_IND: // 0x45c4
		cmd_str = "ZDO_SRC_RTG_IND"
	case SYS_PING: //      0x2101,
		cmd_str = "SYS_PING"
	case SYS_PING_SRSP: //      0x6101,
		cmd_str = "SYS_PING_SRSP"
	case SYS_OSAL_NV_READ: // 0x2108,
		cmd_str = "SYS_OSAL_NV_READ"
	case SYS_OSAL_NV_READ_SRSP: //    0x6108,
		cmd_str = "SYS_OSAL_NV_READ_SRSP"
	case SYS_OSAL_NV_WRITE: //    0x2109,
		cmd_str = "SYS_OSAL_NV_WRITE"
	case SYS_OSAL_NV_WRITE_SRSP: //   0x6109,
		cmd_str = "SYS_OSAL_NV_WRITE_SRSP"
	case SYS_OSAL_NV_ITEM_INIT: //    0x2107,
		cmd_str = "SYS_OSAL_NV_ITEM_INIT"
	case SYS_OSAL_NV_ITEM_INIT_SRSP: //  0x6107,
		cmd_str = "SYS_OSAL_NV_ITEM_INIT_SRSP"
	case SYS_OSAL_NV_LENGTH: //   0x2113:
		cmd_str = "SYS_OSAL_NV_LENGTH"
	case SYS_OSAL_NV_LENGTH_SRSP: //  0x6113,
		cmd_str = "SYS_OSAL_NV_LENGTH_SRSP"
	case SYS_OSAL_NV_DELETE: //   0x2112:
		cmd_str = "SYS_OSAL_NV_DELETE"
	case SYS_OSAL_NV_DELETE_SRSP: //  0x6112,
		cmd_str = "SYS_OSAL_NV_DELETE_SRSP"
	case UTIL_GET_DEVICE_INFO: //    0x2700,
		cmd_str = "UTIL_GET_DEVICE_INFO"
	case UTIL_GET_DEVICE_INFO_SRSP: //  0x6700,
		cmd_str = "UTIL_GET_DEVICE_INFO_SRSP"
	case SYS_SET_TX_POWER: //      0x2114,
		cmd_str = "SYS_SET_TX_POWER"
	case SYS_SET_TX_POWER_SRSP: //     0x6114,
		cmd_str = "SYS_SET_TX_POWER_SRSP"
	case SYS_VERSION: //           0x2102,
		cmd_str = "SYS_VERSION"
	case SYS_VERSION_SRSP: //       0x6102,
		cmd_str = "SYS_VERSION_SRSP"
	case AF_DATA_REQUEST: //      0x2401,
		cmd_str = "AF_DATA_REQUEST"
	case AF_DATA_REQUEST_SRSP: //     0x6401,
		cmd_str = "AF_DATA_REQUEST_SRSP"
	case AF_DATA_CONFIRM: //       0x4480,
		cmd_str = "AF_DATA_CONFIRM"
	case ZDO_BIND_REQ: //      0x2521,
		cmd_str = "ZDO_BIND_REQ"
	case ZDO_UNBIND_REQ: // 0x2522,
		cmd_str = "ZDO_UNBIND_REQ"
	case ZDO_BIND_RSP: //  0x45a1,
		cmd_str = "ZDO_BIND_RSP"
	case ZDO_BIND_SRSP: // 0x6521
		cmd_str = "ZDO_BIND_SRSP"
	case ZDO_UNBIND_RSP: //         0x45a2
		cmd_str = "ZDO_UNBIND_RSP"
	case ZDO_IEEE_ADDR_REQ: // 0x2501,
		cmd_str = "ZDO_IEEE_ADDR_REQ"
	case ZDO_IEEE_ADDR_REQ_SRSP: // 0x6501,
		cmd_str = "ZDO_IEEE_ADDR_REQ_SRSP"
	default:
		cmd_str = fmt.Sprintf("Unknown command 0x%04x \n", uint16(c))
	}

	return cmd_str
}

// commands
type CommandId uint16

const (
	//System  (команды синхронные)
	SYS_RESET_REQ              CommandId = 0x4100
	SYS_RESET_IND              CommandId = 0x4180
	SYS_PING                   CommandId = 0x2101
	SYS_PING_SRSP              CommandId = 0x6101
	SYS_OSAL_NV_READ           CommandId = 0x2108
	SYS_OSAL_NV_READ_SRSP      CommandId = 0x6108
	SYS_OSAL_NV_WRITE          CommandId = 0x2109
	SYS_OSAL_NV_WRITE_SRSP     CommandId = 0x6109
	SYS_OSAL_NV_ITEM_INIT      CommandId = 0x2107
	SYS_OSAL_NV_ITEM_INIT_SRSP CommandId = 0x6107
	SYS_OSAL_NV_LENGTH         CommandId = 0x2113
	SYS_OSAL_NV_LENGTH_SRSP    CommandId = 0x6113
	SYS_OSAL_NV_DELETE         CommandId = 0x2112
	SYS_OSAL_NV_DELETE_SRSP    CommandId = 0x6112
	SYS_SET_TX_POWER           CommandId = 0x2114
	SYS_SET_TX_POWER_SRSP      CommandId = 0x6114
	SYS_VERSION                CommandId = 0x2102
	SYS_VERSION_SRSP           CommandId = 0x6102

	// ZDO (команды исходящие синхронные)
	ZDO_STARTUP_FROM_APP      CommandId = 0x2540
	ZDO_STARTUP_FROM_APP_SRSP CommandId = 0x6540
	ZDO_BIND_REQ              CommandId = 0x2521
	ZDO_BIND_RSP              CommandId = 0x45a1
	ZDO_BIND_SRSP             CommandId = 0x6521
	ZDO_UNBIND_REQ            CommandId = 0x2522
	ZDO_UNBIND_RSP            CommandId = 0x45a2
	ZDO_MGMT_LQI_REQ          CommandId = 0x2531
	ZDO_MGMT_LQI_SRSP         CommandId = 0x6531
	ZDO_MGMT_LQI_RSP          CommandId = 0x45b1
	ZDO_SRC_RTG_IND           CommandId = 0x45C4
	ZDO_MGMT_PERMIT_JOIN_REQ  CommandId = 0x2536
	ZDO_MGMT_PERMIT_JOIN_SRSP CommandId = 0x6536
	ZDO_MGMT_PERMIT_JOIN_RSP  CommandId = 0x45b6
	ZDO_PERMIT_JOIN_IND       CommandId = 0x45cb
	ZDO_TC_DEV_IND            CommandId = 0x45ca
	ZDO_LEAVE_IND             CommandId = 0x45c9
	ZDO_END_DEVICE_ANNCE_IND  CommandId = 0x45c1
	ZDO_ACTIVE_EP_REQ         CommandId = 0x2505
	ZDO_ACTIVE_EP_SRSP        CommandId = 0x6505
	ZDO_ACTIVE_EP_RSP         CommandId = 0x4585
	ZDO_SIMPLE_DESC_REQ       CommandId = 0x2504
	ZDO_SIMPLE_DESC_SRSP      CommandId = 0x6504
	ZDO_SIMPLE_DESC_RSP       CommandId = 0x4584
	ZDO_POWER_DESC_REQ        CommandId = 0x2503
	ZDO_POWER_DESC_SRSP       CommandId = 0x6503
	ZDO_POWER_DESC_RSP        CommandId = 0x4583
	ZDO_IEEE_ADDR_REQ         CommandId = 0x2501
	ZDO_IEEE_ADDR_REQ_SRSP    CommandId = 0x6501
	ZDO_IEEE_ADDR_RSP         CommandId = 0x4581
	ZDO_STATE_CHANGE_IND      CommandId = 0x45c0

	// AF
	AF_REGISTER          CommandId = 0x2400
	AF_REGISTER_SRSP     CommandId = 0x6400
	AF_INCOMING_MSG      CommandId = 0x4481
	AF_DATA_REQUEST      CommandId = 0x2401
	AF_DATA_REQUEST_SRSP CommandId = 0x6401
	AF_DATA_CONFIRM      CommandId = 0x4480

	// UTIL
	UTIL_GET_DEVICE_INFO      CommandId = 0x2700
	UTIL_GET_DEVICE_INFO_SRSP CommandId = 0x6700

	// ZB
	ZB_GET_DEVICE_INFO      CommandId = 0x2606
	ZB_GET_DEVICE_INFO_SRSP CommandId = 0x6606

	// APP
	APP_CNF_BDB_COMMISSIONING_NOTIFICATION CommandId = 0x4F80
)
