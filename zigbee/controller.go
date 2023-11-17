/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2023 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
*/
package zigbee

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gbatanov/sim800l/modem"
	"github.com/gbatanov/zhub4/telega32"
	"github.com/gbatanov/zhub4/zigbee/clusters"
	"github.com/gbatanov/zhub4/zigbee/zdo"
	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

func ControllerCreate(config *GlobalConfig) (*Controller, error) {
	chn1 := make(chan zdo.Command, 16)
	chn2 := make(chan []byte, 12) // chan for join command shortAddr + macAddrj
	chn3 := make(chan clusters.MotionMsg, 16)
	chn4 := make(chan clusters.MotionMsg, 1)
	ts := time.Now()

	zdoo, err := zdo.ZdoCreate(config.Port, config.Os, chn1, chn2)
	if err != nil {
		return &Controller{}, err
	}

	// Modem block
	config.WithModem = false
	//var mdm *modem.GsmModem
	mdm := modem.GsmModemCreate(config.ModemPort, 9600, config.MyPhoneNumber)
	err = mdm.Open()
	config.WithModem = err == nil

	// telegram bot block
	tlgMsgChan := make(chan telega32.Message, 16)
	tlgCmdChan := make(chan string, 2)
	tlg32 := telega32.Tlg32Create(config.BotName, config.Mode, config.TokenPath, config.MyId, tlgMsgChan, tlgCmdChan) //your bot name
	tlgBlock := TlgBlock{tlg32, tlgMsgChan, tlgCmdChan}

	// http server block
	httpBlock := HttpBlock{}
	httpBlock.answerChan = make(chan interface{}, 8)
	httpBlock.queryChan = make(chan map[string]string, 8)
	httpBlock.withHttp = true

	controller := Controller{
		zdobj:              zdoo,
		config:             config,
		devices:            map[uint64]*zdo.EndDevice{},
		devicessAddressMap: map[uint16]uint64{},
		flag:               true,
		chargerChan:        chn4,
		msgChan:            chn1,
		joinChan:           chn2,
		motionMsgChan:      chn3,
		lastMotion:         time.Now(),
		smartPlugTS:        ts,
		switchOffTS:        false,
		mapFileMutex:       sync.Mutex{},
		tlg:                tlgBlock,
		http:               httpBlock,
		startTime:          time.Now(),
		mdm:                mdm}
	return &controller, nil

}
func (c *Controller) GetZdo() *zdo.Zdo {
	return c.zdobj
}
func (c *Controller) StartNetwork() error {

	log.Println("Controller start network")
	var defconf zdo.RF_Channels
	defconf.Channels = c.config.Channels

	err := c.tlg.tlg32.Run()
	if err == nil {
		c.config.WithTlg = true
		log.Println("Telegram bot started")
		if c.config.WithModem {
			outMsg := telega32.Message{ChatId: c.config.MyId, Msg: "SIM800 started"}
			c.tlg.tlgMsgChan <- outMsg
		} else {
			outMsg := telega32.Message{ChatId: c.config.MyId, Msg: "SIM800 not started"}
			c.tlg.tlgMsgChan <- outMsg
		}

	} else {
		c.config.WithTlg = false
		log.Println("Telebot error:", err.Error())
	}

	// thread for commands handle
	go func() {
		c.GetZdo().InputCommand()
	}()

	// thread for incoming commands from uart adapter
	go func() {
		c.GetZdo().Uart.Loop(c.GetZdo().Cmdinput)
	}()

	// Incoming messages handler
	go func() {
		c.onMessage()
	}()

	// Event handler of joined device
	go func() {
		c.joinDevice()
	}()

	// Message handler from motion sensors
	go func() {
		for c.flag {
			msg := <-c.motionMsgChan
			if msg.Cmd == 2 {
				break
			}
			c.handleMotion(msg.Ed, msg.Cmd)
		}
	}()

	// Message handler from charger
	go func() {
		for c.flag {
			msg := <-c.chargerChan
			if msg.Cmd == 2 {
				break
			} else if msg.Cmd == 1 {
				outMsg := telega32.Message{ChatId: c.config.MyId, Msg: "Заряд включен"}
				c.tlg.tlgMsgChan <- outMsg
			} else if msg.Cmd == 0 {
				c.switchRelay(msg.Ed.MacAddress, 0, 1)
				outMsg := telega32.Message{ChatId: c.config.MyId, Msg: "Заряд выключен"}
				c.tlg.tlgMsgChan <- outMsg
			}
		}
	}()

	// reset of zhub
	log.Println("Controller reset adapter (wait about 1 minute)")
	err = c.GetZdo().Reset()
	if err != nil {
		return err
	}

	// set up desired RF-channels
	rf := c.GetZdo().ReadRfChannels()
	if !rf.Compare(defconf) {
		err = c.GetZdo().WriteRfChannels(defconf)
		if err != nil {
			return err
		}
	}

	err = c.GetZdo().FinishConfiguration()
	if err != nil {
		return err
	}

	// startup
	log.Println("Controller startup")
	err = c.GetZdo().Startup(100 * time.Millisecond)
	if err != nil {
		return err
	}
	log.Println("Controller register endpoint")
	err = c.GetZdo().RegisterEndpointDescriptor(zdo.Default_endpoint)
	if err != nil {
		return err
	}

	// http
	if c.http.withHttp {
		//	httpBlock.web, err = NewHttpServer(config.HttpAddress, httpBlock.answerChan, httpBlock.queryChan, config.Os, config.ProgramDir)
		c.http.web, err = NewHttpServer(c)
		c.http.withHttp = err == nil
		if c.http.withHttp {
			c.http.web.Start()
			log.Println("Web server started")
		}
	}

	c.createDevicesByMap()

	// permit join during 1 minute
	c.GetZdo().PermitJoin(60 * time.Second)

	// we will get SmurtPlug parameters  every 30 seconds
	// and check valves state
	// chek rely every 60 seconds
	go func() {
		for c.flag {
			time.Sleep(30 * time.Second)
			c.getSmartPlugParams()
			c.getCheckValves()
			time.Sleep(30 * time.Second)
			c.getSmartPlugParams()
			c.getCheckValves()
			c.getCheckRelay()

		}
	}()

	if c.config.WithTlg {
		outMsg := telega32.Message{ChatId: c.config.MyId, Msg: "Zhub4 start"}
		c.tlg.tlgMsgChan <- outMsg
		go func() {
			for c.flag {
				cmd := <-c.tlg.tlgCmdChan
				c.executeCmd(cmd)
			}
		}()
	}

	// Обработка команд с модема
	if c.config.WithModem {
		go func() {
			for c.flag {
				cmd := <-c.mdm.CmdToController
				c.executeCmd(cmd)
			}
		}()
	}

	log.Println("Controller start network success")
	return nil
}

func (c *Controller) Stop() {
	log.Println("Controller stop")
	c.flag = false
	c.tlg.tlg32.Stop()
	c.GetZdo().Stop()
	if c.http.withHttp {
		c.http.web.Stop()
	}
	if c.config.WithModem {
		defer c.mdm.Stop()
	}
	// release channels
	c.msgChan <- *zdo.NewCommand(0)
	c.chargerChan <- clusters.MotionMsg{Ed: &zdo.EndDevice{}, Cmd: 2}
	c.motionMsgChan <- clusters.MotionMsg{Ed: &zdo.EndDevice{}, Cmd: 2}
	c.joinChan <- []byte{}
}

// command with incomming message handler
func (c *Controller) onMessage() {
	for c.flag {
		command := <-c.msgChan
		if c.flag && command.Id > 0 {
			//log.Printf("Command  0x%04x\n", command.Id)
			go func(cmd zdo.Command) { c.messageHandler(cmd) }(command)
		}
	}
}

func (c *Controller) writeMapToFile() error {
	log.Println("writeMapToFile")
	m := sync.Mutex{}
	m.Lock()
	defer m.Unlock()

	filename := c.config.MapPath

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
func (c *Controller) readMapFromFile() error {

	m := sync.Mutex{}
	m.Lock()
	defer m.Unlock()
	c.devicessAddressMap = map[uint16]uint64{}

	filename := c.config.MapPath

	fd, err := os.OpenFile(filename, os.O_RDONLY, 0755)
	if err != nil {
		log.Println("ReadMap:: OpenFile error: ", err)
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
				log.Printf("0x%04x : 0x%016x \n", a, b)
			}
		}
		log.Printf("\n")
	}

	return nil
}

// Called immediately after the start of the configurator
// create devices by c.devicesAddressMap
func (c *Controller) createDevicesByMap() {

	err := c.readMapFromFile()
	if err == nil {
		for shortAddress, macAddress := range c.devicessAddressMap {
			ed := zdo.EndDeviceCreate(macAddress, shortAddress)
			c.devices[macAddress] = ed
		}
	}
}

func (c *Controller) joinDevice() {
	for c.flag {
		FullAddr := <-c.joinChan
		if c.flag && len(FullAddr) > 5 { // TODO: ??
			var shortAddress uint16 = zcl.UINT16_(FullAddr[0], FullAddr[1])
			var macAddress uint64 = binary.LittleEndian.Uint64(FullAddr[2:])

			_, deviceInList := zdo.KNOWN_DEVICES[macAddress]
			if !deviceInList {
				log.Printf("Controller::joinDevice: macAddress: 0x%016x is not in KNOWN_DEVICES\n", macAddress)
				continue
			} else {
				log.Printf("Controller::joinDevice: macAddress: 0x%016x new shortAddress: 0x%04x\n", macAddress, shortAddress)
			}

			_, keyExists := c.devices[macAddress]
			if keyExists {
				// device exists, check shortAddr
				_, keyExists := c.devicessAddressMap[shortAddress]
				if keyExists {
					// self rejoin
					continue
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
					c.writeMapToFile()
					ed := c.getDeviceByMac(macAddress)
					ed.ShortAddress = shortAddress
					c.onJoin(shortAddress, macAddress)
				}

			} else {
				log.Printf("Controller::joinDevice: create device\n")
				ed := zdo.EndDeviceCreate(macAddress, shortAddress)
				c.devices[macAddress] = ed
				c.devicessAddressMap[shortAddress] = macAddress
				c.writeMapToFile()
				c.onJoin(shortAddress, macAddress)
			}
		}
	}
}

// adjusting joined devices
func (c *Controller) onJoin(shortAddress uint16, macAddress uint64) {
	ed := c.getDeviceByMac(macAddress)
	if ed.ShortAddress == 0 || ed.ShortAddress != shortAddress {
		log.Printf("Controller:: onJoin: device 0x%016x doesn't exist, ed.shortAddress 0x%04x != shortAddres 0x%04x  \n", macAddress, ed.ShortAddress, shortAddress)
		return
	}
	c.GetZdo().Bind(shortAddress, macAddress, 1, zcl.ON_OFF)
	if ed.GetDeviceType() != 4 {
		// for customs no report adjusting in ON_OFF cluster
		c.configureReporting(shortAddress, zcl.ON_OFF, uint16(0), zcl.DataType_UINT8, uint16(0))
	}
	// SmartPlug, WaterValves - no binding and no report adjusting in ON_OFF cluster
	//
	// motion sensors and door sensors Sonoff
	if ed.GetDeviceType() == 2 || ed.GetDeviceType() == 3 {
		c.GetZdo().Bind(shortAddress, macAddress, 1, zcl.IAS_ZONE)
		c.configureReporting(shortAddress, zcl.IAS_ZONE, uint16(0), zcl.DataType_UINT8, uint16(0))
	}
	// IKEA motion sensors
	if ed.GetDeviceType() == 8 {
		c.GetZdo().Bind(shortAddress, macAddress, 1, zcl.IAS_ZONE)
		c.configureReporting(shortAddress, zcl.IAS_ZONE, uint16(0), zcl.DataType_UINT8, uint16(0))
	}
	// IKEA devices
	if ed.GetDeviceType() == 7 || ed.GetDeviceType() == 8 {
		c.getPower(ed)
	}
	c.GetZdo().Bind(shortAddress, macAddress, 1, zcl.POWER_CONFIGURATION)
	if ed.GetDeviceType() != 4 {
		c.configureReporting(shortAddress, zcl.POWER_CONFIGURATION, uint16(0), zcl.DataType_UINT8, uint16(0))
	}

	//
	c.GetZdo().ActiveEndpoints(shortAddress) // descriptors will be obtained later for each endpoint
	c.getIdentifier(shortAddress)            // For many devices this request is required!!!! Without it, the device does not work, only registration on the network

}
func (c *Controller) getIdentifier(address uint16) {
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
	frame.FrameControl.Ftype = zcl.FrameType_GLOBAL
	frame.FrameControl.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.FrameControl.DisableDefaultResponse = 1
	frame.FrameControl.ManufacturerSpecific = 0
	frame.Command = uint8(zcl.READ_ATTRIBUTES) // 0x00
	frame.TransactionSequenceNumber = c.GetZdo().Generate_transaction_sequence_number()
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

	c.GetZdo().SendMessage(endpoint, cl, frame)
}

func (c *Controller) getDeviceByShortAddr(shortAddres uint16) *zdo.EndDevice {
	// get macAddress
	_, keyExists := c.devicessAddressMap[shortAddres]
	if keyExists {
		macAddress := c.devicessAddressMap[shortAddres]
		ed := c.getDeviceByMac(macAddress)
		return ed
	}
	return &zdo.EndDevice{MacAddress: 0, ShortAddress: 0}
}

func (c *Controller) getDeviceByMac(macAddress uint64) *zdo.EndDevice {

	_, keyExists := c.devices[macAddress]
	if keyExists {
		return c.devices[macAddress]
	} else {
		return &zdo.EndDevice{MacAddress: 0, ShortAddress: 0}
	}
}

func (c *Controller) messageHandler(command zdo.Command) {

	var message zdo.Message = zdo.Message{}
	message.Cluster = zcl.Cluster(zcl.UINT16_(command.Payload[2], command.Payload[3]))
	message.Source.Address = zcl.UINT16_(command.Payload[4], command.Payload[5])
	message.Destination.Address = c.GetZdo().ShortAddress
	message.Source.Number = command.Payload[6]
	message.Destination.Number = command.Payload[7]
	message.LinkQuality = command.Payload[9]
	length := command.Payload[16]
	message.ZclFrame = c.GetZdo().ParseZclData(command.Payload[17 : 17+length])

	ed := c.getDeviceByShortAddr(message.Source.Address)
	if ed.MacAddress == 0 {
		log.Printf("message handler: device 0x%04x not found\n", message.Source.Address)
		return
	}

	//	var ts uint32 = uint32(command.Payload[11]) + uint32(command.Payload[12])<<8 + uint32(command.Payload[13])<<16 + uint32(command.Payload[14])<<24
	//	log.Printf("Cluster %s (0x%04X) device: %s \n", zcl.ClusterToString(message.Cluster), message.Cluster, ed.GetHumanName())
	/*
		if message.Cluster != zcl.TIME { // too often
			fmt.Printf("source endpoint shortAddr: 0x%04x ", message.Source.Address)
			fmt.Printf("number: %d \n", message.Source.Number)
			fmt.Printf("linkQuality: %d \n", message.LinkQuality)
			//	fmt.Printf("ts %d \n", uint32(ts/1000))
			fmt.Printf("length of ZCL data %d \n", length)
			if message.ZclFrame.ManufacturerCode != 0xffff { // Manufacturer Code absent
				fmt.Printf(" zcl_frame.manufacturer_code: %04x \n", message.ZclFrame.ManufacturerCode)
			}
			fmt.Printf("zclFrame.FrameControl.Ftype: %02x ", message.ZclFrame.FrameControl.Ftype)
			fmt.Printf("message.ZclFrame.Command: 0x%02x \n", message.ZclFrame.Command)
			fmt.Printf("message.ZclFrame.Payload: ")
			for _, b := range message.ZclFrame.Payload {
				fmt.Printf("0x%02x ", b)
			}
			fmt.Print("\n\n")
		}
	*/
	if message.LinkQuality > 0 {
		ed.Set_linkquality(message.LinkQuality)
	}
	now := time.Now()
	ed.Set_last_seen(now)

	withStatus := message.Cluster != zcl.ANALOG_INPUT &&
		message.Cluster != zcl.XIAOMI_SWITCH &&
		message.ZclFrame.Command != uint8(zcl.REPORT_ATTRIBUTES)

	if message.ZclFrame.FrameControl.Ftype == zcl.FrameType_GLOBAL {
		// commands requiring attribute parsing
		if message.ZclFrame.Command == uint8(zcl.READ_ATTRIBUTES_RESPONSE) ||
			message.ZclFrame.Command == uint8(zcl.REPORT_ATTRIBUTES) {
			if len(message.ZclFrame.Payload) > 0 {
				//				if ed.MacAddress == zdo.RELAY_7_KITCHEN {
				//					log.Println(message.ZclFrame.Payload)
				//				}
				attributes := zcl.ParseAttributesPayload(message.ZclFrame.Payload, withStatus)
				if len(attributes) > 0 {
					c.onAttributeReport(ed, message.Source, message.Cluster, attributes)
				}
			}
		}
	} else {
		// further not attributes, cluster-dependent commands that need to be responded to, cause some kind of action
		// custom does not come here, they always have AttributeReport, even when activated
		switch message.Cluster {
		case zcl.ON_OFF:
			//			log.Printf("message handler::ON_OFF: command 0x%02x \n", message.ZclFrame.Command)
			// commands from the IKEA motion sensor also come here
			c.onOffCommand(ed, message)
			c.getPower(ed) // TODO: by timer

		case zcl.LEVEL_CONTROL:
			//			log.Printf("message handler::LEVEL_CONTROL: command 0x%02x \n", message.ZclFrame.Command)
			c.level_command(ed, message)
			c.getPower(ed) // TODO: by timer

		case zcl.IAS_ZONE:
			// this cluster includes motion sensors from Sonoff and door sensors from Sonoff
			// split by device type
			if ed.GetDeviceType() == 2 { // motion sensors from Sonoff
				msg := clusters.MotionMsg{Ed: ed, Cmd: message.ZclFrame.Payload[0]}
				c.motionMsgChan <- msg
				c.getPower(ed)
			} else if ed.GetDeviceType() == 3 { // door sensors from Sonoff
				c.handleSonoffDoor(ed, message.ZclFrame.Payload[0])
				c.getPower(ed)
			} else if ed.GetDeviceType() == 5 { // water leak sensor from Aqara
				var state string = "NORMAL"
				if message.ZclFrame.Payload[0] == 1 {
					state = "ALARM"
				}
				ed.SetCurrentState(state, 1)

				if state == "ALARM" {
					c.iasZoneCommand(uint8(0), uint16(0)) // close valves, switch off wash machine
					ts := time.Now()                      // get time now
					ed.SetLastAction(ts)
					if c.config.WithTlg {
						alarmMsg := "Сработал датчик протечки: " + ed.GetHumanName()
						c.tlg.tlgMsgChan <- telega32.Message{ChatId: c.config.MyId, Msg: alarmMsg}
					}
					// gsmmodem->master_call()
				}

			}
		case zcl.IDENTIFY:
			//			log.Printf("Cluster IDENTIFY:: command 0x%02x \n", message.ZclFrame.Command)
		case zcl.ALARMS:
			//			log.Printf("Cluster ALARMS:: command 0x%02x payload %q \n", message.ZclFrame.Command, message.ZclFrame.Payload)
		case zcl.TIME:
			//fmt.Println("")
			// Approximately 30 seconds pass with the Aqara relay, no useful information
			//			log.Printf("Cluster TIME:: command 0x%02x \n\n", message.ZclFrame.Command)
		} //switch
	}
	c.afterMessageAction(ed)
}
func (c *Controller) onAttributeReport(ed *zdo.EndDevice, ep zcl.Endpoint, cluster zcl.Cluster, attributes []zcl.Attribute) {
	//	zcl.HandlerAttributes(cluster, ep, attributes)

	switch cluster {
	case zcl.BASIC:
		cl := clusters.BasicCluster{Ed: ed}
		cl.HandlerAttributes(ep, attributes)

	case zcl.POWER_CONFIGURATION:
		cl := clusters.PowerConfigurationCluster{Ed: ed}
		cl.HandlerAttributes(ep, attributes)

	case zcl.IDENTIFY:
		cl := clusters.IdentifyCluster{Ed: ed}
		cl.HandlerAttributes(ep, attributes)

	case zcl.ON_OFF:
		cl := clusters.OnOffCluster{Ed: ed, MsgChan: c.motionMsgChan}
		cl.HandlerAttributes(ep, attributes)

	case zcl.ANALOG_INPUT:
		cl := clusters.AnalogInputCluster{Ed: ed}
		cl.HandlerAttributes(ep, attributes)

	case zcl.MULTISTATE_INPUT:
		cl := clusters.MultistateInputCluster{}
		cl.HandlerAttributes(ep, attributes)

	case zcl.XIAOMI_SWITCH:
		cl := clusters.XiaomiCluster{Ed: ed}
		cl.HandlerAttributes(ep, attributes)

	case zcl.SIMPLE_METERING:
		cl := clusters.SimpleMeteringCluster{Ed: ed}
		cl.HandlerAttributes(ep, attributes)

	case zcl.ELECTRICAL_MEASUREMENTS:
		cl := clusters.ElectricalMeasurementCluster{Ed: ed, ChargerChan: c.chargerChan}
		cl.HandlerAttributes(ep, attributes)

	case zcl.TUYA_ELECTRICIAN_PRIVATE_CLUSTER:
		cl := clusters.TuyaCluster{}
		cl.HandlerAttributes1(ep, attributes)

	case zcl.TUYA_SWITCH_MODE_0:
		cl := clusters.TuyaCluster{}
		cl.HandlerAttributes2(ep, attributes)

	case zcl.IAS_ZONE:
		cl := clusters.IasZoneCluster{}
		cl.HandlerAttributes(ep, attributes)

	case zcl.ALARMS:
		cl := clusters.AlarmsCluster{}
		cl.HandlerAttributes(ep, attributes)

	case zcl.POLL_CONTROL:
		cl := clusters.PollControlCluster{}
		cl.HandlerAttributes(ep, attributes)

	case zcl.LIGHT_LINK:
		cl := clusters.LightLinkCluster{}
		cl.HandlerAttributes(ep, attributes)

	case zcl.IKEA_BUTTON:
		cl := clusters.IkeaCluster{}
		cl.HandlerAttributes(ep, attributes)

	case zcl.GROUPS:
		cl := clusters.GroupsCluster{}
		cl.HandlerAttributes(ep, attributes)

	case zcl.TIME:
		cl := clusters.TimeCluster{}
		cl.HandlerAttributes(ep, attributes)

	default: // unattended clusters

		log.Printf("unattended clusters::endpoint address: 0x%04x number = %d \n", ep.Address, ep.Number)

		for _, attribute := range attributes {
			log.Printf("Cluster 0x%04x, attribute id =0x%04x \n", cluster, attribute.Id)
		}
	}

}

// call every 30 sec - SmartPlugs
func (c *Controller) getSmartPlugParams() {
	ed := c.getDeviceByMac(zdo.PLUG_2_CHARGER) // SmartPlug charger
	if ed == nil || ed.ShortAddress == 0 {
		return
	}

	var idsAV []uint16 = []uint16{0x0505, 0x0508, 0x050B} // Voltage, Current, Energy
	c.readAttribute(ed.ShortAddress, zcl.ELECTRICAL_MEASUREMENTS, idsAV)

	var idsAVSM []uint16 = []uint16{0x0000} // Power
	c.readAttribute(ed.ShortAddress, zcl.SIMPLE_METERING, idsAVSM)

	plugs := zdo.GetDevicesByType(uint8(10))
	for _, di := range plugs {
		ed := c.getDeviceByMac(di)
		if ed.ShortAddress != 0 {
			// if the state has not yet been received
			if ed.GetCurrentState(1) != "On" && ed.GetCurrentState(1) != "Off" {
				var idsAV []uint16 = []uint16{0x0000} // state On / Off
				c.readAttribute(ed.ShortAddress, zcl.ON_OFF, idsAV)
			}
		}
	}
}

// call every 30 sec - Relay check
// отдаются "левые" параметры
func (c *Controller) getCheckRelay() {
	ed := c.getDeviceByMac(zdo.RELAY_7_KITCHEN) // Relay in kitchen
	if ed == nil || ed.ShortAddress == 0 {
		return
	}
	c.getPower(ed)
	//	var idsAV []uint16 = []uint16{0x0505, 0x0508} // Voltage, Current
	//	c.readAttribute(ed.ShortAddress, zcl.ELECTRICAL_MEASUREMENTS, idsAV)

}

// call every 30 sec - Valves check
func (c *Controller) getCheckValves() {

	valves := zdo.GetDevicesByType(uint8(6))
	for _, di := range valves {
		ed := c.getDeviceByMac(di) // valves
		if ed == nil || ed.ShortAddress == 0 {
			continue
		}
		// if the state has not yet been received
		if ed.GetCurrentState(1) != "On" && ed.GetCurrentState(1) != "Off" {
			var idsAV []uint16 = []uint16{0x0000} // state On / Off
			c.readAttribute(ed.ShortAddress, zcl.ON_OFF, idsAV)
		}
	}
}

// action after any message (they happen quite often, I use them as a timer)
func (c *Controller) afterMessageAction(ed *zdo.EndDevice) {

	var interval float64 = 20
	if c.config.Mode == "test" {
		interval = 10.0
	}
	// 20 minutes after the last movement, I capture the state "No one at home"
	// write to log and send to telegram
	lastMotion := c.getLastMotionSensorActivity()
	diffOff := time.Since(lastMotion)
	if diffOff.Minutes() > interval && !c.switchOffTS {
		c.switchOffTS = true
		c.switchOffWithList()
		log.Printf("There is no one at home\n")
		if c.config.WithTlg {
			alarmMsg := "Никого нет дома 20 минут "
			c.tlg.tlgMsgChan <- telega32.Message{ChatId: c.config.MyId, Msg: alarmMsg}
		}
	}

}

// make a request to read an attribute (attributes)
func (c *Controller) readAttribute(address uint16, cl zcl.Cluster, ids []uint16) error {

	endpoint := zcl.Endpoint{Address: address, Number: 1}

	// ZCL Header
	var frame zcl.Frame = zcl.Frame{}
	frame.FrameControl.Ftype = zcl.FrameType_GLOBAL
	frame.FrameControl.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.FrameControl.DisableDefaultResponse = 1
	frame.FrameControl.ManufacturerSpecific = 0
	frame.Command = uint8(zcl.READ_ATTRIBUTES) // 0x00
	frame.TransactionSequenceNumber = c.GetZdo().Generate_transaction_sequence_number()
	frame.ManufacturerCode = 0
	// end ZCL Header
	frame.Payload = make([]byte, 2*len(ids))

	for i := 0; i < len(ids); i++ {
		frame.Payload[0+i*2] = zcl.LOWBYTE(ids[i])
		frame.Payload[1+i*2] = zcl.HIGHBYTE(ids[i])
	}
	return c.GetZdo().SendMessage(endpoint, cl, frame)
}

// make a request to read power attributes
func (c *Controller) getPower(ed *zdo.EndDevice) {

	cluster := zcl.POWER_CONFIGURATION
	endpoint := zcl.Endpoint{Address: ed.ShortAddress, Number: 1}

	frame := zcl.Frame{}
	frame.FrameControl.Ftype = zcl.FrameType_GLOBAL
	frame.FrameControl.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.FrameControl.DisableDefaultResponse = 0
	frame.FrameControl.ManufacturerSpecific = 0
	frame.ManufacturerCode = 0
	frame.TransactionSequenceNumber = c.GetZdo().Generate_transaction_sequence_number()
	frame.Command = uint8(zcl.READ_ATTRIBUTES) // 0x00
	// in payload set of required attributes
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(uint16(zcl.PowerConfiguration_MAINS_VOLTAGE)))    // 0x0000 main voltage, 0.1V, UINT16
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(uint16(zcl.PowerConfiguration_MAINS_VOLTAGE)))   //
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(uint16(zcl.PowerConfiguration_BATTERY_VOLTAGE)))  // 0x0020 Battery voltage, 0.1V. UINT8
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(uint16(zcl.PowerConfiguration_BATTERY_VOLTAGE))) //
	frame.Payload = append(frame.Payload, zcl.LOWBYTE(uint16(zcl.PowerConfiguration_BATTERY_REMAIN)))   //  0x0021 Battery remain level, 0.5%, UINT8
	frame.Payload = append(frame.Payload, zcl.HIGHBYTE(uint16(zcl.PowerConfiguration_BATTERY_REMAIN)))  //

	c.GetZdo().SendMessage(endpoint, cluster, frame)
}

// Turn off the relay according to the list with a long press on the buttons Sonoff1 Sonoff2
func (c *Controller) switchOffWithList() {

	for _, macAddr := range zdo.OFF_LIST {
		c.switchRelay(macAddr, 0, 1)
		if macAddr == zdo.RELAY_7_KITCHEN { // the relay in the kitchen has two channel
			c.switchRelay(macAddr, 0, 2)
		}
	}

}
func (c *Controller) switchRelay(macAddress uint64, cmd uint8, channel uint8) {
	log.Printf("Relay 0x%016x switch to %d\n", macAddress, cmd)
	ed := c.getDeviceByMac(macAddress)
	if ed.ShortAddress > 0 && ed.Di.Available == 1 {
		c.sendCommandToOnoffDevice(ed.ShortAddress, cmd, channel)
		ts := time.Now() // get time now
		ed.SetLastAction(ts)
	} else {
		log.Printf("Relay 0x%016x not found\n", macAddress)
	}
}

// sending the On/Off/Toggle command to the device
// 0x01/0x00/0x02, the rest are ignored in this configuration
func (c *Controller) sendCommandToOnoffDevice(address uint16, cmd uint8, ep uint8) {
	if cmd > 2 {
		return
	}
	endpoint := zcl.Endpoint{Address: address, Number: ep}
	cluster := zcl.ON_OFF

	frame := zcl.Frame{}
	frame.FrameControl.Ftype = zcl.FrameType_SPECIFIC
	frame.FrameControl.ManufacturerSpecific = 0
	frame.FrameControl.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.FrameControl.DisableDefaultResponse = 0
	frame.ManufacturerCode = 0
	frame.TransactionSequenceNumber = c.GetZdo().Generate_transaction_sequence_number()
	frame.Command = cmd

	c.GetZdo().SendMessage(endpoint, cluster, frame)
}

func (c *Controller) configureReporting(address uint16,
	cluster zcl.Cluster,
	attributeId uint16,
	attributeDataType zcl.DataType,
	reportable_change uint16) error {

	endpoint := zcl.Endpoint{Address: address, Number: 1}
	// ZCL Header
	frame := zcl.Frame{}
	frame.FrameControl.Ftype = zcl.FrameType_GLOBAL
	frame.FrameControl.ManufacturerSpecific = 0
	frame.FrameControl.Direction = zcl.FROM_CLIENT_TO_SERVER
	frame.FrameControl.DisableDefaultResponse = 0
	frame.ManufacturerCode = 0
	frame.TransactionSequenceNumber = c.GetZdo().Generate_transaction_sequence_number()
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

	return c.GetZdo().SendMessage(endpoint, cluster, frame)
}

func (c *Controller) setLastMotionSensorActivity(lastTime time.Time) {
	if lastTime.Compare(c.lastMotion) > 0 {
		c.lastMotion = lastTime
	}
}
func (c *Controller) getLastMotionSensorActivity() time.Time { return c.lastMotion }

// Исполнение команд из СМС и от телеграм-бота
// Ответ отправляем в СМС и в телеграм
// Команды могут быть информационные - /balance и управляющие /cmnd
func (c *Controller) executeCmd(cmnd string) {
	log.Println("Execute cmd ", cmnd)
	cmnd = strings.Trim(cmnd, " ")
	if strings.HasPrefix(cmnd, "/cmnd") {
		var cmd int
		n, err := fmt.Sscanf(cmnd, "/cmnd %d", &cmd)
		if n == 0 || err != nil || cmd < 400 || cmd > 499 {
			return
		}

		switch cmd {
		case 401: // Запрос баланса сим-карты
			c.mdm.GetBalance()
		case 412: // Запрос состояния датчиков протечек
		case 423: // Запрос состояния датчиков движения
		}
	} else if strings.HasPrefix(cmnd, "/balance") {
		// Пришел ответ на запрос баланса, отправим СМС и в телеграм
		cmnd = strings.Replace(cmnd, "/balance ", "", 1)
		c.mdm.SendSms(cmnd)
		c.tlg.tlgMsgChan <- telega32.Message{ChatId: c.config.MyId, Msg: cmnd}
	}
}

func Mapkey(m map[uint16]uint64, value uint64) (key uint16, ok bool) {
	for k, v := range m {
		if v == value {
			return k, true
		}
	}
	return 0, false
}
