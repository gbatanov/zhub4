package zigbee

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"zhub4/zigbee/zcl"
)

type Controller struct {
	zdo                *Zdo
	mode               string
	devices            map[uint64]*EndDevice
	devicessAddressMap map[uint16]uint64
	flag               bool
	msgChan            chan Command   // chanel for receive incoming message command from zdo
	joinChan           chan []byte    // chanel for receive command join device from zdo
	motionMsgChan      chan MotionMsg // chanel for get message from motion sensors
	lastMotion         time.Time      // last motion any motion sensor
	smartPlugTS        time.Time      // timestamp for smart plug timer
	switchOffTS        time.Time      // timestamp for switch off timer
	mapFileMutex       sync.Mutex
}

func init() {
	fmt.Println("Init in zigbee: controller")
}
func controllerCreate(Ports map[string]string, Os string, mode string) (*Controller, error) {
	chn1 := make(chan Command, 1)
	chn2 := make(chan []byte, 1)
	chn3 := make(chan MotionMsg, 16)
	ts := time.Now()

	zdo, err := zdoCreate(Ports[Os], Os, chn1, chn2)
	if err != nil {
		zdo, err = zdoCreate(Ports[Os+"2"], Os, chn1, chn2)
	}
	if err != nil {
		return &Controller{}, err
	}

	controller := Controller{zdo: zdo,
		mode:               mode,
		devices:            map[uint64]*EndDevice{},
		devicessAddressMap: map[uint16]uint64{},
		flag:               true,
		msgChan:            chn1,
		joinChan:           chn2,
		motionMsgChan:      chn3,
		lastMotion:         time.Time{},
		smartPlugTS:        ts,
		switchOffTS:        ts,
		mapFileMutex:       sync.Mutex{}}
	return &controller, nil

}
func (c *Controller) Get_zdo() *Zdo {
	return c.zdo
}
func (c *Controller) startNetwork(defconf NetworkConfiguration) error {

	log.Println("Controller start network")
	// thread for commands handle
	go func() {
		c.zdo.input_command()
	}()

	// thread for incoming commands from uart adapter
	go func() {
		c.zdo.Uart.Loop(c.zdo.Cmdinput)
	}()

	//
	go func() {
		c.on_message()
	}()
	go func() {
		c.on_join()
	}()

	go func() {
		for c.flag {
			msg := <-c.motionMsgChan
			c.handle_motion(msg.ed, msg.cmd)
		}
	}()

	// reset of coordinator
	log.Println("Controller reset adapter")
	_, err := c.zdo.Reset(RESET_TYPE_HARD)
	if err != nil {
		return err
	}

	// we have hard reset, check network configuration
	log.Println("Controller read NetworkConfiguration")
	nc, err := c.zdo.readNetworkConfiguration()
	if err != nil {
		return err
	}
	if !nc.Compare(defconf) {
		// rewrite configuration in coordinator
		log.Println("Controller write NetworkConfiguration")
		err = c.zdo.writeNetworkConfiguration(defconf)
		if err != nil {
			return err
		}
		// soft reset of coordinator after reconfiguration
		log.Println("Controller soft reset coordinator")
		_, err := c.zdo.Reset(RESET_TYPE_SOFT)
		if err != nil {
			return err
		}

	}

	// startup
	log.Println("Controller startup")
	err = c.zdo.Startup(100 * time.Millisecond)
	if err != nil {
		return err
	}
	log.Println("Controller register endpoint")
	err = c.zdo.registerEndpointDescriptor(default_endpoint)
	if err != nil {
		return err
	}

	c.create_devices_by_map()
	c.zdo.Permit_join(60 * time.Second)

	log.Println("Controller start network success")
	return nil
}

func (c *Controller) Stop() {
	log.Println("Controller stop")
	c.flag = false
	c.zdo.stop()
	// release channels
	c.msgChan <- *NewCommand(0)
	c.joinChan <- []byte{}
}

// command with incomming message handler
func (c *Controller) on_message() {
	for c.flag {
		command := <-c.msgChan
		if c.flag && command.Id > 0 {
			log.Printf("Command %s \n", command.Id.String())
			go func(cmd Command) { c.message_handler(cmd) }(command)
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
				fmt.Printf("%d 0x%04x 0x%016x\n", r, shortAddr, macAddr)
			}
		}
		fd.Close()
		for a, b := range c.devicessAddressMap {
			fmt.Printf("0x%04x : 0x%016x \n", a, b)
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
			ed := EndDeviceCreate(macAddress, shortAddress)
			c.devices[macAddress] = ed
		}
	}
}

func (c *Controller) on_join() {
	for c.flag {
		FullAddr := <-c.joinChan
		if c.flag && len(FullAddr) > 5 { // TODO: ??
			var shortAddress uint16 = UINT16_(FullAddr[0], FullAddr[1])
			var macAddress uint64 = binary.LittleEndian.Uint64(FullAddr[2:])
			log.Printf("Controller::on_join: macAddress: 0x%016x \n", macAddress)
			log.Printf("Controller::on_join: new shortAddress: 0x%04x\n", shortAddress)

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
					c.join_device(shortAddress, macAddress)
				}

			} else {
				log.Printf("Controller::on_join: create device\n")
				ed := EndDeviceCreate(macAddress, shortAddress)
				c.devices[macAddress] = ed
				c.devicessAddressMap[shortAddress] = macAddress
				c.write_map_to_file()
				c.join_device(shortAddress, macAddress)
			}
		}
	}
}

func (c *Controller) join_device(shortAddress uint16, macAddress uint64) {
	ed := c.get_device_by_mac(macAddress)
	if ed.shortAddress == 0 {
		return
	}
	c.zdo.bind(shortAddress, macAddress, 1, zcl.ON_OFF)
	if ed.get_device_type() != 4 {
		// for customs no report adjusting in ON_OFF cluster
		c.configureReporting(shortAddress, zcl.ON_OFF, uint16(0), DataType_UINT8, uint16(0))
	}
	// SmartPlug, WaterValves - no binding and no report adjusting in ON_OFF cluster
	// motion sensors and door sensors Sonoff
	if ed.get_device_type() == 2 || ed.get_device_type() == 3 {
		c.zdo.bind(shortAddress, macAddress, 1, zcl.IAS_ZONE)
		c.configureReporting(shortAddress, zcl.IAS_ZONE, uint16(0), DataType_UINT8, uint16(0))
	}
	// IKEA motion sensors
	if ed.get_device_type() == 8 {
		c.zdo.bind(shortAddress, macAddress, 1, zcl.IAS_ZONE)
		c.configureReporting(shortAddress, zcl.IAS_ZONE, uint16(0), DataType_UINT8, uint16(0))
	}
	// IKEA devices
	if ed.get_device_type() == 7 || ed.get_device_type() == 8 {
		c.get_power(ed)
	}
	c.zdo.bind(shortAddress, macAddress, 1, zcl.POWER_CONFIGURATION)
	if ed.get_device_type() != 4 {
		c.configureReporting(shortAddress, zcl.POWER_CONFIGURATION, uint16(0), DataType_UINT8, uint16(0))
	}

	//
	c.zdo.activeEndpoints(shortAddress)
	c.zdo.simpleDescriptor(shortAddress, 1) // TODO: получить со всех эндпойнотов, полученных на предыдущем этапе

	c.get_identifier(shortAddress) // Для многих устройств этот запрос обязателен!!!! Без него не работатет устройство, только регистрация в сети

}
func (c *Controller) get_identifier(address uint16) {
	//   zigbee::Message message;
	cl := zcl.BASIC
	var id, id2, id3, id4, id5, id6 uint16
	endpoint := Endpoint{address, 1}
	id = uint16(Basic_MODEL_IDENTIFIER)
	id2 = uint16(Basic_MANUFACTURER_NAME)
	id3 = uint16(Basic_SW_BUILD_ID)          // SW_BUILD_ID = 0x4000
	id4 = uint16(Basic_PRODUCT_LABEL)        //
	id5 = uint16(Basic_DEVICE_ENABLED)       // у датчиков движения Sonoff его нет
	id6 = uint16(Basic_PHYSICAL_ENVIRONMENT) //

	// ZCL Header
	frame := zcl.Frame{}
	frame.Frame_control.Ftype = zcl.FrameType_GLOBAL
	frame.Frame_control.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.Frame_control.DisableDefaultResponse = 1
	frame.Frame_control.ManufacturerSpecific = 0
	frame.Command = uint8(zcl.READ_ATTRIBUTES) // 0x00
	frame.TransactionSequenceNumber = c.zdo.generateTransactionNumber()
	// end ZCL Header

	frame.Payload = make([]byte, 0)
	frame.Payload = append(frame.Payload, LOWBYTE(id))
	frame.Payload = append(frame.Payload, HIGHBYTE(id))
	frame.Payload = append(frame.Payload, LOWBYTE(id2))
	frame.Payload = append(frame.Payload, HIGHBYTE(id2))
	frame.Payload = append(frame.Payload, LOWBYTE(id3))
	frame.Payload = append(frame.Payload, HIGHBYTE(id3))
	frame.Payload = append(frame.Payload, LOWBYTE(id4))
	frame.Payload = append(frame.Payload, HIGHBYTE(id4))
	frame.Payload = append(frame.Payload, LOWBYTE(id5))
	frame.Payload = append(frame.Payload, HIGHBYTE(id5))
	frame.Payload = append(frame.Payload, LOWBYTE(id6))
	frame.Payload = append(frame.Payload, HIGHBYTE(id6))

	c.zdo.send_message(endpoint, cl, frame, 3*time.Second)
}

func (c *Controller) get_device_by_short_addr(shortAddres uint16) *EndDevice {
	// get macAddress
	_, keyExists := c.devicessAddressMap[shortAddres]
	if keyExists {
		macAddress := c.devicessAddressMap[shortAddres]
		ed := c.get_device_by_mac(macAddress)
		return ed
	}
	return &EndDevice{macAddress: 0, shortAddress: 0}
}

func (c *Controller) get_device_by_mac(macAddress uint64) *EndDevice {
	_, keyExists := c.devices[macAddress]
	if keyExists {
		return c.devices[macAddress]
	} else {
		return &EndDevice{macAddress: 0, shortAddress: 0}
	}
}

func (c *Controller) message_handler(command Command) {

	var message Message = Message{}
	message.cluster = zcl.Cluster(UINT16_(command.Payload[2], command.Payload[3]))
	message.source.address = UINT16_(command.Payload[4], command.Payload[5])
	message.destination.address = c.zdo.shortAddress
	message.source.number = command.Payload[6]
	message.destination.number = command.Payload[7]
	message.linkQuality = command.Payload[9]
	length := command.Payload[16]
	message.zclFrame = c.zdo.parse_zcl_data(command.Payload[17 : 17+length])

	ed := c.get_device_by_short_addr(message.source.address)
	if ed.macAddress == 0 {
		log.Printf("message handler: device not found\n")
		return
	}
	//	macAddress := ed.get_mac_address()
	var ts uint32 = uint32(command.Payload[11]) + uint32(command.Payload[12])<<8 + uint32(command.Payload[13])<<16 + uint32(command.Payload[14])<<24
	log.Printf("ClusterID: 0x%04x \n", message.cluster)
	log.Printf("source shortAddr: 0x%04x \n", message.source.address)
	log.Printf("source endpoint: 0x%02x \n", message.source.number)
	log.Printf("linkQuality: %d \n", message.linkQuality)
	log.Printf("ts %d \n", uint32(ts/1000))
	log.Printf("length of ZCL data %d \n", length)
	log.Printf("zcl_frame.frame_type: %02x \n", message.zclFrame.Frame_control.Ftype)
	if message.zclFrame.ManufacturerCode != 0xffff { // Manufacturer Code absent
		log.Printf(" zcl_frame.manufacturer_code: %04x \n", message.zclFrame.ManufacturerCode)
	}
	log.Printf(" message.zcl_frame.command: 0x%02x \n", message.zclFrame.Command)
	fmt.Printf("zdo: message.zcl_frame.payload: ")
	for b := range message.zclFrame.Payload {
		fmt.Printf("0x%02x ", b)
	}
	fmt.Println("")

	if message.linkQuality > 0 {
		ed.set_linkQuality(message.linkQuality)
	}
	now := time.Now()
	ed.set_last_seen(now)

	withStatus := (message.cluster != zcl.ANALOG_INPUT &&
		message.cluster != zcl.XIAOMI_SWITCH &&
		message.zclFrame.Command != uint8(zcl.REPORT_ATTRIBUTES))

	// commands requiring attribute parsing
	if ((message.zclFrame.Command == uint8(zcl.READ_ATTRIBUTES_RESPONSE) || message.zclFrame.Command == uint8(zcl.REPORT_ATTRIBUTES)) && message.cluster != zcl.ON_OFF && message.cluster != zcl.LEVEL_CONTROL) ||
		(message.zclFrame.Command == uint8(zcl.REPORT_ATTRIBUTES) && message.cluster == zcl.ON_OFF) ||
		(message.zclFrame.Command == uint8(zcl.READ_ATTRIBUTES_RESPONSE) && message.cluster == zcl.ON_OFF && len(message.zclFrame.Payload) > 0) {

		if len(message.zclFrame.Payload) > 0 {
			attributes := parse_attributes_payload(message.zclFrame.Payload, withStatus)
			if len(attributes) > 0 {
				c.on_attribute_report(ed, message.source, message.cluster, attributes)
			}
		}
	} else {
		// further not attributes, cluster-dependent commands that need to be responded to, cause some kind of action
		// custom does not come here, they always have AttributeReport, even when activated
		switch message.cluster {
		case zcl.ON_OFF:
			// commands from the IKEA motion sensor also come here
			c.onoff_command(ed, message)
		case zcl.LEVEL_CONTROL:
			c.level_command(ed, message)
		case zcl.IAS_ZONE:
			// this cluster includes motion sensors from Sonoff and door sensors from Sonoff
			// split by device type
			if ed.get_device_type() == 2 { // motion sensors from Sonoff
				msg := MotionMsg{ed: ed, cmd: message.zclFrame.Payload[0]}
				c.motionMsgChan <- msg
				c.get_power(ed)
			} else if ed.get_device_type() == 3 { // door sensors from Sonoff
				c.handle_sonoff_door(ed, message.zclFrame.Payload[0])
				c.get_power(ed)
			} else if ed.get_device_type() == 5 { // water leak sensor from Aqara
				var state string = "NORMAL"
				if message.zclFrame.Payload[0] == 1 {
					state = "ALARM"
				}
				ed.set_current_state(state, 1)

				if state == "ALARM" {
					c.ias_zone_command(uint8(0), uint16(0)) // close valves, switch off wash machine
					ts := time.Now()                        // get time now
					ed.set_last_action(ts)
					// alarm_msg := "Сработал датчик протечки: "
					// alarm_msg = alarm_msg + ed.get_human_name()
					// tlg32->send_message(alarm_msg)
					// gsmmodem->master_call()
				}

			}
		} //switch
	}
	c.after_message_action(ed)
}
func (c *Controller) on_attribute_report(ed *EndDevice, ep Endpoint, cluster zcl.Cluster, attributes []Attribute) {
	switch cluster {
	case zcl.BASIC:
		c := BasicCluster{ed: ed} // TODO: remove controller dependency
		c.handler_attributes(ep, attributes)
	case zcl.POWER_CONFIGURATION:
		c := PowerConfigurationCluster{ed: ed}
		c.handler_attributes(ep, attributes)
	case zcl.ON_OFF:
		c := OnOffCluster{ed: ed, msgChan: c.motionMsgChan}
		c.handler_attributes(ep, attributes)
	case zcl.ANALOG_INPUT:
		c := AnalogInputCluster{ed: ed}
		c.handler_attributes(ep, attributes)
	case zcl.MULTISTATE_INPUT:
		c := MultistateInputCluster{}
		c.handler_attributes(ep, attributes)
	case zcl.XIAOMI_SWITCH:
		c := XiaomiCluster{ed: ed}
		c.handler_attributes(ep, attributes)
	case zcl.SIMPLE_METERING:
		c := SimpleMeteringCluster{}
		c.handler_attributes(ep, attributes)
	case zcl.ELECTRICAL_MEASUREMENTS:
		c := ElectricalMeasurementCluster{ed: ed}
		c.handler_attributes(ep, attributes)
	case zcl.TUYA_ELECTRICIAN_PRIVATE_CLUSTER:
		c := TuyaCluster{}
		c.handler_attributes1(ep, attributes)
	case zcl.TUYA_SWITCH_MODE_0:
		c := TuyaCluster{}
		c.handler_attributes2(ep, attributes)
	default: // unattended clusters
		for _, attribute := range attributes {
			log.Printf("Cluster 0x%04x, attribute id =0x%04x \n", cluster, attribute.id)
		}
	}
}
func (c *Controller) after_message_action(ed *EndDevice) {
	if ed.get_device_type() == 10 { // SmartPlug
		// request current and voltage for every 5 minutes
		diff := time.Since(c.smartPlugTS)
		if diff.Seconds() > 300 {
			c.smartPlugTS = time.Now()
			var idsAV []uint16 = []uint16{0x0505, 0x508}

			c.read_attribute(ed.shortAddress, zcl.ELECTRICAL_MEASUREMENTS, idsAV)

			if ed.get_current_state(1) != "On" && ed.get_current_state(1) != "Off" {
				var idsAV []uint16 = []uint16{0x0000}
				c.read_attribute(ed.shortAddress, zcl.ON_OFF, idsAV)
			}
		}
	}
	diffOff := time.Since(c.switchOffTS)
	if diffOff.Minutes() > 30 {
		c.switchOffTS = time.Now()
		c.switch_off_with_list()
	}
}
func (c *Controller) read_attribute(address uint16, cl zcl.Cluster, ids []uint16) error {

	endpoint := Endpoint{address, 1}

	// ZCL Header
	var frame zcl.Frame = zcl.Frame{}
	frame.Frame_control.Ftype = zcl.FrameType_GLOBAL
	frame.Frame_control.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.Frame_control.DisableDefaultResponse = 1
	frame.Frame_control.ManufacturerSpecific = 0
	frame.Command = uint8(zcl.READ_ATTRIBUTES) // 0x00
	frame.TransactionSequenceNumber = c.zdo.generateTransactionNumber()
	frame.ManufacturerCode = 0
	// end ZCL Header
	frame.Payload = make([]byte, 2*len(ids))

	for i := 0; i < len(ids); i++ {
		frame.Payload[0+i*2] = LOWBYTE(ids[i])
		frame.Payload[1+i*2] = HIGHBYTE(ids[i])
	}
	return c.zdo.send_message(endpoint, cl, frame, 3*time.Second)
}
func (c *Controller) get_power(ed *EndDevice) {
	// var cluster zcl.Cluster = zcl.POWER_CONFIGURATION
}

// Turn off the relay according to the list with a long press on the buttons Sonoff1 Sonoff2
func (c *Controller) switch_off_with_list() {

	for _, macAddr := range OFF_LIST {
		c.switch_relay(macAddr, 0, 1)
		if macAddr == 0x00158d0009414d7e { // the relay in the kitchen has two channel
			c.switch_relay(macAddr, 0, 2)
		}
	}

}
func (c *Controller) switch_relay(macAddress uint64, cmd uint8, channel uint8) {
	ed := c.get_device_by_mac(macAddress)
	if ed.shortAddress > 0 {
		c.send_command_to_onoff_device(ed.shortAddress, cmd, channel)
		ts := time.Now() // get time now
		ed.set_last_action(ts)
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
	endpoint := Endpoint{address, ep}
	cluster := zcl.ON_OFF

	frame := zcl.Frame{}
	frame.Frame_control.Ftype = zcl.FrameType_SPECIFIC
	frame.Frame_control.ManufacturerSpecific = 0
	frame.Frame_control.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.Frame_control.DisableDefaultResponse = 0
	frame.ManufacturerCode = 0
	frame.TransactionSequenceNumber = c.zdo.generateTransactionNumber()
	frame.Command = cmd

	c.zdo.send_message(endpoint, cluster, frame, 3*time.Second)
}

func (c *Controller) configureReporting(address uint16,
	cluster zcl.Cluster,
	attributeId uint16,
	attributeDataType DataType,
	reportable_change uint16) error {

	endpoint := Endpoint{address, 1}
	// ZCL Header
	frame := zcl.Frame{}
	frame.Frame_control.Ftype = zcl.FrameType_GLOBAL
	frame.Frame_control.ManufacturerSpecific = 0
	frame.Frame_control.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.Frame_control.DisableDefaultResponse = 0
	frame.ManufacturerCode = 0
	frame.TransactionSequenceNumber = c.zdo.generateTransactionNumber()
	frame.Command = byte(zcl.CONFIGURE_REPORTING) // 0x06
	// end ZCL Header

	var min_interval uint16 = 60   // 1 minutes
	var max_interval uint16 = 3600 // 1 hours

	frame.Payload = make([]byte, 9)
	frame.Payload = append(frame.Payload, LOWBYTE(attributeId))
	frame.Payload = append(frame.Payload, HIGHBYTE(attributeId))
	frame.Payload = append(frame.Payload, byte(attributeDataType))
	frame.Payload = append(frame.Payload, LOWBYTE(min_interval))
	frame.Payload = append(frame.Payload, HIGHBYTE(min_interval))
	frame.Payload = append(frame.Payload, LOWBYTE(max_interval))
	frame.Payload = append(frame.Payload, HIGHBYTE(max_interval))
	frame.Payload = append(frame.Payload, LOWBYTE(reportable_change))
	frame.Payload = append(frame.Payload, HIGHBYTE(reportable_change))

	return c.zdo.send_message(endpoint, cluster, frame, 3*time.Second)
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
