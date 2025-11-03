package main

import (
	"github.com/MrMohebi/tel-link-to-file/common"
	telBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"github.com/kkdai/youtube/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gopkg.in/ini.v1"
	"io"
	"log"
	"net/http"
	"os"
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
				onStart(bot, chatId)
			} else {
				text := message.Message.Text
				isMP3Link := strings.Contains(text, ".mp3")
				isYTMLink := strings.Contains(text, "music.youtube.com")

				fileName := ""

				var audioFile io.ReadCloser

				if isYTMLink {
					_, err = bot.Send(telBot.NewMessage(chatId, "Downloading from YTM... :D"))
					audioFile, fileName = downloadFromYTM(text)

				} else if isMP3Link {
					_, err = bot.Send(telBot.NewMessage(chatId, "I'm working on it :D"))
					audioFile, fileName = downloadFromMP3Link(text)

				} else {
					_, err = bot.Send(telBot.NewMessage(chatId, "Invalid input!"))
					continue
				}

				if audioFile == nil {
					_, err = bot.Send(telBot.NewMessage(chatId, "Invalid input!"))
					continue
				}

				file := telBot.FileReader{
					Name:   fileName,
					Reader: audioFile,
				}

				audio := telBot.NewAudio(chatId, file)
				audio.ReplyToMessageID = message.Message.MessageID

				_, err = bot.Send(audio)
				if err != nil {
					common.IsErr(err, false)
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

func downloadFromMP3Link(link string) (io.ReadCloser, string) {

	extension := filepath.Base(link)
	name, _ := strings.CutSuffix(extension, "mp3")
	extension, _ = strings.CutPrefix(extension, name)
	name = strings.ReplaceAll(name, "%20", " ")
	name = strings.ReplaceAll(name, "+", " ")
	filePath := name + extension
	kind := filetype.GetType(extension)

	if kind != types.Unknown && isAudioType(kind.MIME.Value) {
		resp, err := http.Get(link)
		if err != nil {
			common.IsErr(err, false, "Failed to download the file.")
		}

		if resp.StatusCode != http.StatusOK {
			return nil, ""
		} else {
			return resp.Body, filePath
		}
	} else {
		return nil, ""
	}
}

func downloadFromYTM(link string) (io.ReadCloser, string) {
	client := youtube.Client{}

	video, err := client.GetVideo(link)
	formats := video.Formats.Type("audio")

	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		common.IsErr(err, false)
	}
	defer stream.Close()

	outputFile := video.Title + ".mp3"

	err = ffmpeg.Input("pipe:").
		Output(outputFile, ffmpeg.KwArgs{"c:a": "libmp3lame", "bitrate": "0", "f": "mp3", "vn": "", "metadata": "artist=" + video.Author}).WithInput(stream).
		OverWriteOutput().ErrorToStdOut().Run()

	audioFile, err := os.Open(outputFile)
	if err != nil {
		common.IsErr(err, false)
	}
	e := os.Remove(outputFile)
	if e != nil {
		common.IsErr(e, false)
	}

	return audioFile, outputFile
}

func onStart(bot *telBot.BotAPI, chatId int64) {
	_, err := bot.Send(telBot.NewMessage(chatId, "Send me the link."))
	if err != nil {
		common.IsErr(err, false, "Failed to send the message.")
	}
}
