/*
GSB, 2023
gbatanov@yandex.ru
*/
package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"zhub4/zigbee"

	"github.com/matishsiao/goInfo"
)

const Version string = "v0.5.38"

func init() {
	fmt.Println("Init in zhub")
}

func main() {

	var err error
	var Flag bool = true
	var controller *zigbee.Controller
	var config zigbee.GlobalConfig

	log.Println("Start zhub4, version " + Version)

	sigs := make(chan os.Signal, 1)
	//	intrpt := false // for gracefull exit
	// signal.Notify registers this channel to receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	// This goroutine performs signal blocking.
	// When goroutine receives signal, it prints signal name out and then notifies the program that it can be terminated.
	go func() {
		sig := <-sigs
		log.Println(sig)
		Flag = false
		//		intrpt = true
		//		controller.Stop()
	}()

	config, err = getGlobalConfig()
	if err != nil {
		log.Println(err.Error())
		Flag = false
	}

	controller, err = zigbee.Controller_create(&config)

	if err != nil {
		log.Println("1. ", err.Error())
		return
	}

	err = controller.Start_network()

	if err != nil {
		log.Println("2. ", err.Error())
		return
	}
	defer controller.Stop()

	var wg sync.WaitGroup

	wg.Add(1)
	//
	go func() {
		for Flag {
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			log.Println(text)
			if len(text) > 0 {
				switch []byte(text)[0] {
				case 'q':
					Flag = false
				case 'j':
					controller.Get_zdo().Permit_join(60 * time.Second)
				} //switch
			}
		} //for
		log.Println("flag false")
		wg.Done()
	}()
	wg.Wait()

	Flag = false

}

func getGlobalConfig() (zigbee.GlobalConfig, error) {
	config := zigbee.GlobalConfig{}
	gi, _ := goInfo.GetInfo()
	config.Os = gi.GoOS
	config.WithModem = false
	config.WithTlg = false
	filename := "/usr/local/etc/zhub4/config.txt"
	if config.Os == "windows" {
		filename = "C:\\work\\my\\zhub4\\config.txt"
	}
	fd, err := os.OpenFile(filename, os.O_RDONLY, 0755)
	if err != nil {
		return zigbee.GlobalConfig{}, errors.New("incorrect file with configuration")
	} else {
		scan := bufio.NewScanner(fd)
		var mode string = ""
		var sectionMode bool = true
		var values []string = []string{}
		// read line by line
		for scan.Scan() {

			line := scan.Text()
			line = strings.Trim(line, " \t")

			if strings.HasPrefix(line, "//") { //comment
				continue
			}
			if len(line) < 3 { //empty string
				continue
			}
			if len(mode) == 0 {
				values = strings.Split(line, " ")
				if values[0] != "Mode" {
					continue
				}
				mode = strings.Trim(line[len(values[0]):], " \t")
				mode = strings.ToLower(strings.Split(mode, " ")[0])
				config.Mode = mode
				continue
			}

			if strings.HasPrefix(line, "[") {
				section := line[1 : len(line)-1]
				sectionMode = section == mode
				continue
			}
			if !sectionMode { //pass section
				continue
			}
			values := strings.Split(line, " ")
			values1 := strings.Trim(line[len(values[0]):], " \t")
			values[1] = strings.Split(values1, " ")[0]

			switch values[0] {
			case "BotName":
				config.BotName = values[1]
			case "MyId":
				valInt, err := strconv.Atoi(values[1])
				if err != nil {
					return zigbee.GlobalConfig{}, errors.New("incorrect MyId")
				} else {
					config.MyId = int64(valInt)
				}
			case "TokenPath":
				config.TokenPath = values[1]
			case "MapPath":
				config.MapPath = values[1]
			case "Port":
				config.Port = values[1]
			case "ModemPort":
				config.ModemPort = values[1]
			case "Http":
				config.HttpAddress = values[1]
			case "Channels":
				config.Channels = make([]uint8, 0)
				channels := strings.Split(values[1], ",")
				for i := 0; i < len(channels); i++ {
					val, err := strconv.Atoi(channels[i])
					if err == nil {
						config.Channels = append(config.Channels, uint8(val))
					}
				}
				//			default: // pass
				//				return zigbee.GlobalConfig{}, errors.New("unknown parametr")
			}
		}
		fd.Close()
	}

	return config, nil
}
