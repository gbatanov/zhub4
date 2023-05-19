package zigbee

import "fmt"

type Command struct {
	Id      CommandId // command ID
	Payload []byte    // payload field
}

// control summ
func (c Command) Fcs() byte {
	var _fcs byte = byte(c.Payload_size()) ^ HIGHBYTE(uint16(c.Id))
	_fcs = _fcs ^ LOWBYTE(uint16(c.Id))
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

// empty command
func NewCommand(cmnd CommandId) *Command {
	cmd := Command{Id: cmnd}
	return &cmd
}

// command with payload
func New2(cmnd CommandId, payload_length uint8) *Command {
	data := make([]byte, payload_length)
	cmd := Command{Id: cmnd, Payload: data}
	return &cmd
}

// байт b будет старшим, b младшим
func UINT16_(a uint8, b uint8) uint16 {
	return uint16(b)<<8 + uint16(a)
}
func HIGHBYTE(x uint16) byte {
	return byte(x >> 8)
}
func LOWBYTE(x uint16) byte {
	return byte(x & 0x00ff)
}

func (c CommandId) String() string {
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
