/*
GSB, 2023
gbatanov@yandex.ru
*/
package zigbee

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"zhub4/zigbee/clhandler"
	"zhub4/zigbee/zdo"
	"zhub4/zigbee/zdo/zcl"
)

type Controller struct {
	zdobj              *zdo.Zdo
	mode               string
	devices            map[uint64]*zdo.EndDevice
	devicessAddressMap map[uint16]uint64
	flag               bool
	msgChan            chan zdo.Command         // chanel for receive incoming message command from zdo
	joinChan           chan []byte              // chanel for receive command join device from zdo
	motionMsgChan      chan clhandler.MotionMsg // chanel for get message from motion sensors
	lastMotion         time.Time                // last motion any motion sensor
	smartPlugTS        time.Time                // timestamp for smart plug timer
	switchOffTS        bool                     // flag for switch off timer
	mapFileMutex       sync.Mutex
}

func init() {
	fmt.Println("Init in zigbee: controller")
}
func controllerCreate(Ports map[string]string, Os string, mode string) (*Controller, error) {
	chn1 := make(chan zdo.Command, 16)
	chn2 := make(chan []byte, 12) // chan for join command shortAddr + macAddrj
	chn3 := make(chan clhandler.MotionMsg, 16)
	ts := time.Now()

	zdoo, err := zdo.ZdoCreate(Ports[Os], Os, chn1, chn2)
	if err != nil {
		zdoo, err = zdo.ZdoCreate(Ports[Os+"2"], Os, chn1, chn2)
	}
	if err != nil {
		return &Controller{}, err
	}

	controller := Controller{
		zdobj:              zdoo,
		mode:               mode,
		devices:            map[uint64]*zdo.EndDevice{},
		devicessAddressMap: map[uint16]uint64{},
		flag:               true,
		msgChan:            chn1,
		joinChan:           chn2,
		motionMsgChan:      chn3,
		lastMotion:         time.Now(),
		smartPlugTS:        ts,
		switchOffTS:        false,
		mapFileMutex:       sync.Mutex{}}
	return &controller, nil

}
func (c *Controller) Get_zdo() *zdo.Zdo {
	return c.zdobj
}
func (c *Controller) startNetwork(defconf zdo.NetworkConfiguration) error {

	log.Println("Controller start network")
	// thread for commands handle
	go func() {
		c.Get_zdo().Input_command()
	}()

	// thread for incoming commands from uart adapter
	go func() {
		c.Get_zdo().Uart.Loop(c.Get_zdo().Cmdinput)
	}()

	//
	go func() {
		c.on_message()
	}()
	go func() {
		c.join_device()
	}()

	go func() {
		for c.flag {
			msg := <-c.motionMsgChan
			c.handle_motion(msg.Ed, msg.Cmd)
		}
	}()

	// reset of zhub
	log.Println("Controller reset adapter")
	err := c.Get_zdo().Reset()
	if err != nil {
		return err
	}

	// we have hard reset, check network configuration
	log.Println("Controller read NetworkConfiguration")
	nc, err := c.Get_zdo().ReadNetworkConfiguration()
	if err != nil {
		return err
	}
	if !nc.Compare(defconf) {
		// rewrite configuration in zhub
		log.Println("Controller write NetworkConfiguration")
		err = c.Get_zdo().WriteNetworkConfiguration(defconf)
		if err != nil {
			return err
		}
		// soft reset of zhub after reconfiguration
		log.Println("Controller soft reset zhub")
		err := c.Get_zdo().Reset()
		if err != nil {
			return err
		}
	}

	// startup
	log.Println("Controller startup")
	err = c.Get_zdo().Startup(100 * time.Millisecond)
	if err != nil {
		return err
	}
	log.Println("Controller register endpoint")
	err = c.Get_zdo().RegisterEndpointDescriptor(zdo.Default_endpoint)
	if err != nil {
		return err
	}

	c.create_devices_by_map()
	c.Get_zdo().Permit_join(60 * time.Second)

	log.Println("Controller start network success")
	return nil
}

func (c *Controller) Stop() {
	log.Println("Controller stop")
	c.flag = false
	c.Get_zdo().Stop()
	// release channels
	c.msgChan <- *zdo.NewCommand(0)
	c.joinChan <- []byte{}
}

// command with incomming message handler
func (c *Controller) on_message() {
	for c.flag {
		command := <-c.msgChan
		if c.flag && command.Id > 0 {
			//			log.Printf("Command  0x%04x\n", command.Id)
			go func(cmd zdo.Command) { c.message_handler(cmd) }(command)
		}
	}
}

func (c *Controller) write_map_to_file() error {
	log.Println("write_map_to_file")
	m := sync.Mutex{}
	m.Lock()
	defer m.Unlock()

	prefix := "/usr/local"
	filename := prefix + "/etc/zhub4/map_addr_test.cfg"

	fd, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	} else {
		for a, b := range c.devicessAddressMap {
			fmt.Fprintf(fd, "%04x %016x\n", a, b)
			fmt.Printf("%04x %016x\n", a, b)
		}
		fd.Sync()
		fd.Close()
	}
	return nil
}

// read map from file on start the program
func (c *Controller) read_map_from_file() error {
	m := sync.Mutex{}
	m.Lock()
	defer m.Unlock()
	c.devicessAddressMap = map[uint16]uint64{}

	prefix := "/usr/local"
	filename := prefix + "/etc/zhub4/map_addr_test.cfg"

	fd, err := os.OpenFile(filename, os.O_RDONLY, 0755)
	if err != nil {
		fmt.Println("OpenFile error: ", err)
	} else {

		var shortAddr uint16
		var macAddr uint64
		var r int
		var err error = nil
		for err == nil {
			r, err = fmt.Fscanf(fd, "%4x %16x\n", &shortAddr, &macAddr)
			if r > 0 {
				c.devicessAddressMap[shortAddr] = macAddr
			}
		}
		fd.Close()
		if true {
			for a, b := range c.devicessAddressMap {
				fmt.Printf("0x%04x : 0x%016x \n", a, b)
			}
		}
	}

	return nil
}

// Вызываем сразу после старта конфигуратора
// создаем устройства по c.devicessAddressMap
func (c *Controller) create_devices_by_map() {

	err := c.read_map_from_file()
	if err == nil {
		for shortAddress, macAddress := range c.devicessAddressMap {
			ed := zdo.EndDeviceCreate(macAddress, shortAddress)
			c.devices[macAddress] = ed
		}
	}
}

func (c *Controller) join_device() {
	for c.flag {
		FullAddr := <-c.joinChan
		if c.flag && len(FullAddr) > 5 { // TODO: ??
			var shortAddress uint16 = zcl.UINT16_(FullAddr[0], FullAddr[1])
			var macAddress uint64 = binary.LittleEndian.Uint64(FullAddr[2:])
			log.Printf("Controller::join_device: macAddress: 0x%016x \n", macAddress)
			log.Printf("Controller::join_device: new shortAddress: 0x%04x\n", shortAddress)

			_, keyExists := c.devices[macAddress]
			if keyExists {
				// device exists, check shortAddr
				_, keyExists := c.devicessAddressMap[shortAddress]
				if keyExists {
					// self rejoin
					return
				} else {
					// rejoin
					// remove old shortAddress
					removingAddress := uint16(0)
					for short, mac := range c.devicessAddressMap {
						if mac == macAddress {
							removingAddress = short
							break
						}
					}
					if removingAddress > 0 {
						delete(c.devicessAddressMap, removingAddress)
					}
					c.devicessAddressMap[shortAddress] = macAddress
					c.write_map_to_file()
					ed := c.get_device_by_mac(macAddress)
					ed.ShortAddress = shortAddress
					c.on_join(shortAddress, macAddress)
				}

			} else {
				log.Printf("Controller::join_device: create device\n")
				ed := zdo.EndDeviceCreate(macAddress, shortAddress)
				c.devices[macAddress] = ed
				c.devicessAddressMap[shortAddress] = macAddress
				c.write_map_to_file()
				c.on_join(shortAddress, macAddress)
			}
		}
	}
}

func (c *Controller) on_join(shortAddress uint16, macAddress uint64) {
	ed := c.get_device_by_mac(macAddress)
	if ed.ShortAddress == 0 || ed.ShortAddress != shortAddress {
		log.Printf("Controller:: on_join: device 0x%016x doesn't exist, ed.shortAddress 0x%04x != shortAddres 0x%04x  \n", macAddress, ed.ShortAddress, shortAddress)
		return
	}
	c.Get_zdo().Bind(shortAddress, macAddress, 1, zcl.ON_OFF)
	if ed.Get_device_type() != 4 {
		// for customs no report adjusting in ON_OFF cluster
		c.configureReporting(shortAddress, zcl.ON_OFF, uint16(0), zcl.DataType_UINT8, uint16(0))
	}
	// SmartPlug, WaterValves - no binding and no report adjusting in ON_OFF cluster
	//
	// motion sensors and door sensors Sonoff
	if ed.Get_device_type() == 2 || ed.Get_device_type() == 3 {
		c.Get_zdo().Bind(shortAddress, macAddress, 1, zcl.IAS_ZONE)
		c.configureReporting(shortAddress, zcl.IAS_ZONE, uint16(0), zcl.DataType_UINT8, uint16(0))
	}
	// IKEA motion sensors
	if ed.Get_device_type() == 8 {
		c.Get_zdo().Bind(shortAddress, macAddress, 1, zcl.IAS_ZONE)
		c.configureReporting(shortAddress, zcl.IAS_ZONE, uint16(0), zcl.DataType_UINT8, uint16(0))
	}
	// IKEA devices
	if ed.Get_device_type() == 7 || ed.Get_device_type() == 8 {
		c.get_power(ed)
	}
	c.Get_zdo().Bind(shortAddress, macAddress, 1, zcl.POWER_CONFIGURATION)
	if ed.Get_device_type() != 4 {
		c.configureReporting(shortAddress, zcl.POWER_CONFIGURATION, uint16(0), zcl.DataType_UINT8, uint16(0))
	}

	//
	c.Get_zdo().ActiveEndpoints(shortAddress)
	c.Get_zdo().SimpleDescriptor(shortAddress, 1) // TODO: получить со всех эндпойнотов, полученных на предыдущем этапе

	c.get_identifier(shortAddress) // Для многих устройств этот запрос обязателен!!!! Без него не работатет устройство, только регистрация в сети

}
func (c *Controller) get_identifier(address uint16) {
	//   zigbee::Message message;
	cl := zcl.BASIC
	var id, id2, id3, id4, id5, id6 uint16
	endpoint := zcl.Endpoint{Address: address, Number: 1}
	id = uint16(zcl.Basic_MODEL_IDENTIFIER)
	id2 = uint16(zcl.Basic_MANUFACTURER_NAME)
	id3 = uint16(zcl.Basic_SW_BUILD_ID)          // SW_BUILD_ID = 0x4000
	id4 = uint16(zcl.Basic_PRODUCT_LABEL)        //
	id5 = uint16(zcl.Basic_DEVICE_ENABLED)       // у датчиков движения Sonoff его нет
	id6 = uint16(zcl.Basic_PHYSICAL_ENVIRONMENT) //

	// ZCL Header
	frame := zcl.Frame{}
	frame.Frame_control.Ftype = zcl.FrameType_GLOBAL
	frame.Frame_control.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.Frame_control.DisableDefaultResponse = 1
	frame.Frame_control.ManufacturerSpecific = 0
	frame.Command = uint8(zcl.READ_ATTRIBUTES) // 0x00
	frame.TransactionSequenceNumber = c.Get_zdo().GenerateTransactionSequenceNumber()
	// end ZCL Header

	frame.Payload = make([]byte, 0)
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(id))
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(id))
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(id2))
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(id2))
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(id3))
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(id3))
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(id4))
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(id4))
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(id5))
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(id5))
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(id6))
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(id6))

	c.Get_zdo().Send_message(endpoint, cl, frame, 3*time.Second)
}

func (c *Controller) get_device_by_short_addr(shortAddres uint16) *zdo.EndDevice {
	// get macAddress
	_, keyExists := c.devicessAddressMap[shortAddres]
	if keyExists {
		macAddress := c.devicessAddressMap[shortAddres]
		ed := c.get_device_by_mac(macAddress)
		return ed
	}
	return &zdo.EndDevice{MacAddress: 0, ShortAddress: 0}
}

func (c *Controller) get_device_by_mac(macAddress uint64) *zdo.EndDevice {
	_, keyExists := c.devices[macAddress]
	if keyExists {
		return c.devices[macAddress]
	} else {
		return &zdo.EndDevice{MacAddress: 0, ShortAddress: 0}
	}
}

func (c *Controller) message_handler(command zdo.Command) {

	var message zdo.Message = zdo.Message{}
	message.Cluster = zcl.Cluster(zcl.UINT16_(command.Payload[2], command.Payload[3]))
	message.Source.Address = zcl.UINT16_(command.Payload[4], command.Payload[5])
	message.Destination.Address = c.Get_zdo().ShortAddress
	message.Source.Number = command.Payload[6]
	message.Destination.Number = command.Payload[7]
	message.LinkQuality = command.Payload[9]
	length := command.Payload[16]
	message.ZclFrame = c.Get_zdo().Parse_zcl_data(command.Payload[17 : 17+length])

	ed := c.get_device_by_short_addr(message.Source.Address)
	if ed.MacAddress == 0 {
		log.Printf("message handler: device not found\n")
		return
	}

	//	var ts uint32 = uint32(command.Payload[11]) + uint32(command.Payload[12])<<8 + uint32(command.Payload[13])<<16 + uint32(command.Payload[14])<<24
	log.Printf("Cluster %s (0x%04X) \n", zcl.Cluster_to_string(message.Cluster), message.Cluster)
	if message.Cluster != zcl.TIME { // too often
		fmt.Printf("source endpoint shortAddr: 0x%04x ", message.Source.Address)
		fmt.Printf("number: 0x%02x \n", message.Source.Number)
		fmt.Printf("linkQuality: %d \n", message.LinkQuality)
		//	fmt.Printf("ts %d \n", uint32(ts/1000))
		fmt.Printf("length of ZCL data %d \n", length)
		if message.ZclFrame.ManufacturerCode != 0xffff { // Manufacturer Code absent
			fmt.Printf(" zcl_frame.manufacturer_code: %04x \n", message.ZclFrame.ManufacturerCode)
		}
		fmt.Printf("zclFrame.Frame_control.Ftype: %02x ", message.ZclFrame.Frame_control.Ftype)
		fmt.Printf("message.ZclFrame.Command: 0x%02x \n", message.ZclFrame.Command)
		fmt.Printf("message.ZclFrame.Payload: ")
		for _, b := range message.ZclFrame.Payload {
			fmt.Printf("0x%02x ", b)
		}
		fmt.Print("\n\n")
	}
	if message.LinkQuality > 0 {
		ed.Set_linkQuality(message.LinkQuality)
	}
	now := time.Now()
	ed.Set_last_seen(now)

	withStatus := message.Cluster != zcl.ANALOG_INPUT &&
		message.Cluster != zcl.XIAOMI_SWITCH &&
		message.ZclFrame.Command != uint8(zcl.REPORT_ATTRIBUTES)
	if message.ZclFrame.Frame_control.Ftype == zcl.FrameType_GLOBAL {
		// commands requiring attribute parsing
		if message.ZclFrame.Command == uint8(zcl.READ_ATTRIBUTES_RESPONSE) ||
			message.ZclFrame.Command == uint8(zcl.REPORT_ATTRIBUTES) {
			if len(message.ZclFrame.Payload) > 0 {
				attributes := zcl.Parse_attributes_payload(message.ZclFrame.Payload, withStatus)
				if len(attributes) > 0 {
					c.on_attribute_report(ed, message.Source, message.Cluster, attributes)
				}
			}
		}
	} else {
		// further not attributes, cluster-dependent commands that need to be responded to, cause some kind of action
		// custom does not come here, they always have AttributeReport, even when activated
		switch message.Cluster {
		case zcl.ON_OFF:
			log.Printf("message handler::ON_OFF: command 0x%02x \n", message.ZclFrame.Command)
			// commands from the IKEA motion sensor also come here
			c.onoff_command(ed, message)
			c.get_power(ed) // TODO: by timer

		case zcl.LEVEL_CONTROL:
			log.Printf("message handler::LEVEL_CONTROL: command 0x%02x \n", message.ZclFrame.Command)
			c.level_command(ed, message)
			c.get_power(ed) // TODO: by timer

		case zcl.IAS_ZONE:
			// this cluster includes motion sensors from Sonoff and door sensors from Sonoff
			// split by device type
			if ed.Get_device_type() == 2 { // motion sensors from Sonoff
				msg := clhandler.MotionMsg{Ed: ed, Cmd: message.ZclFrame.Payload[0]}
				c.motionMsgChan <- msg
				c.get_power(ed)
			} else if ed.Get_device_type() == 3 { // door sensors from Sonoff
				c.handle_sonoff_door(ed, message.ZclFrame.Payload[0])
				c.get_power(ed)
			} else if ed.Get_device_type() == 5 { // water leak sensor from Aqara
				var state string = "NORMAL"
				if message.ZclFrame.Payload[0] == 1 {
					state = "ALARM"
				}
				ed.Set_current_state(state, 1)

				if state == "ALARM" {
					c.ias_zone_command(uint8(0), uint16(0)) // close valves, switch off wash machine
					ts := time.Now()                        // get time now
					ed.Set_last_action(ts)
					// alarm_msg := "Сработал датчик протечки: "
					// alarm_msg = alarm_msg + ed.get_human_name()
					// tlg32->send_message(alarm_msg)
					// gsmmodem->master_call()
				}

			}
		case zcl.IDENTIFY:
			log.Printf("Cluster IDENTIFY:: command 0x%02x \n", message.ZclFrame.Command)
		} //switch
	}
	c.after_message_action(ed)
}
func (c *Controller) on_attribute_report(ed *zdo.EndDevice, ep zcl.Endpoint, cluster zcl.Cluster, attributes []zcl.Attribute) {
	//	zcl.Handler_attributes(cluster, ep, attributes)

	switch cluster {
	case zcl.BASIC:

		c := clhandler.BasicCluster{Ed: ed}
		c.Handler_attributes(ep, attributes)

	case zcl.POWER_CONFIGURATION:

		c := clhandler.PowerConfigurationCluster{Ed: ed}
		c.Handler_attributes(ep, attributes)

	case zcl.IDENTIFY:

		c := clhandler.IdentifyCluster{Ed: ed}
		c.Handler_attributes(ep, attributes)

	case zcl.ON_OFF:

		c := clhandler.OnOffCluster{Ed: ed, MsgChan: c.motionMsgChan}
		c.Handler_attributes(ep, attributes)

	case zcl.ANALOG_INPUT:

		c := clhandler.AnalogInputCluster{Ed: ed}
		c.Handler_attributes(ep, attributes)

	case zcl.MULTISTATE_INPUT:

		c := clhandler.MultistateInputCluster{}
		c.Handler_attributes(ep, attributes)

	case zcl.XIAOMI_SWITCH:

		c := clhandler.XiaomiCluster{Ed: ed}
		c.Handler_attributes(ep, attributes)

	case zcl.SIMPLE_METERING:

		c := clhandler.SimpleMeteringCluster{}
		c.Handler_attributes(ep, attributes)

	case zcl.ELECTRICAL_MEASUREMENTS:

		c := clhandler.ElectricalMeasurementCluster{Ed: ed}
		c.Handler_attributes(ep, attributes)

	case zcl.TUYA_ELECTRICIAN_PRIVATE_CLUSTER:

		c := clhandler.TuyaCluster{}
		c.Handler_attributes1(ep, attributes)

	case zcl.TUYA_SWITCH_MODE_0:

		c := clhandler.TuyaCluster{}
		c.Handler_attributes2(ep, attributes)

	default: // unattended clusters

		log.Printf("unattended clusters::endpoint address: 0x%04x number = %d \n", ep.Address, ep.Number)

		for _, attribute := range attributes {
			log.Printf("Cluster 0x%04x, attribute id =0x%04x \n", cluster, attribute.Id)
		}
	}

}
func (c *Controller) after_message_action(ed *zdo.EndDevice) {
	if ed.Get_device_type() == 10 { // SmartPlug
		// request current and voltage for every 5 minutes
		diff := time.Since(c.smartPlugTS)
		if diff.Seconds() > 300 {
			c.smartPlugTS = time.Now()
			var idsAV []uint16 = []uint16{0x0505, 0x508}

			c.read_attribute(ed.ShortAddress, zcl.ELECTRICAL_MEASUREMENTS, idsAV)

			if ed.Get_current_state(1) != "On" && ed.Get_current_state(1) != "Off" {
				var idsAV []uint16 = []uint16{0x0000}
				c.read_attribute(ed.ShortAddress, zcl.ON_OFF, idsAV)
			}
		}
	}
	lastMotion := c.get_last_motion_sensor_activity()
	diffOff := time.Since(lastMotion)
	if diffOff.Minutes() > 30 && !c.switchOffTS {
		c.switchOffTS = true
		c.switch_off_with_list()
	}
}
func (c *Controller) read_attribute(address uint16, cl zcl.Cluster, ids []uint16) error {

	endpoint := zcl.Endpoint{Address: address, Number: 1}

	// ZCL Header
	var frame zcl.Frame = zcl.Frame{}
	frame.Frame_control.Ftype = zcl.FrameType_GLOBAL
	frame.Frame_control.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.Frame_control.DisableDefaultResponse = 1
	frame.Frame_control.ManufacturerSpecific = 0
	frame.Command = uint8(zcl.READ_ATTRIBUTES) // 0x00
	frame.TransactionSequenceNumber = c.Get_zdo().GenerateTransactionSequenceNumber()
	frame.ManufacturerCode = 0
	// end ZCL Header
	frame.Payload = make([]byte, 2*len(ids))

	for i := 0; i < len(ids); i++ {
		frame.Payload[0+i*2] = zcl.LOWBYTE(ids[i])
		frame.Payload[1+i*2] = zcl.HIGHBYTE(ids[i])
	}
	return c.Get_zdo().Send_message(endpoint, cl, frame, 3*time.Second)
}
func (c *Controller) get_power(ed *zdo.EndDevice) {
	// var cluster zcl.Cluster = zcl.POWER_CONFIGURATION
	cluster := zcl.POWER_CONFIGURATION
	endpoint := zcl.Endpoint{Address: ed.ShortAddress, Number: 1}

	frame := zcl.Frame{}
	frame.Frame_control.Ftype = zcl.FrameType_GLOBAL
	frame.Frame_control.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.Frame_control.DisableDefaultResponse = 0
	frame.Frame_control.ManufacturerSpecific = 0
	frame.ManufacturerCode = 0
	frame.TransactionSequenceNumber = c.Get_zdo().GenerateTransactionSequenceNumber()
	frame.Command = uint8(zcl.READ_ATTRIBUTES) // 0x00
	// in payload set of required attributes
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(uint16(zcl.PowerConfiguration_MAINS_VOLTAGE)))    // 0x0000 Напряжение основного питания в 0,1 В UINT16
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(uint16(zcl.PowerConfiguration_MAINS_VOLTAGE)))   //
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(uint16(zcl.PowerConfiguration_BATTERY_VOLTAGE)))  // 0x0020 возвращает напряжение батарейки в десятых долях вольта UINT8
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(uint16(zcl.PowerConfiguration_BATTERY_VOLTAGE))) //
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(uint16(zcl.PowerConfiguration_BATTERY_REMAIN)))   //  0x0021 Остаток заряда батареи в процентах
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(uint16(zcl.PowerConfiguration_BATTERY_REMAIN)))  //

	c.Get_zdo().Send_message(endpoint, cluster, frame, 10*time.Second)
}

// Turn off the relay according to the list with a long press on the buttons Sonoff1 Sonoff2
func (c *Controller) switch_off_with_list() {

	for _, macAddr := range zdo.OFF_LIST {
		c.switch_relay(macAddr, 0, 1)
		if macAddr == 0x00158d0009414d7e { // the relay in the kitchen has two channel
			c.switch_relay(macAddr, 0, 2)
		}
	}

}
func (c *Controller) switch_relay(macAddress uint64, cmd uint8, channel uint8) {
	ed := c.get_device_by_mac(macAddress)
	if ed.ShortAddress > 0 && ed.Di.Available == 1 {
		c.send_command_to_onoff_device(ed.ShortAddress, cmd, channel)
		ts := time.Now() // get time now
		ed.Set_last_action(ts)
	} else {
		log.Printf("Relay 0x%016x not found\n", macAddress)
	}
}

// sending the On/Off/Toggle command to the device
// 0x01/0x00/0x02, the rest are ignored in this configuration
func (c *Controller) send_command_to_onoff_device(address uint16, cmd uint8, ep uint8) {
	if cmd > 2 {
		return
	}
	endpoint := zcl.Endpoint{Address: address, Number: ep}
	cluster := zcl.ON_OFF

	frame := zcl.Frame{}
	frame.Frame_control.Ftype = zcl.FrameType_SPECIFIC
	frame.Frame_control.ManufacturerSpecific = 0
	frame.Frame_control.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.Frame_control.DisableDefaultResponse = 0
	frame.ManufacturerCode = 0
	frame.TransactionSequenceNumber = c.Get_zdo().GenerateTransactionSequenceNumber()
	frame.Command = cmd

	c.Get_zdo().Send_message(endpoint, cluster, frame, 3*time.Second)
}

func (c *Controller) configureReporting(address uint16,
	cluster zcl.Cluster,
	attributeId uint16,
	attributeDataType zcl.DataType,
	reportable_change uint16) error {

	endpoint := zcl.Endpoint{Address: address, Number: 1}
	// ZCL Header
	frame := zcl.Frame{}
	frame.Frame_control.Ftype = zcl.FrameType_GLOBAL
	frame.Frame_control.ManufacturerSpecific = 0
	frame.Frame_control.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.Frame_control.DisableDefaultResponse = 0
	frame.ManufacturerCode = 0
	frame.TransactionSequenceNumber = c.Get_zdo().GenerateTransactionSequenceNumber()
	frame.Command = byte(zcl.CONFIGURE_REPORTING) // 0x06
	// end ZCL Header

	var min_interval uint16 = 60   // 1 minutes
	var max_interval uint16 = 3600 // 1 hours

	frame.Payload = make([]byte, 9)
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(attributeId))
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(attributeId))
	frame.Payload = append(frame.Payload, byte(attributeDataType))
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(min_interval))
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(min_interval))
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(max_interval))
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(max_interval))
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(reportable_change))
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(reportable_change))

	return c.Get_zdo().Send_message(endpoint, cluster, frame, 3*time.Second)
}

func (c *Controller) set_last_motion_sensor_activity(lastTime time.Time) {
	if lastTime.Compare(c.lastMotion) > 0 {
		c.lastMotion = lastTime
	}
}
func (c *Controller) get_last_motion_sensor_activity() time.Time { return c.lastMotion }

func Mapkey(m map[uint16]uint64, value uint64) (key uint16, ok bool) {
	for k, v := range m {
		if v == value {
			return k, true
		}
	}
	return 0, false
}
