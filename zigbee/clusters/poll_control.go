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

type PollControlCluster struct {
	Ed *zdo.EndDevice
}

func (i PollControlCluster) Handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("PollControlCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {
		log.Printf("PollControlCluster: attribute id =0x%04x \n", attribute.Id)
		switch zcl.PowerConfigurationAttribute(attribute.Id) {

		default:
			log.Printf("PollControlCluster: unused attribute 0x%04x \n", attribute.Id)
		} //switch
	} //for
}
