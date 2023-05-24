package zcl

type GlobalCommands uint8

const (
	READ_ATTRIBUTES                       GlobalCommands = 0x00
	READ_ATTRIBUTES_RESPONSE              GlobalCommands = 0x01
	WRITE_ATTRIBUTES                      GlobalCommands = 0x02
	WRITE_ATTRIBUTES_UNDIVIDED            GlobalCommands = 0x03
	WRITE_ATTRIBUTES_RESPONSE             GlobalCommands = 0x04
	WRITE_ATTRIBUTES_NO_RESPONSE          GlobalCommands = 0x05
	CONFIGURE_REPORTING                   GlobalCommands = 0x06
	CONFIGURE_REPORTING_RESPONSE          GlobalCommands = 0x07
	READ_REPORTING_CONFIGURATION          GlobalCommands = 0x08
	READ_REPORTING_CONFIGURATION_RESPONSE GlobalCommands = 0x0
	REPORT_ATTRIBUTES                     GlobalCommands = 0x0a
	DEFAULT_RESPONSE                      GlobalCommands = 0x0b
	DISCOVER_ATTRIBUTES                   GlobalCommands = 0x0c
	DISCOVER_ATTRIBUTES_RESPONSE          GlobalCommands = 0x0d
	READ_ATTRIBUTES_STRUCTURED            GlobalCommands = 0x0e
	WRITE_ATTRIBUTES_STRUCTURED           GlobalCommands = 0x0f
	WRITE_ATTRIBUTES_STRUCTURED_RESPONSE  GlobalCommands = 0x10
	DISCOVER_COMMANDS_RECEIVED            GlobalCommands = 0x11
	DISCOVER_COMMANDS_RECEIVED_RESPONSE   GlobalCommands = 0x12
	DISCOVER_COMMANDS_GENERATED           GlobalCommands = 0x13
	DISCOVER_COMMANDS_GENERATED_RESPONSE  GlobalCommands = 0x14
	DISCOVER_ATTRIBUTES_EXTENDED          GlobalCommands = 0x15
	DISCOVER_ATTRIBUTES_EXTENDED_RESPONSE GlobalCommands = 0x16
)
