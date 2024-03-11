package main

import (
	"github.com/MrMohebi/tel-link-to-file/common"
	"github.com/MrMohebi/tel-link-to-file/spoty"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"gopkg.in/ini.v1"
	tele "gopkg.in/telebot.v3"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var isIniInitOnce = false
var IniData *ini.File

/////////// nodemon --exec go run main.go --signal SIGTERM

func main() {
	pref := tele.Settings{
		Token:  IniGet("", "TOKEN"),
		Poller: &tele.LongPoller{Timeout: 60 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/start", func(c tele.Context) error {
		welcomeMsg := c.Send("کافیه که فقط لینک مستقیم آهنگ یا لینک اسپاتیفای رو برام بفرستی و منم آهنگی که دنبالشی رو برات میفرستم.\nیادت نره که سلام مارو به همونی که داری براش آهنگ دانلود میکنی برسونی... XD ")
		return welcomeMsg
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		url := c.Text()

		println(url)

		if len(url) > 27 && url[:24] == "https://open.spotify.com" {
			return spoty.SaveAndSend(url, c)
		}

		extension := filepath.Base(url)
		name, extension, _ := strings.Cut(extension, ".")
		extension = strings.TrimLeft(extension, ".")
		name = strings.ReplaceAll(name, "%20", " ")
		name = strings.ReplaceAll(name, "+", " ")
		filePath := name + "." + extension
		println(filePath)

		kind := filetype.GetType(extension)

		println(kind.MIME.Value)

		if kind != types.Unknown && isAudioType(kind.MIME.Value) {
			err = c.Send("دو دقه بل الان برات دانلودش میکنم... ")
			common.IsErr(err, true)

			resp, err := http.Get(url)
			common.IsErr(err, true)
			if resp.StatusCode != http.StatusOK {
				return c.Send("این لینک رو نمیتونم جایی پیدا کنم... ):")
			}

			defer func(Body io.ReadCloser) {
				err = Body.Close()
				common.IsErr(err, true)
			}(resp.Body)

			output, err := os.Create(filePath)

			if err != nil {
				return err
			}
			defer func(output *os.File) {
				err := output.Close()
				if err != nil {
					common.IsErr(err, true)
				}
			}(output)
			_, err = io.Copy(output, resp.Body)

			audio := &tele.Audio{File: tele.FromDisk(filePath)}

			err = c.Send(audio, &tele.SendOptions{
				ReplyTo: c.Message(),
			})
			if err != nil {
				common.IsErr(err, true, "Failed to send the file.")
			}

			err = os.Remove(filePath)
			if err != nil {
				common.IsErr(err, true, "Failed to remove the file.")
			}
			return err
		}

		return c.Send("میدونی که باید برام لینک بفرستی درسته؟!")
	})

	b.Start()

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
