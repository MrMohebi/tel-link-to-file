package spoty

import (
	"github.com/MrMohebi/tel-link-to-file/common"
	tele "gopkg.in/telebot.v3"
	"os"
	"os/exec"
)

func SaveAndSend(link string, c tele.Context) error {
	folderName := common.RandStr(5)
	cmdMkdir := exec.Command("mkdir", folderName)
	_, err := cmdMkdir.Output()

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

	return err
}
