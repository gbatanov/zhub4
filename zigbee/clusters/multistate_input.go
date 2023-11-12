/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2023 GSB, Georgii Batanov gbatanov @ yandex.ru
*/
package clusters

import (
	"log"

	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

type MultistateInputCluster struct {
}

func (m MultistateInputCluster) HandlerAttributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	//	log.Printf("MultistateInputCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)
	for _, attribute := range attributes {
		//		log.Printf("MULTISTATE_INPUT attribute id =0x%04x \n", attribute.Id)
		switch zcl.MultiStateInputAttribute(attribute.Id) {
		case zcl.MultiStateInput_000E,
			zcl.MultiStateInput_001C,
			zcl.MultiStateInput_004A,
			zcl.MultiStateInput_0051,
			zcl.MultiStateInput_0055,
			zcl.MultiStateInput_0067,
			zcl.MultiStateInput_006F,
			zcl.MultiStateInput_0100: // ApplicationType
			log.Printf("MULTISTATE_INPUT unattended attribute Id 0x%04x device: 0x%04x\n", attribute.Id, endpoint.Address)

		default:
			log.Printf("MULTISTATE_INPUT unknown attribute Id 0x%04x device: 0x%04x\n", attribute.Id, endpoint.Address)
		} //switch
	} //for
}
