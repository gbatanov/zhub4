/*
GSB, 2023
gbatanov@yandex.ru
*/
package zdo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
	"zhub4/serial3"
	"zhub4/zigbee/zdo/zcl"
)

type SimpleDescriptor struct {
	endpointNumber uint16
	profileId      uint16
	deviceId       uint16
	deviceVersion  uint16
	inputClusters  []uint16
	outputClusters []uint16
}

var Default_endpoint SimpleDescriptor = SimpleDescriptor{1, // Enpoint number.
	0x0104,     // Profile ID.
	0x05,       // Device ID.
	0,          // Device version.
	[]uint16{}, // Input clusters list.
	[]uint16{}} // Output clusters list.

type Message struct {
	Source      zcl.Endpoint
	Destination zcl.Endpoint
	Cluster     zcl.Cluster
	ZclFrame    zcl.Frame
	LinkQuality uint8
}

type Zdo struct {
	eh                        EventHandler
	Flag                      bool
	Uart                      *serial3.Uart
	Cmdinput                  chan []byte
	transactionNumber         uint8
	transactionSecuenseNumber uint8
	tsnMutex                  sync.Mutex
	macAddress                uint64
	ShortAddress              uint16
	isReady                   bool
	os                        string
	msgChan                   chan Command // chanel for send command to controller
	joinChan                  chan []byte  // chanel for send command join device to controller
	tmpBuff                   []byte
}

func init() {
	fmt.Println("Init in zigbee: zdo")
}

func ZdoCreate(port string, Os string, chn chan Command, jchn chan []byte) (*Zdo, error) {
	eh := CreateEventHandler()
	uart := serial3.UartCreate(port, Os)
	cmdinput := make(chan []byte, 256)
	err := uart.Open()
	if err != nil {
		return &Zdo{}, err
	}
	zdo := Zdo{eh: *eh,
		Flag:                      true,
		Uart:                      uart,
		Cmdinput:                  cmdinput,
		transactionNumber:         0,
		transactionSecuenseNumber: 0,
		tsnMutex:                  sync.Mutex{},
		macAddress:                0x0000000000000000,
		ShortAddress:              0x0000,
		isReady:                   false,
		os:                        Os,
		msgChan:                   chn,
		joinChan:                  jchn,
		tmpBuff:                   []byte{}}

	return &zdo, nil
}

func (zdo *Zdo) Stop() {
	zdo.Uart.Stop()
	zdo.Flag = false
	zdo.Cmdinput <- []byte{0}
}

// call sync request
func (zdo *Zdo) sync_request(request Command, timeout time.Duration) Command {

	var id CommandId = CommandId((uint16(request.Id) | 0b0100000000000000)) // идентификатор синхронного ответа
	zdo.eh.AddEvent(id)
	buff := zdo.prepare_command(request)
	err := zdo.Uart.Send_command_to_device(buff)
	if err != nil {
		return *NewCommand(0)
	}
	cmd := zdo.eh.wait(id, timeout)

	return cmd
}

// async request. Answer will get in command handler
func (zdo *Zdo) async_request(request Command, timeout time.Duration) error {
	response := zdo.sync_request(request, timeout)
	if uint16(response.Id) != 0 && response.Payload[0] == byte(zcl.SUCCESS) {
		return nil
	} else {
		return errors.New("async request error")
	}
}

func (zdo *Zdo) prepare_command(command Command) []byte {
	buff := make([]byte, command.Payload_size()+5)
	buff[0] = serial3.SOF
	buff[1] = command.Payload_size()
	buff[2] = zcl.HIGHBYTE(uint16(command.Id))
	buff[3] = zcl.LOWBYTE(uint16(command.Id))
	for i := 0; i < int(command.Payload_size()); i++ {
		buff[4+i] = command.Payload[i]
	}
	buff[command.Payload_size()+4] = command.Fcs()
	//	 log.Printf("prepare command buff %02x ", buff)
	return buff
}

// receive command from UART and call command handler
func (zdo *Zdo) Input_command() {

	for zdo.Flag {
		command_src := <-zdo.Cmdinput
		if zdo.Flag {
			if len(zdo.tmpBuff) > 0 {
				command_src = append(zdo.tmpBuff, command_src...)
			}
			commands, next := zdo.parse_command(command_src)
			if next {
				zdo.tmpBuff = append(zdo.tmpBuff, command_src...)
			} else {
				zdo.tmpBuff = []byte{}
				for _, command := range commands {
					go func(cmd Command) { zdo.handle_command(cmd) }(command)
				}
			}
		}
	}
	zdo.Uart.Flag = false
}

// len(BufRead) >= 5!!! SOF Length Cmd0 Cmd1 FCS
func (zdo *Zdo) parse_command(BufRead []byte) ([]Command, bool) {

	if len(BufRead) < 5 {
		return []Command{}, true
	}
	var result []Command = make([]Command, 0)

	if false {
		fmt.Printf("parse_command:: BufRead: len = %d , data: ", len(BufRead))
		for i := 0; i < len(BufRead); i++ {
			fmt.Printf("0x%02x ", BufRead[i])
		}
		fmt.Println("")
	}

	for i := 0; i < len(BufRead); i++ {
		b := BufRead[i]
		if b == serial3.SOF {
			i++
			payload_length := BufRead[i]
			if payload_length > byte(len(BufRead)-5) {
				return []Command{}, true
			}

			i++
			cmd0 := BufRead[i]
			i++
			cmd1 := BufRead[i]
			i++
			// в команде сначала идет старший байт Cmd0, за ним младший Cmd1
			var cmd CommandId = CommandId(zcl.UINT16_(cmd1, cmd0))
			var command *Command = NewCommand(cmd)

			for j := 0; j < int(payload_length); j++ {
				command.Payload = append(command.Payload, BufRead[i])
				i++
			}
			//			fmt.Println("")

			if BufRead[i] == command.Fcs() {
				result = append(result, *command)
			}
		} //if
	} //for
	return result, false
}

// reset zigbee-adapter
func (zdo *Zdo) Reset() error {
	var cmd Command = *NewCommand(0)
	//	zdo.eh.clear(SYS_RESET_IND)
	// writeNv call sync request
	startup_options := make([]byte, 1)
	startup_options[0] = 0
	err := zdo.writeNv(zcl.STARTUP_OPTION, startup_options) // STARTUP_OPTION = 0x0003
	if err != nil {
		return err
	}
	//	log.Println("WriteNv success")

	reset_request := New2(SYS_RESET_REQ, 1)
	reset_request.Payload[0] = byte(zcl.RESET_TYPE_SOFT)

	buff := zdo.prepare_command(*reset_request)
	if false {
		fmt.Print("Reset buff: ")
		for i := 0; i < len(buff); i++ {
			fmt.Printf(" 0x%02x", buff[i])
		}
		fmt.Println("")
	}
	err = zdo.Uart.Send_command_to_device(buff)
	if err != nil {
		return err
	}
	cmd = zdo.eh.wait(SYS_RESET_IND, 10*time.Second)
	if cmd.Id == 0 {
		return errors.New("bad reset")
	}

	if cmd.Payload_size() > 5 {
		log.Printf("Version: %d.%d.%d \n", cmd.Payload[3], cmd.Payload[4], cmd.Payload[5])
	} else {
		log.Printf("reset answer: %q \n", cmd)
	}
	return nil
}

// write into NVRAM of zhub
func (zdo *Zdo) writeNv(item zcl.NvItems, item_data []byte) error {
	write_nv_request := New2(SYS_OSAL_NV_WRITE, uint8(len(item_data)+4))
	data := make([]byte, len(item_data)+4)
	data[0] = zcl.LOWBYTE(uint16(item))
	data[1] = zcl.HIGHBYTE(uint16(item))
	data[2] = 0
	data[3] = uint8(len(item_data))
	for i, e := range item_data {
		data[i+4] = e
	}
	write_nv_request.Set_data(data)
	response := zdo.sync_request(*write_nv_request, 30*time.Second)
	//	 log.Printf("response- %q \n", response)
	if response.Id != 0x0000 && response.Payload[0] == byte(zcl.SUCCESS) {
		return nil
	} else {
		return errors.New("writeNv: bad answer")
	}
}

// read from zhub NVRAM
func (zdo *Zdo) readNv(item zcl.NvItems) []byte {
	//	log.Println("read from NVRAM")
	read_nv_request := New2(SYS_OSAL_NV_READ, 3)
	read_nv_request.Payload[0] = zcl.LOWBYTE(uint16(item))
	read_nv_request.Payload[1] = zcl.HIGHBYTE(uint16(item))
	read_nv_request.Payload[2] = 0 // Number of bytes offset from the beginning or the NV value.

	read_nv_response := zdo.sync_request(*read_nv_request, 10*time.Second)
	if read_nv_response.Id != 0x0000 && read_nv_response.Payload[0] == byte(zcl.SUCCESS) {
		return read_nv_response.Payload[2:]
	} else {
		items := make([]byte, 0)
		return items
	}

}

// read network configuration from zhub
func (zdo *Zdo) ReadNetworkConfiguration() (NetworkConfiguration, error) {
	nc := NetworkConfiguration{
		pan_id:            0,
		extended_pan_id:   0,
		logical_type:      zcl.LogicalType_COORDINATOR,
		channels:          []uint8{},
		precfg_key:        [16]uint8{},
		precfg_key_enable: false,
	}
	item_data := zdo.readNv(zcl.PAN_ID) //PAN_ID = 0x0083, идентификатор сети
	if len(item_data) == 2 {
		nc.pan_id = zcl.UINT16_(item_data[0], item_data[1])
		log.Printf("nc.pan_id: 0x%04x\n", nc.pan_id)
	}
	item_data = zdo.readNv(zcl.EXTENDED_PAN_ID) // EXTENDED_PAN_ID = 0x002D MAC address
	if len(item_data) == 8 {
		nc.extended_pan_id = binary.LittleEndian.Uint64(item_data)
		log.Printf("nc.extended_pan_id: 0x%016x\n", nc.extended_pan_id)
	}
	item_data = zdo.readNv(zcl.LOGICAL_TYPE) // LOGICAL_TYPE = 0x0087 zhub|router|endpoint
	if len(item_data) == 1 {
		nc.logical_type = zcl.LogicalType(item_data[0])
		log.Printf("nc.logical_type: 0x%02x\n", nc.logical_type)
	}
	item_data = zdo.readNv(zcl.PRECFG_KEYS_ENABLE) //PRECFG_KEYS_ENABLE = 0x0063
	if len(item_data) == 1 {
		nc.precfg_key_enable = (item_data[0] == 1)
		log.Printf("nc.precfg_key_enable: %d\n", item_data[0])
	}
	if nc.precfg_key_enable {
		item_data = zdo.readNv(zcl.PRECFG_KEY) // PRECFG_KEY = 0x0062
		if len(item_data) == 16 {
			log.Println("nc.precfg_key:")
			for i := 0; i < 16; i++ {
				nc.precfg_key[i] = item_data[i]
				log.Printf("0x%02x ", nc.precfg_key[i])
			}
			log.Println("")
		}
	}
	item_data = zdo.readNv(zcl.CHANNEL_LIST) // CHANNEL_LIST = 0x00000084 //channel bit mask. Little endian. Default is 0x00000800 for CH11;  Ex: value: [ 0x00, 0x00, 0x00, 0x04 ] for CH26, [ 0x00, 0x00, 0x20, 0x00 ] for CH15.
	if len(item_data) == 4 {                 //CHANNEL_LIST = 0x00000800 CH11 0x00008000 CH15
		var channelBitMask uint32 = binary.LittleEndian.Uint32(item_data)
		log.Printf("nc.channels bitMask: 0x%08x \n", channelBitMask)
		for i := 0; i < 32; i++ {
			if (channelBitMask & uint32(1<<i)) != 0 {
				nc.channels = append(nc.channels, uint8(i))
				log.Printf("channel %d\n", i)
			}
		}
	}
	return nc, nil
}

// write new configuration into zhub
func (zdo *Zdo) WriteNetworkConfiguration(configuration NetworkConfiguration) error {

	err := zdo.writeNv(zcl.LOGICAL_TYPE, []byte{byte(configuration.logical_type)})
	if err != nil {
		return err
	}
	channelBitMask := uint32(0)
	for _, channel := range configuration.channels {
		channelBitMask |= (1 << channel)
	}
	//	log.Printf("write bitMask: 0x%08x \n", channelBitMask)

	chann := []byte{0, 0, 0, 0}
	for i := 0; i < 4; i++ {
		ch := byte(channelBitMask >> (i * 8))
		chann[i] = ch
	}
	//	log.Printf("write channels: %q\n", chann)

	err = zdo.writeNv(zcl.CHANNEL_LIST, chann) // старший байт последний
	if err != nil {
		return err
	}

	var v byte = 0
	if configuration.precfg_key_enable {
		v = 1

		err = zdo.writeNv(zcl.PRECFG_KEYS_ENABLE, []byte{v})
		if err != nil {
			return err
		}
		var temp_v []byte = make([]byte, 16)
		for i := 0; i < 16; i++ {
			temp_v[i] = configuration.precfg_key[i]
		}
		err = zdo.writeNv(zcl.PRECFG_KEY, temp_v)
		if err != nil {
			return err
		}
	}
	err = zdo.writeNv(zcl.ZDO_DIRECT_CB, []byte{1})
	if err != nil {
		return err
	}
	err = zdo.initNv(zcl.ZNP_HAS_CONFIGURED, 1, []byte{0})
	if err != nil {
		return err
	}
	err = zdo.writeNv(zcl.ZNP_HAS_CONFIGURED, []byte{0x55})
	if err != nil {
		return err
	}
	return nil
}

// init element in zhub NVRAM
func (zdo *Zdo) initNv(item zcl.NvItems, length uint16, item_data []byte) error {

	init_nv_request := New2(SYS_OSAL_NV_ITEM_INIT, uint8(len(item_data)+5))
	init_nv_request.Payload[0] = zcl.LOWBYTE(uint16(item))    // The Id of the NV item.
	init_nv_request.Payload[1] = zcl.HIGHBYTE(uint16(item))   //
	init_nv_request.Payload[2] = zcl.LOWBYTE(uint16(length))  // Number of bytes in the NV item.
	init_nv_request.Payload[3] = zcl.HIGHBYTE(uint16(length)) //
	init_nv_request.Payload[4] = byte(len(item_data))         // Number of bytes in the initialization data.
	for i := 0; i < len(item_data); i++ {
		init_nv_request.Payload[5+i] = item_data[i]
	}
	init_nv_response := zdo.sync_request(*init_nv_request, 3*time.Second)
	// 0x00 = Item already exists, no action taken
	// 0x09 = Success, item created and initialized
	// 0x0A = Initialization failed, item not created
	if init_nv_response.Id != 0 && init_nv_response.Payload[0] != 0x0a {
		return nil
	} else {
		return errors.New("init NVRAM error")
	}
}

func (zdo *Zdo) Startup(delay time.Duration) error {

	startup_request := New2(ZDO_STARTUP_FROM_APP, 2)
	startup_request.Payload[0] = zcl.LOWBYTE(uint16(delay))
	startup_request.Payload[1] = zcl.HIGHBYTE(uint16(delay))

	startup_response := zdo.sync_request(*startup_request, 3*time.Second)

	if startup_response.Id == 0 || startup_response.Payload[0] == byte(zcl.NOT_STARTED) {
		log.Println("startup error 1")
		return errors.New("startup error 1")
	}
	device_info_response := zdo.sync_request(*NewCommand(UTIL_GET_DEVICE_INFO), 3*time.Second)
	if device_info_response.Id != 0 && device_info_response.Payload[0] == byte(zcl.SUCCESS) {

		zdo.macAddress = binary.LittleEndian.Uint64(device_info_response.Payload[1:10])
		log.Printf("Configurator info: IEEE address: 0x%016x \n", zdo.macAddress)
		zdo.ShortAddress = zcl.UINT16_(device_info_response.Payload[9], device_info_response.Payload[10])
		log.Printf("Configurator info: shortAddr: 0x%04x \n", zdo.ShortAddress)
		log.Printf("Configurator info: Device Type: 0x%02x \n", device_info_response.Payload[11])
		log.Printf("Configurator info: Device State: 0x%02x \n", device_info_response.Payload[12])
		fmt.Printf("Configurator info: Associated devices count: %d \n", device_info_response.Payload[13])
		if device_info_response.Payload[13] > 0 {
			for i := 0; i < int(device_info_response.Payload[13]); i++ {
				fmt.Printf("0x%04x ", zcl.UINT16_(device_info_response.Payload[i+14], device_info_response.Payload[i+15]))
			}
			fmt.Println("")
		}
	} else {
		return errors.New("startup error 2")
	}
	return nil
}
func (zdo *Zdo) RegisterEndpointDescriptor(endpoint_descriptor SimpleDescriptor) error {

	register_ep_request := New2(AF_REGISTER, 9)
	register_ep_request.Payload[0] = 1 //uint8(endpoint_descriptor.endpoint_number)
	register_ep_request.Payload[1] = zcl.LOWBYTE(endpoint_descriptor.profileId)
	register_ep_request.Payload[2] = zcl.HIGHBYTE(endpoint_descriptor.profileId)
	register_ep_request.Payload[3] = zcl.LOWBYTE(endpoint_descriptor.deviceId)
	register_ep_request.Payload[4] = zcl.HIGHBYTE(endpoint_descriptor.deviceId)
	register_ep_request.Payload[5] = uint8(endpoint_descriptor.deviceVersion)
	register_ep_request.Payload[6] = 0 // 0x00 - No latency*, 0x01 - fast beacons, 0x02 - slow beacons.
	register_ep_request.Payload[7] = 0 // input cluster size
	register_ep_request.Payload[8] = 0 // output cluster size

	response := zdo.sync_request(*register_ep_request, 10*time.Second)
	if response.Id != 0 && response.Payload[0] == byte(zcl.SUCCESS) {
		return nil
	} else {
		return errors.New("register endpoint error")
	}
}

// Enable pairing mode for duration seconds
func (zdo *Zdo) Permit_join(duration time.Duration) error {

	permitJoinRequest := New2(ZDO_MGMT_PERMIT_JOIN_REQ, 5)
	permitJoinRequest.Payload[0] = 0x0F     // Destination address type : 0x02 - Address 16 bit, 0x0F - Broadcast.
	permitJoinRequest.Payload[1] = 0xFC     // Specifies the network address of the destination device whose Permit Join information is to be modified.
	permitJoinRequest.Payload[2] = 0xFF     // (address || 0xFFFC)
	permitJoinRequest.Payload[3] = byte(60) //  duration.
	permitJoinRequest.Payload[4] = 0x00     // Trust Center Significance (0).

	return zdo.async_request(*permitJoinRequest, 3*time.Second)
}

func (zdo *Zdo) Parse_zcl_data(data []byte) zcl.Frame {
	var zclFrame zcl.Frame

	zclFrame.Frame_control.Ftype = zcl.FrameType(data[0] & 0b00000011)
	zclFrame.Frame_control.ManufacturerSpecific = (data[0] & 0b00000100) >> 2
	zclFrame.Frame_control.Direction = zcl.FrameDirection(data[0] & 0b00001000)
	zclFrame.Frame_control.DisableDefaultResponse = (data[0] & 0b00010000) >> 4

	var i uint8 = 1
	if zclFrame.Frame_control.ManufacturerSpecific == 1 {
		zclFrame.ManufacturerCode = zcl.UINT16_(data[1], data[2])
		i = 3
	} else {
		zclFrame.ManufacturerCode = 0xffff
	}
	zclFrame.TransactionSequenceNumber = data[i]
	i++
	zclFrame.Command = data[i]
	i++
	zclFrame.Payload = make([]byte, len(data[i:]))
	copy(zclFrame.Payload, data[i:])

	return zclFrame
}

func (zdo *Zdo) Send_message(ep zcl.Endpoint, cl zcl.Cluster, frame zcl.Frame, timeout time.Duration) error {
	var message Message = Message{}
	message.Cluster = cl
	message.Source = zcl.Endpoint{Address: 0x0000, Number: 1}
	message.Destination = ep
	message.ZclFrame = frame
	message.LinkQuality = 0
	transactionNumber := zdo.GenerateTransactionNumber()

	afDataRequest := New2(AF_DATA_REQUEST, 255)
	afDataRequest.Payload[0] = zcl.LOWBYTE(message.Destination.Address)
	afDataRequest.Payload[1] = zcl.HIGHBYTE(message.Destination.Address)
	afDataRequest.Payload[2] = message.Destination.Number
	afDataRequest.Payload[3] = message.Source.Number
	afDataRequest.Payload[4] = zcl.LOWBYTE(uint16(message.Cluster))
	afDataRequest.Payload[5] = zcl.HIGHBYTE(uint16(message.Cluster))
	afDataRequest.Payload[6] = transactionNumber
	afDataRequest.Payload[7] = 0
	afDataRequest.Payload[8] = 7 // DEFAULT_RADIUS
	afDataRequest.Payload[10] = byte(message.ZclFrame.Frame_control.Ftype&0b00000011) +
		byte(message.ZclFrame.Frame_control.ManufacturerSpecific)<<2 +
		byte(message.ZclFrame.Frame_control.Direction)<<3 +
		message.ZclFrame.Frame_control.DisableDefaultResponse<<4

	var i uint8 = 11
	if message.ZclFrame.Frame_control.ManufacturerSpecific == 1 {
		afDataRequest.Payload[i] = zcl.LOWBYTE(message.ZclFrame.ManufacturerCode)
		i++
		afDataRequest.Payload[i] = zcl.HIGHBYTE(message.ZclFrame.ManufacturerCode)
		i++
	}
	afDataRequest.Payload[i] = message.ZclFrame.TransactionSequenceNumber
	i++
	afDataRequest.Payload[i] = message.ZclFrame.Command
	i++

	for n := 0; n < len(message.ZclFrame.Payload); n++ {
		afDataRequest.Payload[i] = message.ZclFrame.Payload[n]
		i++
	}
	afDataRequest.Payload[9] = i                      // data length
	afDataRequest.Payload = afDataRequest.Payload[:i] // cut superfluous

	return zdo.async_request(*afDataRequest, 3*time.Second)
}

// get endpoint list from device
func (zdo *Zdo) ActiveEndpoints(address uint16) error {
	activeEndpointsRequest := New2(ZDO_ACTIVE_EP_REQ, 4)
	activeEndpointsRequest.Payload[0] = zcl.LOWBYTE(address)
	activeEndpointsRequest.Payload[1] = zcl.HIGHBYTE(address)
	activeEndpointsRequest.Payload[2] = zcl.LOWBYTE(address)
	activeEndpointsRequest.Payload[3] = zcl.HIGHBYTE(address)
	return zdo.async_request(*activeEndpointsRequest, 3*time.Second)
}

// get endpoint descriptor from device
func (zdo *Zdo) SimpleDescriptor(address uint16, endpointNumber uint8) error {
	activeEndpointsRequest := New2(ZDO_SIMPLE_DESC_REQ, 5)
	activeEndpointsRequest.Payload[0] = zcl.LOWBYTE(address)
	activeEndpointsRequest.Payload[1] = zcl.HIGHBYTE(address)
	activeEndpointsRequest.Payload[2] = zcl.LOWBYTE(address)
	activeEndpointsRequest.Payload[3] = zcl.HIGHBYTE(address)
	activeEndpointsRequest.Payload[4] = endpointNumber
	return zdo.async_request(*activeEndpointsRequest, 3*time.Second)

}

// bind device with zhub
func (zdo *Zdo) Bind(shortAddress uint16, macAddress uint64, endpoint uint8, cluster zcl.Cluster) error {
	bindRequest := NewCommand(ZDO_BIND_REQ)
	bindRequest.Payload = []byte{}
	bindRequest.Payload = append(bindRequest.Payload, zcl.LOWBYTE(shortAddress))
	bindRequest.Payload = append(bindRequest.Payload, zcl.HIGHBYTE(shortAddress))
	var b uint8 = 0
	for j := 0; j < 8; j++ {
		b = uint8(macAddress >> uint64(8*j))
		bindRequest.Payload = append(bindRequest.Payload, b)
	}
	bindRequest.Payload = append(bindRequest.Payload, endpoint)
	bindRequest.Payload = append(bindRequest.Payload, zcl.LOWBYTE(uint16(cluster)))
	bindRequest.Payload = append(bindRequest.Payload, zcl.HIGHBYTE(uint16(cluster)))
	bindRequest.Payload = append(bindRequest.Payload, 0x03) // ADDRESS_64_BIT BindAddressMode

	for j := 0; j < 8; j++ {
		b = uint8(zdo.macAddress >> uint64(8*j))
		bindRequest.Payload = append(bindRequest.Payload, b)

	}
	bindRequest.Payload = append(bindRequest.Payload, 1)

	return zdo.async_request(*bindRequest, 3*time.Second)
}

// handler the particular command
func (zdo *Zdo) handle_command(command Command) {
	//	log.Printf("zdo.handle_command:: input_command cmd.id: 0x%04x %s \n", uint16(command.Id), Command_to_string(command.Id))
	switch command.Id {
	case AF_INCOMING_MSG: // 0x4481
		if !zdo.isReady {
			return
		}
		zdo.msgChan <- command // send incoming message to controller

	case ZDO_STATE_CHANGE_IND:
		log.Printf("New zhub status = %d \n", command.Payload[0])
		if command.Payload[0] == 9 {
			zdo.isReady = true
		}

	case ZDO_MGMT_PERMIT_JOIN_RSP:
		log.Printf("Zhub permit join status = %d\n", command.Payload[2])

	case ZDO_PERMIT_JOIN_IND:
		log.Printf("Zhub permit for %d seconds \n", command.Payload[0])

	case ZDO_END_DEVICE_ANNCE_IND: //  0x45c1
		fmt.Printf("ZDO_END_DEVICE_ANNCE_IND: payload len = %d, payload:  ", command.Payload_size())
		for i := 0; i < int(command.Payload_size()); i++ {
			fmt.Printf("0x%02x ", command.Payload[i])
		}
		fmt.Println()
		zdo.joinChan <- command.Payload[2:]

	case ZDO_ACTIVE_EP_RSP: // 0x4585
		fmt.Printf("ZDO_ACTIVE_EP_RSP: payload len = %d, payload:  ", command.Payload_size())
		for i := 0; i < int(command.Payload_size()); i++ {
			fmt.Printf("0x%02x ", command.Payload[i])
		}
		fmt.Println("")

		if command.Payload[2] == byte(zcl.SUCCESS) {
			shortAddr := zcl.UINT16_(command.Payload[0], command.Payload[1])

			ep_count := command.Payload[5]
			var endpoints []byte = make([]byte, ep_count)
			log.Printf("Zdo:: Device 0x%04x Endpoints count: %d list: ", shortAddr, ep_count)
			for i := 0; i < int(ep_count); i++ { // Number of active endpoint in the list
				endpoints[i] = command.Payload[6+i]
				log.Printf("Query descriptor for endpoint %d \n", endpoints[i])
				zdo.SimpleDescriptor(shortAddr, endpoints[i])

			}
			log.Println("")
		}

	case ZDO_SIMPLE_DESC_RSP: // 0x4584
		len := command.Payload_size()
		if len > 0 {
			// fmt.Println("zdo.handle_command::ZDO_SIMPLE_DESC_RSP:: Payload: ")
			i := byte(0)
			for len > 0 {
				// fmt.Printf(" %02x ", command.Payload[i])
				i++
				len--
			}
			// fmt.Println("")
			if command.Payload[2] == byte(zcl.SUCCESS) {
				shortAddr := zcl.UINT16_(command.Payload[0], command.Payload[1])

				// descriptorLen := command.Payload[5] // длина дескриптора, начинается с номера эндпойнта

				descriptor := SimpleDescriptor{}
				descriptor.endpointNumber = uint16(command.Payload[6])                     // номер эндпойнта, для которого пришел дескриптор
				descriptor.profileId = zcl.UINT16_(command.Payload[7], command.Payload[8]) // профиль эндпойнта
				descriptor.deviceId = zcl.UINT16_(command.Payload[9], command.Payload[10]) // ID устройства
				descriptor.deviceVersion = uint16(command.Payload[11])                     // Версия устройства

				//				log.Printf("ZDO_SIMPLE_DESC_RSP:: Device 0x%04x Descriptor length %d \n", shortAddr, descriptorLen)
				log.Printf("ZDO_SIMPLE_DESC_RSP:: Device 0x%04x Endpoint %d ProfileId 0x%04x DeviceId 0x%04x \n", shortAddr, descriptor.endpointNumber, descriptor.profileId, descriptor.deviceId)
				i := 12 // Index of number of input clusters/

				inputClustersNumber := command.Payload[i]
				i++
				log.Printf("ZDO_SIMPLE_DESC_RSP: Input Cluster count %d \n ", inputClustersNumber)
				for inputClustersNumber > 0 {
					descriptor.inputClusters = []uint16{}
					p1 := command.Payload[i]
					i++
					p2 := command.Payload[i]
					i++
					fmt.Printf("ZDO_SIMPLE_DESC_RSP: Input Cluster %s 0x%04X \n", zcl.Cluster_to_string(zcl.Cluster(zcl.UINT16_(p1, p2))), zcl.UINT16_(p1, p2))

					descriptor.inputClusters = append(descriptor.inputClusters, zcl.UINT16_(p1, p2)) // List of input cluster Id's supported.
					inputClustersNumber--
				}
				fmt.Println("")
				outputClustersNumber := command.Payload[i]
				i++
				log.Printf("ZDO_SIMPLE_DESC_RSP: Output Cluster count %d \n ", outputClustersNumber)
				for outputClustersNumber > 0 {

					p1 := command.Payload[i]
					i++
					p2 := command.Payload[i]
					i++

					fmt.Printf("ZDO_SIMPLE_DESC_RSP: Output Cluster %s 0x%04X \n", zcl.Cluster_to_string(zcl.Cluster(zcl.UINT16_(p1, p2))), zcl.UINT16_(p1, p2))

					descriptor.outputClusters = append(descriptor.outputClusters, zcl.UINT16_(p1, p2)) // List of output cluster Id's supported.
					outputClustersNumber--
				}
				fmt.Println("")
			}
		}

		// unattended commands
	case ZDO_TC_DEV_IND, // 0x45ca
		ZDO_SRC_RTG_IND, // 0x45c4
		ZDO_BIND_RSP,    // 0x45a1
		ZDO_LEAVE_IND:
		{
		}
		// commands with mandatory event emitter
	case SYS_RESET_IND: // 0x4180
		zdo.eh.emit(command.Id, command)
	default: // all sync commands and unknown commands
		zdo.eh.emit(command.Id, command)
	}

}

// in payload
func (zdo *Zdo) GenerateTransactionNumber() uint8 {

	ret := zdo.transactionNumber
	zdo.transactionNumber++
	return ret
}

// in zcl.Frame
func (zdo *Zdo) GenerateTransactionSequenceNumber() uint8 {
	zdo.tsnMutex.Lock()
	defer zdo.tsnMutex.Unlock()
	zdo.transactionSecuenseNumber++
	number := zdo.transactionSecuenseNumber
	return number
}
