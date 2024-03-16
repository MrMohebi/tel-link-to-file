package main

import (
	"github.com/MrMohebi/tel-link-to-file/common"
	telBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"gopkg.in/ini.v1"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

var isIniInitOnce = false
var IniData *ini.File

/////////// nodemon --exec go run main.go --signal SIGTERM

func main() {
	bot, err := telBot.NewBotAPI(IniGet("", "TOKEN"))
	if err != nil {
		common.IsErr(err, true, "Failed to start the bot!")
	}

	u := telBot.NewUpdate(0)
	u.Timeout = 60

	message := bot.GetUpdatesChan(u)

	for message := range message {
		if message.Message != nil {
			log.Printf("[%s] %s", message.Message.From.UserName, message.Message.Text)

			chatId := message.Message.Chat.ID

			if message.Message.Text == "/start" {
				_, err = bot.Send(telBot.NewMessage(chatId, "Send me the link."))
				if err != nil {
					log.Print("Failed to send the message.")
				}
			} else {
				url := message.Message.Text

				extension := filepath.Base(url)
				name, extension, _ := strings.Cut(extension, ".")
				name = strings.ReplaceAll(name, "%20", " ")
				name = strings.ReplaceAll(name, "+", " ")
				filePath := name + "." + extension

				kind := filetype.GetType(extension)

				if kind != types.Unknown && isAudioType(kind.MIME.Value) {
					_, err = bot.Send(telBot.NewMessage(chatId, "I'm working on it :D"))

					resp, err := http.Get(url)
					if err != nil {
						common.IsErr(err, false, "Failed to download the file.")
						return
					}

					if resp.StatusCode != http.StatusOK {
						_, err = bot.Send(telBot.NewMessage(chatId, "404, Not Found"))
					} else {
						//defer func(Body io.ReadCloser) {
						//    err = Body.Close()
						//    common.IsErr(err, false)
						//}(resp.Body)

						file := telBot.FileReader{
							Name:   filePath,
							Reader: resp.Body,
						}

						audio := telBot.NewAudio(chatId, file)
						audio.ReplyToMessageID = message.Message.MessageID

						_, err = bot.Send(audio)
						if err != nil {
							common.IsErr(err, false)
							return
						}
					}

				} else {
					_, err = bot.Send(telBot.NewMessage(chatId, "Invalid input!"))
				}
			}
		}
	}
}

func IniSetup() {
	if !isIniInitOnce {
		var err error
		IniData, err = ini.Load("config.ini")
		common.IsErr(err, true, "Error loading .ini file")
		isIniInitOnce = true
	} else {
		println("initialized inis once")
	}
}

func IniGet(section string, key string) string {
	if IniData == nil {
		IniSetup()
	}
	return IniData.Section(section).Key(key).String()
}

func isAudioType(filetype string) bool {
	return filetype[:6] == "audio/"
}
