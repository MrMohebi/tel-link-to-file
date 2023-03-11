package main

import (
	"github.com/MrMohebi/tel-link-to-file/common"
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

	b.Handle(tele.OnText, func(c tele.Context) error {
		url := c.Text()

		extension := filepath.Ext(url)
		extension = strings.TrimLeft(extension, ".")

		kind := filetype.GetType(extension)

		println(kind.MIME.Value)

		if kind != types.Unknown && isAudioType(kind.MIME.Value) {
			c.Send("working on it... :)")
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

		return c.Send("please send an audio link")
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
