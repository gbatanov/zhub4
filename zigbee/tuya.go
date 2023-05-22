package zigbee

import "log"

type TuyaCluster struct {
}

// unattended clusters
// TUYA_ELECTRICIAN_PRIVATE_CLUSTER
// SmartPlug and WaterValve
func (b TuyaCluster) handler_attributes1(endpoint Endpoint, attributes []Attribute) {
	log.Printf("TuyaCluster::endpoint address: 0x%04x number = %d \n", endpoint.address, endpoint.number)
	for _, attribute := range attributes {
		log.Printf("attribute id =0x%04x \n", attribute.id)
	}

}

// TUYA_SWITCH_MODE_0
func (b TuyaCluster) handler_attributes2(endpoint Endpoint, attributes []Attribute) {
	log.Printf("TuyaCluster::endpoint address: 0x%04x number = %d \n", endpoint.address, endpoint.number)
	for _, attribute := range attributes {
		log.Printf("attribute id =0x%04x \n", attribute.id)
	}

}
