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

// superfluous, doesn't use
/*
type NetworkConfiguration struct {
	pan_id            uint16          // = 0;
	extended_pan_id   uint64          // = 0;
	logical_type      zcl.LogicalType //= LogicalType_COORDINATOR
	channels          []uint8         //  really use default list - CH11
	precfg_key        [16]uint8       //{}
	precfg_key_enable bool            //= false; // value: 0 (FALSE) only coord defualtKey need to be set, and OTA to set other devices in the network.
}

var DefaultConfiguration NetworkConfiguration = NetworkConfiguration{0xFFFF, // Pan ID.
	0xDDDDDDDDDDDDDDDD,          // Extended pan ID. (mac address of coordinator)
	zcl.LogicalType_COORDINATOR, // Logical type.
	[]uint8{11},                 // RF channel list.
	[16]uint8{},                 // Precfg key.
	false}

var TestConfiguration NetworkConfiguration = NetworkConfiguration{0x1a62, // Pan ID.
	0xDDDDDDDDDDDDDDDE,          // Extended pan ID.(mac address of coordinator)
	zcl.LogicalType_COORDINATOR, // Logical type.
	[]uint8{15},                 // RF channel list.
	[16]uint8{},                 // Precfg key.
	false}

func (nc NetworkConfiguration) Compare(nc2 NetworkConfiguration) bool {
	if nc.pan_id != nc2.pan_id {
		return false
	}
	if nc.extended_pan_id != nc2.extended_pan_id {
		return false
	}
	if nc.logical_type != nc2.logical_type {
		return false
	}
	if len(nc.channels) != len(nc2.channels) {
		return false
	}

	for i := 0; i < len(nc.channels); i++ {
		if nc.channels[i] != nc2.channels[i] {
			return false
		}
	}
	if nc2.precfg_key_enable {
		for i := 0; i < len(nc.precfg_key); i++ {
			if nc.precfg_key[i] != nc2.precfg_key[i] {
				return false
			}
		}
	}
	return nc.precfg_key_enable == nc2.precfg_key_enable
}
*/
