package main

import (
	"bytes"
	"fmt"
	"github.com/MrMohebi/tel-link-to-file/common"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-resty/resty/v2"
	telBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gopkg.in/ini.v1"
	"io"
	"log"
	"net/url"
	"os"
	"path"
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
					audioFile, fileName, err = DownloadFromLink(text)
					if err != nil {
						common.IsErr(err, false)
						_, err = bot.Send(telBot.NewMessage(chatId, "Invalid input!"))
					}

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

func DownloadFromLink(link string) (io.ReadCloser, string, error) {
	client := resty.New()

	resp, err := client.R().SetDoNotParseResponse(true).Get(link)
	if err != nil {
		return nil, "", err
	}
	defer resp.RawBody().Close()

	data, err := io.ReadAll(resp.RawBody())
	if err != nil {
		return nil, "", err
	}

	mime := mimetype.Detect(data)
	if mime == nil || !strings.HasPrefix(mime.String(), "audio/") {
		return nil, "", fmt.Errorf("not an audio file (detected: %s)", mime.String())
	}

	fileName := extractFileName(resp, link)

	reader := io.NopCloser(bytes.NewReader(data))
	return reader, fileName, nil
}

func extractFileName(resp *resty.Response, fileURL string) string {
	cd := resp.Header().Get("Content-Disposition")
	if cd != "" {
		if parts := strings.Split(cd, "filename="); len(parts) > 1 {
			name := strings.Trim(parts[1], "\"; ")
			if name != "" {
				return name
			}
		}
	}

	u, err := url.Parse(fileURL)
	if err == nil {
		name := path.Base(u.Path)
		if name != "" && name != "/" {
			return name
		}
	}

	return "unknown_audio_file"
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
