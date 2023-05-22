package zigbee

import "log"

type SimpleMeteringCluster struct {
}

func (s SimpleMeteringCluster) handler_attributes(endpoint Endpoint, attributes []Attribute) {
	log.Printf("SimpleMeteringCluster::endpoint address: 0x%04x number = %d \n", endpoint.address, endpoint.number)
	for _, attribute := range attributes {
		log.Printf("SimpleMeteringCluster unattended attribute id =0x%04x \n", attribute.id)
	}

}
