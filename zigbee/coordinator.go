package zigbee

import (
	"fmt"
	"log"
)

type Coordinator struct {
	controller *Controller
	mode       string
	Flag       bool
}

func init() {
	fmt.Println("Init in zigbee: coordinator")
}

func CoordinatorCreate(Ports map[string]string, Os string, mode string) (*Coordinator, error) {
	controller, err := controllerCreate(Ports, Os, mode)
	if err != nil {
		return &Coordinator{}, err
	}
	coordinator := Coordinator{controller: controller, Flag: false, mode: mode}
	return &coordinator, nil
}

func (coordinator *Coordinator) Start() error {
	coordinator.Flag = true
	var err error
	if coordinator.mode == "prod" {
		err = coordinator.controller.startNetwork(DefaultConfiguration)
	} else {
		err = coordinator.controller.startNetwork(TestConfiguration)
	}
	if err != nil {
		log.Fatal(err)
	}

	return nil

}

func (coordinator *Coordinator) Stop() {
	coordinator.controller.Stop()
}

func (coordinator *Coordinator) Get_controller() *Controller {
	return coordinator.controller
}
