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

type IkeaCluster struct {
	Ed *zdo.EndDevice
}

func (i IkeaCluster) Handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("IkeaCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {
		log.Printf("IkeaCluster: attribute id =0x%04x \n", attribute.Id)
		switch zcl.PowerConfigurationAttribute(attribute.Id) {

		default:
			log.Printf("IkeaCluster: unused attribute 0x%04x \n", attribute.Id)
		} //switch
	} //for
}
