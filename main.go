package main

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/MrMohebi/tel-link-to-file/common"
	telBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	aria "github.com/siku2/arigo"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gopkg.in/ini.v1"
)

var isIniInitOnce = false
var IniData *ini.File

type File struct {
    Reader io.ReadCloser
    Name string
}

/////////// nodemon --exec go run main.go --signal SIGTERM
// Arigo does NOT start aria2 WS at the moment. Run the following command to start aria2c: aria2c --enable-rpc --rpc-listen-all 

func main() {
    bot, err := telBot.NewBotAPI(IniGet("", "TOKEN"))
	if err != nil {
		common.IsErr(err, true)
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
                    files := downloadViaAria(text)
                    for _, file := range files {
				        file := telBot.FileReader{
				        	Name:   file.Name,
				        	Reader: file.Reader,
				        }
				        audio := telBot.NewAudio(chatId, file)
				        audio.ReplyToMessageID = message.Message.MessageID

				        _, err = bot.Send(audio)
				        if err != nil {
				        	common.IsErr(err, false)
				        }
                    }
                    continue
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

func startArigo ()*aria.Client {
	c, err := aria.Dial("ws://localhost:6800/jsonrpc", "")
	if err != nil {
        common.IsErr(err, false)
	}

    return c
}

func downloadViaAria(link string) []File { 
    c := startArigo()

    p, err := c.Download(aria.URIs(link), nil)
	if err != nil {
        common.IsErr(err, false)
	}

    if p.Status == aria.StatusCompleted {
        ariaFiles := p.Files
        files := []File{}

        if (len(ariaFiles) > 0) {
            for _, file := range ariaFiles {
                outputFile := file.Path
            	audioFile, err := os.Open(outputFile)
	            if err != nil {
                    common.IsErr(err, false)
	            }
	            err = os.Remove(outputFile)
	            if err != nil {
	            	common.IsErr(err, false)
	            }

                newFile := File{
                    Reader: audioFile,
                    Name:   outputFile,
                }

                files = append(files, newFile) 
            }
            return files
        }
    }
    return nil
}
