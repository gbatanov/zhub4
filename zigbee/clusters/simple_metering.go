/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2023 GSB, Georgii Batanov gbatanov @ yandex.ru
*/

package clusters

import (
	"github.com/gbatanov/zhub4/zigbee/zdo"

	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

type SimpleMeteringCluster struct {
	Ed *zdo.EndDevice
}

func (s SimpleMeteringCluster) HandlerAttributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	//Подсчет суммарной потребленной мощности пока не использую
	/*
		//	log.Printf("SimpleMeteringCluster:: %s, endpoint address: 0x%04x number = %d \n", s.Ed.GetHumanName(), endpoint.Address, endpoint.Number)
		a0000 := false
		for _, attribute := range attributes {
			switch attribute.Id {
			case 0x0000: // CurrentSummationDelivered UINT48
				if !a0000 {
					a0000 = true
					var val64 []byte = append(attribute.Value, 0x0, 0x0)
					val := binary.LittleEndian.Uint64(val64)
					s.Ed.Set_energy(float64(val) / 1000)
					//CurrentSummationDelivered represents the most recent summed value of Energy, Gas, or Water delivered and consumed in the premises.
					log.Printf("SimpleMeteringCluster attribute id =0x%04x val = %0.2fkWh\n", attribute.Id, float64(val)/100)
				}
			default:
				log.Printf("SimpleMeteringCluster unattended attribute id =0x%04x \n", attribute.Id)
			}
		}
	*/
}
