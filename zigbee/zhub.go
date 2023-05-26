package zigbee

import (
	"fmt"
	"log"
	"zhub4/zigbee/zdo"
)

type Zhub struct {
	controller *Controller
	mode       string
	Flag       bool
}

func init() {
	fmt.Println("Init in zigbee: zhub")
}

func Zhub_create(Ports map[string]string, Os string, mode string) (*Zhub, error) {
	controller, err := controller_create(Ports, Os, mode)
	if err != nil {
		return &Zhub{}, err
	}
	zhub := Zhub{controller: controller, Flag: false, mode: mode}
	return &zhub, nil
}

func (zhub *Zhub) Start() error {
	zhub.Flag = true
	var err error
	if zhub.mode == "prod" {
		err = zhub.controller.start_network(zdo.DefaultRFChannels)
	} else {
		err = zhub.controller.start_network(zdo.TestRFChannels)
	}
	if err != nil {
		log.Fatal(err)
	}

	return nil

}

func (zhub *Zhub) Stop() {
	zhub.controller.Stop()
}

func (zhub *Zhub) Get_controller() *Controller {
	return zhub.controller
}
