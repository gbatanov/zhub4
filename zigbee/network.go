package zigbee

type NetworkConfiguration struct {
	pan_id            uint16      // = 0;
	extended_pan_id   uint64      // = 0;
	logical_type      LogicalType //= LogicalType_COORDINATOR
	channels          []uint8     //  really use default list - CH11
	precfg_key        [16]uint8   //{}
	precfg_key_enable bool        //= false; // value: 0 (FALSE) only coord defualtKey need to be set, and OTA to set other devices in the network.
}

var DefaultConfiguration NetworkConfiguration = NetworkConfiguration{0xFFFF, // Pan ID.
	0xDDDDDDDDDDDDDDDD,      // Extended pan ID.
	LogicalType_COORDINATOR, // Logical type.
	[]uint8{11},             // RF channel list.
	[16]uint8{0x01, 0x03, 0x05, 0x07, 0x09, 0x0B, 0x0D, 0x0F, 0x00, 0x02, 0x04, 0x06, 0x08, 0x0A, 0x0C, 0x0D}, // Precfg key.
	false}

var TestConfiguration NetworkConfiguration = NetworkConfiguration{0x1234, // Pan ID.
	0xDDDDDDDDDDDDDDDE,      // Extended pan ID.
	LogicalType_COORDINATOR, // Logical type.
	[]uint8{11},             // RF channel list.
	[16]uint8{0x01, 0x03, 0x05, 0x07, 0x09, 0x0B, 0x0D, 0x0F, 0x00, 0x02, 0x04, 0x06, 0x08, 0x0A, 0x0C, 0x0D}, // Precfg key.
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
	for i := 0; i < len(nc.precfg_key); i++ {
		if nc.precfg_key[i] != nc2.precfg_key[i] {
			return false
		}
	}
	return nc.precfg_key_enable == nc2.precfg_key_enable
}
