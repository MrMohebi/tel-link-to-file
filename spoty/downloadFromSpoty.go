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
	if err != nil {
		return err
	}

	folderName := common.RandStr(5)
	cmdMkdir := exec.Command("mkdir", folderName)
	if _, err := cmdMkdir.CombinedOutput(); err != nil {
		return err
	}

	defer func() {
		cmdRemoveDir := exec.Command("rm", "-rf", folderName)
		_, _ = cmdRemoveDir.CombinedOutput()
	}()

	cmd := exec.Command("spotdl", "download", link, "--output", folderName+"/{artist} - {title}.{output-ext}'")
	if _, err := cmd.CombinedOutput(); err != nil {
		time.Sleep(5 * time.Second)
		_ = c.Send("اینو نمیتونم برات دانلود کنم برو سراغ یه آهنگ دیگه... :(")
		return err
	}

	entries, err := os.ReadDir(folderName)
	if err != nil {
		return err
	}

	for _, e := range entries {
		audio := &tele.Audio{File: tele.FromDisk(folderName + "/" + e.Name())}
		err = c.Send(audio, &tele.SendOptions{
			ReplyTo: c.Message(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
