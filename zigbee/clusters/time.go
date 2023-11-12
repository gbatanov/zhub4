/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2023 GSB, Georgii Batanov gbatanov @ yandex.ru
*/

package clusters

import (
	"log"

	"github.com/gbatanov/zhub4/zigbee/zdo"

	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

type TimeCluster struct {
	Ed *zdo.EndDevice
}

func (i TimeCluster) HandlerAttributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	//	log.Printf("TimeCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {
		//		log.Printf("TimeCluster: attribute id =0x%04x \n", attribute.Id)
		switch zcl.PowerConfigurationAttribute(attribute.Id) {

		default:
			log.Printf("TimeCluster: unused attribute 0x%04x \n", attribute.Id)
		} //switch
	} //for
}
