package zigbee

import (
	"examples/comport/pi4"
	"fmt"
	"log"
	"time"
)

// scenarios
func (c *Controller) handle_motion(ed *EndDevice, cmd uint8) {

}

// 0x01 On
// 0x02 Toggle
// 0x40 Off with effect
// 0x41 On with recall global scene
// 0x42 On with timed off  payload:0x00 (исполняем безусловно) 0x08 0x07(ON на 0x0708 (180,0)секунд) 0x00 0x00
func (c *Controller) onoff_command(ed *EndDevice, message Message) {
	fmt.Println("onoff_command")

	macAddress := ed.macAddress
	cmd := message.zclFrame.Command
	state := "Off"
	if cmd == 1 {
		state = "On"
	}
	ed.set_current_state(state, 1)
	ts := time.Now() // get time now
	ed.set_last_action(ts)

	if macAddress == 0x8cf681fffe0656ef {
		// IKEA button on/off only
		c.get_power(ed)
		c.ikea_button_action(cmd)
	} else if macAddress == 0x0c4314fffe17d8a8 {
		// IKEA motion sensor
		c.get_power(ed)
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
func (c *Controller) level_command(ed *EndDevice, message Message) {
	fmt.Println("level_command")
	ts := time.Now() // get time now
	ed.set_last_action(ts)

	cmd1 := uint8(0)
	if len(message.zclFrame.Payload) > 0 {
		cmd1 = message.zclFrame.Payload[0]
	}
	cmd := message.zclFrame.Command // 5 - Hold+, 7 - button realised, 1 - Hold-

	log.Printf("IKEA button level command: %d %d \n", cmd1, cmd)

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
}

func (c *Controller) ringer() {
	if pi4.Pi4Available {
		pi4 := pi4.Pi4{}
		pi4.Ringer()
	}
}
