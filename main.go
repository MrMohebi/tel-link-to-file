package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/MrMohebi/tel-link-to-file/common"
	telBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"github.com/kkdai/youtube/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gopkg.in/ini.v1"
)

var isIniInitOnce = false
var IniData *ini.File

/////////// nodemon --exec go run main.go --signal SIGTERM

func main() {
    var link string

    fmt.Println("Enter the link:")
    fmt.Scan(&link)

    x, y := downloadFromMP3Link(link)
    fmt.Println(x, y)
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
