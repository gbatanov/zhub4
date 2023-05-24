package zigbee

import (
	"log"
	"zhub4/zigbee/zdo/zcl"
)

type SimpleMeteringCluster struct {
}

func (s SimpleMeteringCluster) handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("SimpleMeteringCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)
	for _, attribute := range attributes {
		log.Printf("SimpleMeteringCluster unattended attribute id =0x%04x \n", attribute.Id)
	}

}
