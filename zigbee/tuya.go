package zigbee

import (
	"log"
	"zhub4/zigbee/zcl"
)

type TuyaCluster struct {
}

// unattended clusters
// TUYA_ELECTRICIAN_PRIVATE_CLUSTER
// SmartPlug and WaterValve
func (b TuyaCluster) handler_attributes1(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("TuyaCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)
	for _, attribute := range attributes {
		log.Printf("attribute id =0x%04x \n", attribute.Id)
	}

}

// TUYA_SWITCH_MODE_0
func (b TuyaCluster) handler_attributes2(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("TuyaCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)
	for _, attribute := range attributes {
		log.Printf("attribute id =0x%04x \n", attribute.Id)
	}

}
