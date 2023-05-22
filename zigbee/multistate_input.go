package zigbee

import "log"

type MultistateInputCluster struct {
}

func (m MultistateInputCluster) handler_attributes(endpoint Endpoint, attributes []Attribute) {
	log.Printf("MultistateInputCluster::endpoint address: 0x%04x number = %d \n", endpoint.address, endpoint.number)
	for _, attribute := range attributes {
		log.Printf("MULTISTATE_INPUT attribute id =0x%04x \n", attribute.id)
		switch MultiStateInputAttribute(attribute.id) {
		case MultiStateInput_000E,
			MultiStateInput_001C,
			MultiStateInput_004A,
			MultiStateInput_0051,
			MultiStateInput_0055,
			MultiStateInput_0067,
			MultiStateInput_006F,
			MultiStateInput_0100: // ApplicationType
			log.Printf("MULTISTATE_INPUT unattended attribute Id 0x%04x device: 0x%04x\n", attribute.id, endpoint.address)

		default:
			log.Printf("MULTISTATE_INPUT unknown attribute Id 0x%04x device: 0x%04x\n", attribute.id, endpoint.address)
		} //switch
	} //for
}
