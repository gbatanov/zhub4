package zigbee

import (
	"log"
	"zhub4/zigbee/zcl"
)

type IdentifyCluster struct {
	ed *EndDevice
}

func (i IdentifyCluster) handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("IdentifyCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {
		log.Printf("Cluster::Identify: attribute id =0x%04x \n", attribute.Id)
		switch zcl.PowerConfigurationAttribute(attribute.Id) {

		default:
			log.Printf(" Cluster::Identify: unused attribute 0x%04x \n", attribute.Id)
		} //switch
	} //for
}
