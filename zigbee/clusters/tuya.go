/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2023 GSB, Georgii Batanov gbatanov @ yandex.ru
*/
package clusters

import (
	"log"

	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

type TuyaCluster struct {
}

// unattended clusters
// TUYA_ELECTRICIAN_PRIVATE_CLUSTER
// SmartPlug and WaterValve
func (b TuyaCluster) HandlerAttributes1(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("TuyaCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)
	for _, attribute := range attributes {
		log.Printf("attribute id =0x%04x \n", attribute.Id)
	}

}

// TUYA_SWITCH_MODE_0
func (b TuyaCluster) HandlerAttributes2(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("TuyaCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)
	for _, attribute := range attributes {
		log.Printf("attribute id =0x%04x \n", attribute.Id)
	}

}
