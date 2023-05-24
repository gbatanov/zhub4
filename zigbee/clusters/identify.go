/*
GSB, 2023
gbatanov@yandex.ru
*/
package clusters

import (
	"log"
	"zhub4/zigbee/zdo"
	"zhub4/zigbee/zdo/zcl"
)

type IdentifyCluster struct {
	Ed *zdo.EndDevice
}

func (i IdentifyCluster) Handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("IdentifyCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {
		log.Printf("Cluster::Identify: attribute id =0x%04x \n", attribute.Id)
		switch zcl.PowerConfigurationAttribute(attribute.Id) {

		default:
			log.Printf(" Cluster::Identify: unused attribute 0x%04x \n", attribute.Id)
		} //switch
	} //for
}
