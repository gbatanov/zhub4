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

type LightLinkCluster struct {
	Ed *zdo.EndDevice
}

func (i LightLinkCluster) Handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("LightLinkCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {
		log.Printf("LightLinkCluster: attribute id =0x%04x \n", attribute.Id)
		switch zcl.PowerConfigurationAttribute(attribute.Id) {

		default:
			log.Printf("LightLinkCluster: unused attribute 0x%04x \n", attribute.Id)
		} //switch
	} //for
}
