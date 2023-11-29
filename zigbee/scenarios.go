/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2023 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
*/
package zigbee

import (
	"fmt"
	"log"
	"time"

	"github.com/gbatanov/zhub4/telega32"
	"github.com/gbatanov/zhub4/zigbee/zdo"
)

// scenarios
func (c *Controller) handleMotion(ed *zdo.EndDevice, cmd uint8) {

	state := "No motion"
	cur_motion := ed.GetMotionState()
	// Fix the last activity of the motion sensor
	// I fix in the activity only the activation of the sensor
	// Since the motion sensor is also in custom, which send messages periodically,
	// need to check the current state and commit the change
	// turn on something for movement and set a sign that there is someone in the house
	// for the water shutdown algorithm in case of power failure
	if int8(cmd) != cur_motion {
		if cmd == 1 {
			ts := time.Now() // get time now
			c.setLastMotionSensorActivity(ts)
			ed.SetLastAction(ts)
			c.switchOffTS = false
		}
		log.Println("handleMotion")
	}
	ed.SetMotionState(cmd) // numeric value
	if cmd == 1 {
		state = "Motion"
	}
	ed.SetCurrentState(state, 1) // text value

	macAddress := ed.MacAddress

	if macAddress == zdo.MOTION_1_CORIDOR { //Sonoff motion sensor 1 (coridor)
		lum := int8(-1)
		/*
			//  on/off  light in coridor zdo.RELAY_4_CORIDOR_LIGHT
			//  works in couple with custom2
			custom2 := c.getDeviceByMac(zdo.MOTION_LIGHT_CORIDOR)
			if custom2.ShortAddress > 0 {
				lum = custom2.Get_luminocity()
			}
			log.Printf("Motion sensor in coridor. cmd = %d, lum = %d\n", cmd, lum)
		*/
		if cmd == 1 {
			if lum != 1 {
				log.Printf("Motion sensor in coridor. Turn on light relay. \n")
				c.switchRelay(zdo.RELAY_4_CORIDOR_LIGHT, 1, 1)
			}
		} else if cmd == 0 {
			cur_motion := -1
			//			if custom2.ShortAddress > 0 {
			//				cur_motion = custom2.GetMotionState()
			//			}
			if cur_motion != 1 {
				log.Printf("Motion sensor in coridor. Turn off light relay. \n")
				c.switchRelay(zdo.RELAY_4_CORIDOR_LIGHT, 0, 1)
			}
		}
	} else if macAddress == zdo.MOTION_LIGHT_CORIDOR {
		// motion sensor in custom2 (hallway)
		// it is necessary to take into account the illumination when turned on and the state of the motion sensor in the corridor
		lum := ed.Get_luminocity()
		log.Printf("Motion sensor in custom2. cmd = %d, lum = %d\n", cmd, lum)

		relay := c.getDeviceByMac(zdo.RELAY_4_CORIDOR_LIGHT)
		relayCurrentState := relay.GetCurrentState(1)

		if cmd == 1 && relayCurrentState != "On" {
			// since the sensor sometimes falsely triggers, we ignore its triggering at night
			h, _, _ := time.Now().Clock()
			if h > 7 && h < 23 {
				if lum != 1 {
					log.Printf("Motion sensor in hallway. Turn on light relay. \n")
					c.switchRelay(zdo.RELAY_4_CORIDOR_LIGHT, 1, 1)
				}
			}
		} else if cmd == 0 && relayCurrentState != "Off" {
			motion1 := c.getDeviceByMac(0x00124b0025137475)
			if motion1.ShortAddress > 0 {
				cur_motion = motion1.GetMotionState()
			}
			if cur_motion != 1 {
				log.Printf("Motion sensor in hallway. Turn off light relay. \n")
				c.switchRelay(zdo.RELAY_4_CORIDOR_LIGHT, 0, 1)
			}
		}
	} else if macAddress == zdo.PRESENCE_1_KITCHEN {
		relay := c.getDeviceByMac(0x00158d0009414d7e)
		relayCurrentState := relay.GetCurrentState(1)
		// presence sensor 1, on/off light in kitchen - relay 7 endpoint 1
		if cmd == 1 && relayCurrentState != "On" {
			log.Printf("Turn on light in kitchen")
			c.switchRelay(0x00158d0009414d7e, 1, 1)
			c.kitchenPresenceChan <- 1
		} /*else if cmd == 0 && relayCurrentState != "Off" {
			log.Printf("Turn off light in kitchen")
			c.switchRelay(0x00158d0009414d7e, 0, 1)
		}*/
	} else if macAddress == 0x00124b002444d159 {
		// motion sensor 3, coridor
		if cmd == 1 {
			log.Println("Motion3 On")
		} else {
			log.Println("Motion3 Off")
		}
	} else if macAddress == 0x00124b0024455048 {
		// motion sensor 2 (bedroom)
		if cmd == 1 {
			log.Println("Motion2 On")
		} else {
			log.Println("Motion2 Off")
		}
	} else if macAddress == zdo.MOTION_IKEA {
		// IKEA motion sensor
		if cmd == 1 {
			log.Println("IKEA motion sensor On")
			// switch off by timer
			c.ikeaMotionChan <- cmd
			c.switchRelay(zdo.RELAY_1, 1, 1)
		} else {
			log.Println("IKEA motion sensor Off")
		}

	} else if macAddress == zdo.MOTION_LIGHT_NURSERY {
		// motion sensor in Custom3(children room)
		if cmd == 1 {
			log.Println("Motion sensor in Custom3(children room) On")
		} else {
			log.Println("Motion sensor in Custom3(children room) Off")
		}
	}
}

// таймер выключения по датчику движения Ikea
func (c *Controller) IkeaMotionTimer() {
	ed := c.getDeviceByMac(zdo.MOTION_IKEA)
	if ed.ShortAddress != 0 {

		var timer1 *time.Timer = &time.Timer{}
		go func() {
			for {
				select {
				case state, ok := <-c.ikeaMotionChan:
					if !ok {
						// channel was closed
						log.Println("Ikea motion channel was closed")
						return
					}
					if state == 1 {
						// Запускаем таймер на 3 минуты
						timer1 = time.NewTimer(180 * time.Second)
						log.Println("Ikea motion timer start")
					}

				case <-timer1.C:
					// таймер сработал
					log.Println("Ikea motion timer signal")
					c.handleMotion(ed, 0) // датчик Икеа сам не подает сигнал выключения
					c.switchRelay(zdo.RELAY_1, 0, 1)
				}
			}
		}()
	}
}

// таймер выключения по датчику присутствия на кухне
func (c *Controller) KitchenPresenceTimer() {
	ed := c.getDeviceByMac(zdo.PRESENCE_1_KITCHEN)
	if ed.ShortAddress != 0 {

		var timer1 *time.Timer = &time.Timer{}
		go func() {
			for {
				select {
				case state, ok := <-c.kitchenPresenceChan:
					if !ok {
						// channel was closed
						log.Println("kitchen presence channel was closed")
						return
					}
					if state == 1 {
						// Запускаем таймер на 3 минуты
						timer1 = time.NewTimer(180 * time.Second)
						log.Println("kitchen presence timer start")
					}

				case <-timer1.C:
					// таймер сработал
					log.Println("kitchen presence timer signal")
					c.switchRelay(zdo.RELAY_7_KITCHEN, 0, 1)
				}
			}
		}()
	}
}

// 0x01 On
// 0x02 Toggle
// 0x40 Off with effect
// 0x41 On with recall global scene
// 0x42 On with timed off  payload:0x00 (исполняем безусловно) 0x08 0x07(ON на 0x0708 (180,0)секунд) 0x00 0x00
func (c *Controller) onOffCommand(ed *zdo.EndDevice, message zdo.Message) {
	//log.Println("onOffCommand")

	macAddress := ed.MacAddress
	cmd := message.ZclFrame.Command
	state := "Off"
	if cmd == 1 {
		state = "On"
	}
	ed.SetCurrentState(state, 1)
	ts := time.Now() // get time now
	ed.SetLastAction(ts)

	if macAddress == zdo.BUTTON_IKEA {
		// IKEA button on/off only
		c.ikea_button_action(cmd)
	} else if macAddress == zdo.MOTION_IKEA {
		// IKEA motion sensor
		c.handleMotion(ed, 1) // switch off with timer
	} else if macAddress == zdo.BUTTON_SONOFF_1 {
		// button Sonoff1
		switch cmd { // 1 - double , 2 - single, 0 - long press
		case 0:
			ed.SetCurrentState("Long click", 1)
			c.switchOffWithList()
		case 1:
			ed.SetCurrentState("Double click", 1)
		case 2:
			ed.SetCurrentState("Single click", 1)
		} //switch
	} else if macAddress == zdo.BUTTON_SONOFF_2 {
		// button Sonoff 2 call ringer with double click
		// switch off relays by list with long press
		switch cmd {
		case 0:
			ed.SetCurrentState("Long click", 1)
			c.switchOffWithList()
		case 1:
			ed.SetCurrentState("Double click", 1)
			c.ringer()
			if c.config.WithTlg {
				c.tlg.tlgMsgChan <- telega32.Message{ChatId: c.config.MyId, Msg: "Вызов с кнопки "}
			}
		case 2:
			ed.SetCurrentState("Single click", 1)
			// toogle relay 0x54ef441000609dcc
			c.switchRelay(0x54ef441000609dcc, 2, 1)
		}
	}
}

// I have IKEA button only with this function
func (c *Controller) level_command(ed *zdo.EndDevice, message zdo.Message) {
	fmt.Println("level_command")
	ts := time.Now() // get time now
	ed.SetLastAction(ts)

	cmd := message.ZclFrame.Command // 5 - Hold+, 7 - button realised, 1 - Hold-

	log.Printf("IKEA button level command: 0x%0x \n", cmd)

	switch cmd {
	case 1: // Hold
		ed.SetCurrentState("Minus down", 1)
		log.Printf("IKEA button: Hold-\n")
	case 5:
		ed.SetCurrentState("Plus down", 1)
		log.Printf("IKEA button: Hold+\n")
	case 7:
		ed.SetCurrentState("No act", 1)
		log.Printf("IKEA button: realised\n")
	}
	fmt.Println("")
}

// processing a command from an active IAS_ZONE device
// in my design - leakage sensors
// you need to turn off the taps and relays of the washing machine
// {0xa4c138d9758e1dcd, {"Water Valve", "TUYA", "Valve", "Кран 1 ГВ", zigbee::zcl::Cluster::ON_OFF}},
// {0xa4c138373e89d731, {"Water Valve", "TUYA", "Valve", "Кран 2 ХВ", zigbee::zcl::Cluster::ON_OFF}}
// {0x54ef441000193352, {"lumi.switch.n0agl1", "Xiaomi", "SSM-U01", "Реле 2(стиральная машина)", zigbee::zcl::Cluster::ON_OFF}}
// washing machine contactor normally closed,
// when the executive relay is turned on, the contactor turns off
func (c *Controller) iasZoneCommand(cmnd uint8, shortAddr uint16) {
	executiveDevices := [3]uint64{
		zdo.VALVE_HOT_WATER,
		zdo.VALVE_COLD_WATER,
		zdo.RELAY_2_WASH}

	cmd := cmnd // automatically turn off only
	// enable/switch command is used only via web api
	if shortAddr > 0 {
		// one device, command via web api
		dev, res := Mapkey(c.devicessAddressMap, zdo.RELAY_2_WASH) // washing machine contactor relay
		if res && dev == shortAddr {
			cmd = 1 - cmnd
		}

		c.sendCommandToOnoffDevice(shortAddr, cmd, 1)
		log.Printf("Close device 0x%04x\n", shortAddr)
	} else {
		// all executive devices
		for _, macAddress := range executiveDevices {
			cmd1 := cmd
			if macAddress == zdo.RELAY_2_WASH {
				cmd1 = 1 - cmd
			}
			_, res := Mapkey(c.devicessAddressMap, macAddress)
			if res {
				ts := time.Now() // get time now
				ed := c.getDeviceByMac(macAddress)
				if ed.ShortAddress > 0 {
					ed.SetLastAction(ts)

					// the Valves close slowly, we start everything in parallel threads
					go func(short uint16, cmd uint8) {
						c.sendCommandToOnoffDevice(short, cmd, 1)
					}(shortAddr, cmd1)
				}
			}
		} //for
	}
}

func (c *Controller) handleSonoffDoor(ed *zdo.EndDevice, cmd uint8) {
	ts := time.Now() // get time now
	ed.SetLastAction(ts)

	state := "Closed"
	if ed.MacAddress == 0x00124b002512a60b { // door sensor 2
		if cmd == 1 {
			state = "Opened"
		}
		ed.SetCurrentState(state, 1)
		// control relay 3, turn on the backlight in the cabinet in the  bedroom
		c.switchRelay(0x54ef44100018b523, cmd, 1)
	} else if ed.MacAddress == 0x00124b00250bba63 { // door sensor 3
		if cmd == 1 {
			state = "Opened"
		}
		ed.SetCurrentState(state, 1)
		if c.config.WithTlg {
			alarmMsg := "Закрыт ящик "
			if cmd == 0x01 {
				alarmMsg = "Открыт ящик "
			}
			c.tlg.tlgMsgChan <- telega32.Message{ChatId: c.config.MyId, Msg: alarmMsg}
		}
	} else if ed.MacAddress == 0x00124b0025485ee6 { //door sensor 1, sensor in toilet
		if cmd == 0 {
			state = "Opened"
		}
		ed.SetCurrentState(state, 1)

		// light "Toilet occupied" in the corridor is turned on only from 9 am to 10 pm
		// turn off at any time (to cancel manual turn on)
		// enable/disable the relay 0x54ef4410005b2639 - light relay "Toilet occupied"
		if cmd == 0x01 {
			c.switchRelay(zdo.RELAY_5_TOILET, 0, 1)
		} else if cmd == 0 {
			h, _, _ := time.Now().Clock()
			if h > 7 && h < 23 {
				c.switchRelay(zdo.RELAY_5_TOILET, 1, 1)
			}
		}
	}
}

// IKEA button on/off action only
func (c *Controller) ikea_button_action(cmd uint8) {
	if cmd == 1 {
		log.Println("IKEA button on action")
	} else {
		log.Println("IKEA button off action")
	}

}

func (c *Controller) ringer() {
	// TODO: системный вызов на ноуте, либо через USB-RS232
}
