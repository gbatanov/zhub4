package zdo

type RF_Channels struct {
	channels []uint8
}

var DefaultRFChannels RF_Channels = RF_Channels{[]uint8{11}}
var TestRFChannels RF_Channels = RF_Channels{[]uint8{15}}

func (current RF_Channels) Compare(new RF_Channels) bool {
	if len(current.channels) != len(new.channels) {
		return false
	}
	for i := 0; i < len(current.channels); i++ {
		if current.channels[i] != new.channels[i] {
			return false
		}
	}
	return true
}
