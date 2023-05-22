// GSB, 2023
// gbatanov@yandex.ru
package main

import (
	"bufio"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"zhub4/zigbee"

	"github.com/matishsiao/goInfo"
)

const Version string = "v0.1.4"

var Os string = ""
var Flag bool = true

func init() {
	fmt.Println("Init in main")
}

func main() {
	sysLog, err := syslog.New(syslog.LOG_INFO|syslog.LOG_SYSLOG, "zhub4")
	sysLog.Info("Start zhub4")

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

	getOsParams()
	var Ports map[string]string = map[string]string{
		"darwin":  "/dev/cu.usbmodem148201",
		"darwin2": "/dev/cu.usbserial-0001",
		"linux":   "/dev/ttyACM0",
		"linux2":  "/dev/ttyACM1"}

	zhub, err := zigbee.ZhubCreate(Ports, Os, "test")
	if err != nil {
		sysLog.Emerg(err.Error())
		log.Println(err)
		Flag = false
	}

	if Flag {
		zhub.Start()
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
		Flag = false
	}
}

func getOsParams() {
	gi, _ := goInfo.GetInfo()
	//	gi.VarDump()
	Os = gi.GoOS
}
