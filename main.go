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
	"log/syslog"
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

const Version string = "v0.3.21"

var Os string = ""
var Flag bool = true

func init() {
	fmt.Println("Init in main")
}

func main() {
	sysLog, err := syslog.New(syslog.LOG_INFO|syslog.LOG_SYSLOG, "zhub4")
	sysLog.Info("Start zhub4, version " + Version)

	if err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	//	intrpt := false // for gracefull exit
	// signal.Notify registers this channel to receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// This goroutine performs signal blocking.
	// When goroutine receives signal, it prints signal name out and then notifies the program that it can be terminated.
	go func() {
		sig := <-sigs
		log.Println(sig)
		Flag = false
		//		intrpt = true
	}()

	get_os_params()
	var Ports map[string]string = map[string]string{
		"darwin":  "/dev/cu.usbmodem148201",
		"darwin2": "/dev/cu.usbserial-0001",
		"linux":   "/dev/ttyACM0",
		"linux2":  "/dev/ttyACM1"}

	config, err := get_global_config()
	if err != nil {
		sysLog.Emerg(err.Error())
		log.Println(err)
		Flag = false
	}

	zhub, err := zigbee.Zhub_create(Ports, Os, config)
	if err != nil {
		sysLog.Emerg(err.Error())
		log.Println(err)
		Flag = false
	}

	if Flag {
		err = zhub.Start()
		if err == nil {
			defer zhub.Stop()
			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				for Flag {
					reader := bufio.NewReader(os.Stdin)
					text, _ := reader.ReadString('\n')
					if len(text) > 0 {
						switch []byte(text)[0] {
						case 'q':
							Flag = false
						case 'j':
							zhub.Get_controller().Get_zdo().Permit_join(60 * time.Second)
						} //switch
					}
				} //for
				wg.Done()
			}()
			wg.Wait()
		}
		Flag = false
	}
}

func get_os_params() {
	gi, _ := goInfo.GetInfo()
	//	gi.VarDump()
	Os = gi.GoOS
}

func get_global_config() (zigbee.GlobalConfig, error) {
	config := zigbee.GlobalConfig{}

	filename := "/usr/local/etc/zhub4/config"
	fd, err := os.OpenFile(filename, os.O_RDONLY, 0755)
	if err != nil {
		return zigbee.GlobalConfig{}, errors.New("incorrect file with configuration")
	} else {
		scan := bufio.NewScanner(fd)
		// read line by line
		for scan.Scan() {

			line := scan.Text()
			if strings.HasPrefix(line, "//") {
				continue
			}
			values := strings.Split(line, " ")
			if len(values) != 2 {
				return zigbee.GlobalConfig{}, errors.New("incorrect line")
			}

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
			case "MapPathTest":
				config.MapPathTest = values[1]
			case "MapPathProd":
				config.MapPathProd = values[1]
			case "Mode":
				config.Mode = strings.ToLower(values[1])
			default:
				return zigbee.GlobalConfig{}, errors.New("unknown parametr")
			}

		}
		fd.Close()
	}
	return config, nil
}
