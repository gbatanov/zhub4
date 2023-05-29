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
	"zhub4/http_server"
	"zhub4/zigbee"

	"github.com/matishsiao/goInfo"
)

const Version string = "v0.4.28"

var Flag bool = true

type Zhub struct {
	controller *zigbee.Controller
	Flag       bool
	config     zigbee.GlobalConfig
	withHttp   bool
	web        *http_server.HttpServer
}

func init() {
	fmt.Println("Init in  zhub")
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

	config, err := get_global_config()
	if err != nil {
		sysLog.Emerg(err.Error())
		log.Println(err)
		Flag = false
	}

	zhub, err := Zhub_create(config)
	if err != nil {
		sysLog.Emerg(err.Error())
		log.Println(err)
		Flag = false
	}
	zhub.web, err = http_server.Http_server_create()
	zhub.withHttp = err == nil

	if zhub.withHttp {
		err = zhub.web.Start()
		zhub.withHttp = err == nil
	}
	if zhub.withHttp {
		fmt.Println("Web server started")
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

func get_global_config() (zigbee.GlobalConfig, error) {
	config := zigbee.GlobalConfig{}
	gi, _ := goInfo.GetInfo()
	config.Os = gi.GoOS

	filename := "/usr/local/etc/zhub4/config"
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

func Zhub_create(config zigbee.GlobalConfig) (*Zhub, error) {
	controller, err := zigbee.Controller_create(config)
	if err != nil {
		return &Zhub{}, err
	}
	zhub := Zhub{controller: controller, Flag: false, config: config}
	return &zhub, nil
}

func (zhub *Zhub) Start() error {
	zhub.Flag = true

	err := zhub.controller.Start_network()
	if err != nil {
		log.Fatal(err)
	}

	return nil

}

func (zhub *Zhub) Stop() {
	zhub.controller.Stop()
	if zhub.withHttp {
		zhub.web.Stop()
	}
}

func (zhub *Zhub) Get_controller() *zigbee.Controller {
	return zhub.controller
}
