package zigbee

import "log"

type SimpleMeteringCluster struct {
}

func (s SimpleMeteringCluster) handler_attributes(ep Endpoint, attributes []Attribute) {
	for _, attribute := range attributes {
		log.Printf("SimpleMeteringCluster unattended attribute id =0x%04x \n", attribute.id)
	}

}
