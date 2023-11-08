/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2023 GSB, Georgii Batanov gbatanov @ yandex.ru
*/

package zdo

type RF_Channels struct {
	Channels []uint8
}

var DefaultRFChannels RF_Channels = RF_Channels{[]uint8{11}}
var TestRFChannels RF_Channels = RF_Channels{[]uint8{15}}

func (current RF_Channels) Compare(new RF_Channels) bool {
	if len(current.Channels) != len(new.Channels) {
		return false
	}
	for i := 0; i < len(current.Channels); i++ {
		if current.Channels[i] != new.Channels[i] {
			return false
		}
	}
	return true
}
