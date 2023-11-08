/*
GSB, 2023
gbatanov@yandex.ru
*/
package zigbee

import (
	"sync"
	"time"

	"github.com/gbatanov/gsm/modem"
	"github.com/gbatanov/zhub4/http_server"
	"github.com/gbatanov/zhub4/telega32"
	"github.com/gbatanov/zhub4/zigbee/clusters"
	"github.com/gbatanov/zhub4/zigbee/zdo"
)

type GlobalConfig struct {
	// telegram bot
	BotName   string
	MyId      int64
	TokenPath string
	// map short address to mac address
	MapPath string
	// working mode
	Mode string
	// channels
	Channels []uint8
	// serial port - zigbee adapter
	Port string
	// serial port - modem adapter
	ModemPort string
	// operating system
	Os string
	// HTTP server
	HttpAddress string
	WithTlg     bool
	WithModem   bool
	// Мой номер телефона
	MyPhoneNumber string
}
type TlgBlock struct {
	tlg32 *telega32.Tlg32
	//	withTlg    bool
	tlgMsgChan chan telega32.Message
}
type HttpBlock struct {
	answerChan chan string
	queryChan  chan string
	withHttp   bool
	web        *http_server.HttpServer
}

type Controller struct {
	zdobj              *zdo.Zdo
	config             *GlobalConfig
	devices            map[uint64]*zdo.EndDevice
	devicessAddressMap map[uint16]uint64
	flag               bool
	msgChan            chan zdo.Command        // chanel for receive incoming message command from zdo
	joinChan           chan []byte             // chanel for receive command join device from zdo
	motionMsgChan      chan clusters.MotionMsg // chanel for get message from motion sensors
	lastMotion         time.Time               // last motion any motion sensor
	smartPlugTS        time.Time               // timestamp for smart plug timer
	switchOffTS        bool                    // flag for switch off timer
	mapFileMutex       sync.Mutex
	tlg                TlgBlock
	http               HttpBlock
	startTime          time.Time
	mdm                *modem.GsmModem
}
