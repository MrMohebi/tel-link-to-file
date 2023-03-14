package main

import (
	"github.com/MrMohebi/tel-link-to-file/common"
	"github.com/MrMohebi/tel-link-to-file/spoty"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"gopkg.in/ini.v1"
	tele "gopkg.in/telebot.v3"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

var isIniInitOnce = false
var IniData *ini.File

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

		extension := filepath.Ext(url)
		extension = strings.TrimLeft(extension, ".")

		kind := filetype.GetType(extension)

		println(kind.MIME.Value)

		if kind != types.Unknown && isAudioType(kind.MIME.Value) {
			err := c.Send("دو دقه بل الان برات دانلودش میکنم... ")
			common.IsErr(err)

			resp, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			audio := &tele.Audio{File: tele.FromReader(resp.Body)}

			return c.Send(audio, &tele.SendOptions{
				ReplyTo: c.Message(),
			})
		}

		return c.Send("میدونی که باید برام لینک بفرستی درسته؟!")
	})

	b.Start()

}

func IniSetup() {
	if !isIniInitOnce {
		var err error
		IniData, err = ini.Load("config.ini")
		common.IsErr(err, "Error loading .ini file")
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
