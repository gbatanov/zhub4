/*
GSB, 2023
gbatanov@yandex.ru
*/
package clusters

import (
	"log"

	"github.com/gbatanov/zhub4/zigbee/zdo"

	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

type GroupsCluster struct {
	Ed *zdo.EndDevice
}

func (i GroupsCluster) Handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("GroupsCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {
		log.Printf("GroupsCluster: attribute id =0x%04x \n", attribute.Id)
		switch zcl.PowerConfigurationAttribute(attribute.Id) {

		default:
			log.Printf("GroupsCluster: unused attribute 0x%04x \n", attribute.Id)
		} //switch
	} //for
}
