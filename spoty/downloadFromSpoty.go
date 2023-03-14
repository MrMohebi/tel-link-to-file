package spoty

import (
	"github.com/MrMohebi/tel-link-to-file/common"
	tele "gopkg.in/telebot.v3"
	"os"
	"os/exec"
	"time"
)

func SaveAndSend(link string, c tele.Context) error {
	err := c.Send("ای کلک میخوای از اسپاتیفای دانلود کنی؟\nالان میرم تو کارش... :)")
	common.IsErr(err)

	folderName := common.RandStr(5)
	cmdMkdir := exec.Command("mkdir", folderName)
	_, err = cmdMkdir.Output()

	cmd := exec.Command("spotdl", "download", link, "--output", folderName+"/{artist} - {title}.{output-ext}'")
	_, err = cmd.Output()

	entries, err := os.ReadDir(folderName)
	common.IsErr(err)

	for _, e := range entries {
		audio := &tele.Audio{File: tele.FromDisk(folderName + "/" + e.Name())}
		err := c.Send(audio, &tele.SendOptions{
			ReplyTo: c.Message(),
		})
		common.IsErr(err)
	}

	cmdRemoveDir := exec.Command("rm", " -rf", folderName)
	_, err = cmdRemoveDir.Output()

	if err != nil {
		time.Sleep(5 * time.Second)
		err := c.Send("اینو نمیتونم برات دانلود کنم برو سراغ یه آهنگ دیگه... :(")
		return err

	}

	return err
}
