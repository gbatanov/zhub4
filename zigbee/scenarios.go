/*
GSB, 2023
gbatanov@yandex.ru
*/
package zigbee

import (
	"fmt"
	"log"
	"time"
	"zhub4/pi4"
	"zhub4/zigbee/zdo"
)

// scenarios
func (c *Controller) handle_motion(ed *EndDevice, cmd uint8) {
	fmt.Println("handle_motion")

	state := "No motion"
	cur_motion := ed.get_motion_state()
	// Fix the last activity of the motion sensor
	// We fix in the activity only the activation of the sensor
	// Since the motion sensor is also in custom, which send messages periodically,
	// need to check the current state and commit the change
	// turn on something for movement and set a sign that there is someone in the house for the water shutdown algorithm in case of power failure
	if int8(cmd) != cur_motion {
		if cmd == 1 {
			ts := time.Now() // get time now
			c.set_last_motion_sensor_activity(ts)
			ed.set_last_action(ts)
			c.switchOffTS = false
		}
	}
	ed.set_motion_state(cmd) // numeric value
	if cmd == 1 {
		state = "Motion"
	}
	ed.set_current_state(state, 1) // text value

	macAddress := ed.macAddress

	if macAddress == 0x00124b0025137475 { //Sonoff motion sensor 1 (coridor)
		lum := int8(-1)

		//  on/off  light in coridor 0x54ef4410001933d3
		//  works in couple with custom2
		custom2 := c.get_device_by_mac(0x00124b0014db2724)
		if custom2.shortAddress > 0 {
			lum = custom2.get_luminocity()
		}
		log.Printf("Motion sensor in coridor. cmd = %d, lum = %d\n", cmd, lum)
		if cmd == 1 {
			if lum != 1 {
				log.Printf("Motion sensor in coridor. Turn on light relay. \n")
				c.switch_relay(0x54ef4410001933d3, 1, 1)
			}
		} else if cmd == 0 {
			if custom2.shortAddress > 0 {
				cur_motion = custom2.get_motion_state()
			}
			if cur_motion != 1 {
				log.Printf("Motion sensor in coridor. Turn off light relay. \n")
				c.switch_relay(0x54ef4410001933d3, 0, 1)
			}
		}
	} else if macAddress == 0x00124b0014db2724 {
		// motion sensor in custom2 (hallway)
		// it is necessary to take into account the illumination when turned on and the state of the motion sensor in the corridor
		lum := ed.get_luminocity()
		log.Printf("Motion sensor in custom2. cmd = %d, lum = %d\n", cmd, lum)

		relay := c.get_device_by_mac(0x54ef4410001933d3)
		relayCurrentState := relay.get_current_state(1)

		if cmd == 1 && relayCurrentState != "On" {
			// since the sensor sometimes falsely triggers, we ignore its triggering at night
			h, _, _ := time.Now().Clock()
			if h > 7 && h < 23 {
				if lum != 1 {
					log.Printf("Motion sensor in hallway. Turn on light relay. \n")
					c.switch_relay(0x54ef4410001933d3, 1, 1)
				}
			}
		} else if cmd == 0 && relayCurrentState != "Off" {
			motion1 := c.get_device_by_mac(0x00124b0025137475)
			if motion1.shortAddress > 0 {
				cur_motion = motion1.get_motion_state()
			}
			if cur_motion != 1 {
				log.Printf("Motion sensor in hallway. Turn off light relay. \n")
				c.switch_relay(0x54ef4410001933d3, 0, 1)
			}
		}
	} else if macAddress == 0x00124b0009451438 {
		relay := c.get_device_by_mac(0x00158d0009414d7e)
		relayCurrentState := relay.get_current_state(1)
		// presence sensor 1, on/off light in kitchen - relay 7 endpoint 1
		if cmd == 1 && relayCurrentState != "On" {
			log.Printf("Turn on light in kitchen")
			c.switch_relay(0x00158d0009414d7e, 1, 1)
		} else if cmd == 0 && relayCurrentState != "Off" {
			log.Printf("Turn off light in kitchen")
			c.switch_relay(0x00158d0009414d7e, 0, 1)
		}
	} else if macAddress == 0x00124b002444d159 {
		// motion sensor 3, children room
		fmt.Print("Motion3 ")
		if cmd == 1 {
			fmt.Println("On")
		} else {
			fmt.Println("Off")
		}
	} else if macAddress == 0x00124b0024455048 {
		// motion sensor 2 (bedroom)
		fmt.Print("Motion2 ")
		if cmd == 1 {
			fmt.Println("On")
		} else {
			fmt.Println("Off")
		}
	} else if macAddress == 0x0c4314fffe17d8a8 {
		// IKEA motion sensor
		fmt.Print("IKEA motion sensor ")
		if cmd == 1 {
			fmt.Println("On")
		} else {
			fmt.Println("Off")
		}
		fmt.Println("")
		// switch off by timer (in test variant)
		if cmd == 1 {
			go func() {
				time.Sleep(30 * time.Second)
				c.handle_motion(ed, 0)
			}()
		}
	} else if macAddress == 0x00124b0007246963 {
		// motion sensor in Custom3(children room)
		fmt.Print("motion sensor in Custom3(children room) ")
		if cmd == 1 {
			fmt.Println("On")
		} else {
			fmt.Println("Off")
		}
	}
}

// 0x01 On
// 0x02 Toggle
// 0x40 Off with effect
// 0x41 On with recall global scene
// 0x42 On with timed off  payload:0x00 (исполняем безусловно) 0x08 0x07(ON на 0x0708 (180,0)секунд) 0x00 0x00
func (c *Controller) onoff_command(ed *EndDevice, message zdo.Message) {
	fmt.Println("onoff_command")

	macAddress := ed.macAddress
	cmd := message.ZclFrame.Command
	state := "Off"
	if cmd == 1 {
		state = "On"
	}
	ed.set_current_state(state, 1)
	ts := time.Now() // get time now
	ed.set_last_action(ts)

	if macAddress == 0x8cf681fffe0656ef {
		// IKEA button on/off only
		//		c.get_power(ed)
		c.ikea_button_action(cmd)
	} else if macAddress == 0x0c4314fffe17d8a8 {
		// IKEA motion sensor
		//		c.get_power(ed)
		c.handle_motion(ed, 1) // switch off with timer
	} else if macAddress == 0x00124b0028928e8a {
		// button Sonoff1
		switch cmd { // 1 - double , 2 - single, 0 - long press
		case 0:
			ed.set_current_state("Long click", 1)
			c.switch_off_with_list()
		case 1:
			ed.set_current_state("Double click", 1)
		case 2:
			ed.set_current_state("Single click", 1)
		} //switch
	} else if macAddress == 0x00124b00253ba75f {
		// button Sonoff 2 call ringer with double click
		// switch off relays by list with long press
		switch cmd {
		case 0:
			ed.set_current_state("Long click", 1)
			c.switch_off_with_list()
		case 1:
			ed.set_current_state("Double click", 1)
			c.ringer()
			// alarm_msg := "Вызов с кнопки "
			// tlg32->send_message(alarm_msg);
		case 2:
			ed.set_current_state("Single click", 1)
		}
	}
}

// I have IKEA button only with this function
func (c *Controller) level_command(ed *EndDevice, message zdo.Message) {
	fmt.Println("level_command")
	ts := time.Now() // get time now
	ed.set_last_action(ts)

	cmd := message.ZclFrame.Command // 5 - Hold+, 7 - button realised, 1 - Hold-

	log.Printf("IKEA button level command: 0x%0x \n", cmd)

	switch cmd {
	case 1: // Hold
		ed.set_current_state("Minus down", 1)
		log.Printf("IKEA button: Hold-\n")
	case 5:
		ed.set_current_state("Plus down", 1)
		log.Printf("IKEA button: Hold+\n")
	case 7:
		ed.set_current_state("No act", 1)
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
func (c *Controller) ias_zone_command(cmnd uint8, shortAddr uint16) {
	executiveDevices := [3]uint64{
		0xa4c138d9758e1dcd,
		0xa4c138373e89d731,
		0x54ef441000193352}

	cmd := cmnd // automatically turn off only
	// enable/switch command is used only via web api
	if shortAddr > 0 {
		// one device, command via web api
		dev, res := Mapkey(c.devicessAddressMap, 0x54ef441000193352) // washing machine contactor relay
		if res && dev == shortAddr {
			cmd = 1 - cmnd
		}
		c.send_command_to_onoff_device(shortAddr, cmd, 1)
		log.Printf("Close device 0x%04x\n", shortAddr)
	} else {
		// all executive devices
		for _, macAddress := range executiveDevices {
			cmd1 := cmd
			if macAddress == 0x54ef441000193352 {
				cmd1 = 1 - cmd
			}
			_, res := Mapkey(c.devicessAddressMap, macAddress)
			if res {
				ts := time.Now() // get time now
				ed := c.get_device_by_mac(macAddress)
				if ed.shortAddress > 0 {
					ed.set_last_action(ts)

					// the Valves close slowly, we start everything in parallel threads
					go func(short uint16, cmd uint8) {
						c.send_command_to_onoff_device(short, cmd, 1)
					}(shortAddr, cmd1)
				}
			}
		} //for
	}
}

func (c *Controller) handle_sonoff_door(ed *EndDevice, cmd uint8) {
	ts := time.Now() // get time now
	ed.set_last_action(ts)

	state := "Closed"
	if ed.macAddress == 0x00124b002512a60b { // door sensor 2
		if cmd == 1 {
			state = "Opened"
		}
		ed.set_current_state(state, 1)
		// control relay 3, turn on the backlight in the bedroom in the cabinet
		c.switch_relay(0x54ef44100018b523, cmd, 1)
	} else if ed.macAddress == 0x00124b00250bba63 { // door sensor 3
		if cmd == 1 {
			state = "Opened"
		}
		ed.set_current_state(state, 1)
		// alarmMsg = "Закрыт ящик "
		// if cmd == 0x01 {
		// alarmMsg = "Открыт ящик "
		// }
		// tlg32->send_message(alarm_msg);
	} else if ed.macAddress == 0x00124b0025485ee6 { //door sensor 1, sensor in toilet
		if cmd == 0 {
			state = "Opened"
		}
		ed.set_current_state(state, 1)

		// light "Toilet occupied" in the corridor is turned on only from 9 am to 10 pm
		// turn off at any time (to cancel manual turn on)
		// enable/disable the relay 0x54ef4410005b2639 - light relay "Toilet occupied"
		if cmd == 0x01 {
			c.switch_relay(0x54ef4410005b2639, 0, 1)
		} else if cmd == 0 {
			h, _, _ := time.Now().Clock()
			if h > 7 && h < 23 {
				c.switch_relay(0x54ef4410005b2639, 1, 1)
			}
		}
	}
}

// IKEA button on/off action only
func (c *Controller) ikea_button_action(cmd uint8) {
	if cmd == 1 {
		fmt.Println("IKEA button on action")
	} else {
		fmt.Println("IKEA button off action")
	}
	fmt.Println("")
}

func (c *Controller) ringer() {
	if pi4.Pi4Available {
		pi4 := pi4.Pi4{}
		pi4.Ringer()
	}
}
