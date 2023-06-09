package telega32

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func init() {
	fmt.Println("init in telega 32")
}

type Message struct {
	ChatId int64
	Msg    string
}

// We inherit the bot to rewrite the function for receiving updates
type Tlg32 struct {
	botApi    *tgbotapi.BotAPI
	mode      string
	MyId      int64
	botName   string
	chatIds   []int64
	Flag      bool
	tokenPath string
	token     string
	MsgChan   chan Message
}

func Tlg32Create(botName string, mode string, tokenPath string, myId int64, msgChan chan Message) *Tlg32 {
	bot := Tlg32{}
	bot.mode = mode
	bot.tokenPath = tokenPath
	bot.botName = botName //your bot name
	bot.MyId = myId
	bot.chatIds = append(bot.chatIds, myId)
	bot.Flag = true
	bot.MsgChan = msgChan
	return &bot
}
func (bot *Tlg32) get_token() error {

	token, err := os.ReadFile("/usr/local/etc/telebot32/.token" + bot.botName + "/.token")
	if err != nil {
		return errors.New("incorrect file with token")
	}
	// remove trailing CR LF SPACE
	l := len(token)
	c := token[l-1]
	for c == 10 || c == 13 || c == 32 {
		token = token[:l-1]
		l = len(token)
		c = token[l-1]
	}
	bot.token = string(token)
	return nil

}
func (bot *Tlg32) Stop() {
	bot.Flag = false
}
func (bot *Tlg32) Run() error {
	var err error
	bot.get_token()
	bot.botApi, err = tgbotapi.NewBotAPI(string(bot.token))
	if err != nil {
		return errors.New("incorrect token")
	}
	bot.botApi.Debug = true

	fmt.Printf("Telebot authorized on account %s\n", bot.botApi.Self.UserName)

	go func() {
		bot.send_msg()
	}()

	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		for bot.Flag {
			updates := bot.My_get_updates_chan(u)
			for update := range updates {
				if !bot.Flag {
					bot.botApi.StopReceivingUpdates()
					return
				}

				if update.Message != nil { // If we got a message
					var outMsg string
					chatId := update.Message.Chat.ID
					msgIn := update.Message.Text
					firstName := update.Message.From.FirstName
					fmt.Printf("[%s] %s \n", update.Message.From.FirstName, update.Message.Text)
					outMsg, err := bot.handle_msg_in(msgIn, chatId, firstName)
					if err != nil {
						outMsg = "I'm understand you"
					}
					bot.MsgChan <- Message{ChatId: chatId, Msg: outMsg}
				}
			}
		}
	}()
	return nil
}

func (bot *Tlg32) send_msg() {
	// Этот код включаем, если нужно цитирование принятого сообщения
	//			msg.ReplyToMessageID = update.Message.MessageID
	for bot.Flag {
		inMsg := <-bot.MsgChan
		chatId := inMsg.ChatId
		text := inMsg.Msg

		msg := tgbotapi.NewMessage(chatId, text)
		bot.botApi.Send(msg)
	}
}

// GetUpdatesChan starts and returns a channel for getting updates.
func (bot *Tlg32) My_get_updates_chan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	ch := make(chan tgbotapi.Update, bot.botApi.Buffer)
	go func() {
		for bot.Flag {

			updates, err := bot.botApi.GetUpdates(config)
			if err != nil {
				log.Println(err)
				log.Println("Failed to get updates, terminate...")
				err_str := fmt.Sprintf("%s", err)
				if strings.Contains(err_str, "Conflict:") {
					bot.Flag = false
					close(ch)
					return
				} else {
					continue
				}
			}

			for _, update := range updates {
				if update.UpdateID >= config.Offset {
					config.Offset = update.UpdateID + 1
					ch <- update
				}
			}
		}
	}()

	return ch
}

// Обработчик входящих сообщений
func (bot *Tlg32) handle_msg_in(msg string, chatId int64, firstName string) (string, error) {
	// ID входит в список разрешенных
	found := false
	for _, acc := range bot.chatIds {
		if chatId == acc {
			found = true
			break
		}
	}
	// "/start" return "Привет, %s", message.from->first_name);
	if strings.Contains(msg, "/start") {
		return fmt.Sprintf("Привет, %s!", firstName), nil
	}
	if strings.Contains(msg, "/stop") {
		bot.Flag = false
		return fmt.Sprintf("Good bye, %s!", firstName), nil
	}

	// "/lwreg" add in chatIds
	if strings.Contains(msg, "/lwreg") {
		bot.chatIds = append(bot.chatIds, chatId)
		return "Ok", nil
	}

	if found {

		//"/lwunreg" remove from chatIds
		if strings.Contains(msg, "/lwreg") {
			for i, acc := range bot.chatIds {
				if acc == chatId {
					copy(bot.chatIds[i:], bot.chatIds[i+1:])
					bot.chatIds = bot.chatIds[:len(bot.chatIds)-1]
					break
				}
			}
			return "Ok", nil
		}

		// "/cmnd" проверяем команду на список допустимых, если Ок, отправляем на выполнение
		if strings.Contains(msg, "/cmnd") {
			return bot.handle_command(msg), nil
		}
		return "Ok", nil
	} else {
		return "Фиг вам!", errors.New("запрос не принят")
	}
}

// Пока жестко один пробел между параметрами команды
func (bot *Tlg32) handle_command(msg string) string {
	//	msg = strings.Replace(msg, "/cmnd ", "", 1)
	var c1 uint8
	var c uint8
	n, err := fmt.Sscanf(msg, "/cmnd %d %c", &c1, &c)
	fmt.Printf("%d: %d, %c \n", n, c1, c)

	if err != nil || n < 2 {
		return "Команда не выполнена"
	}
	//	c := msg[2]
	//	c1 := msg[0]
	switch c {
	case 'o': //
		bot.do_command(c, c1)
		return "Команда отправлена"
	default:
		return "Команда не выполнена"
	}

}

// команда
func (bot *Tlg32) do_command(c uint8, c1 uint8) {
	fmt.Printf("command %c %d \n", c, c1)
}
