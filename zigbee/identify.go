package zigbee

import (
	"log"
)

type IdentifyCluster struct {
	ed *EndDevice
}

func (i IdentifyCluster) handler_attributes(endpoint Endpoint, attributes []Attribute) {
	for _, attribute := range attributes {
		log.Printf("Cluster::Identify: attribute id =0x%04x \n", attribute.id)
		switch PowerConfigurationAttribute(attribute.id) {

		default:
			log.Printf(" Cluster::Identify: unused attribute 0x%04x \n", attribute.id)
		} //switch
	} //for
}
