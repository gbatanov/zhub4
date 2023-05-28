/*
GSB, 2023
gbatanov@yandex.ru
*/
package zigbee

import (
	"fmt"
	"log"
)

type Zhub struct {
	controller *Controller
	Flag       bool
	config     GlobalConfig
}

func init() {
	fmt.Println("Init in zigbee: zhub")
}

func Zhub_create(config GlobalConfig) (*Zhub, error) {
	controller, err := controller_create(config)
	if err != nil {
		return &Zhub{}, err
	}
	zhub := Zhub{controller: controller, Flag: false, config: config}
	return &zhub, nil
}

func (zhub *Zhub) Start() error {
	zhub.Flag = true

	err := zhub.controller.start_network()
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
