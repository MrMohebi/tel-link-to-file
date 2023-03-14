package spoty

import (
	"github.com/MrMohebi/tel-link-to-file/common"
	"os/exec"
)

func DownloadAndSave(link string) {
	folderName := common.RandStr(5)
	cmdMkdir := exec.Command("mkdir", folderName)
	_, err := cmdMkdir.Output()

	cmd := exec.Command("/bin/sh", "-c", "spotdl", "download", link, "--output", "/root/"+folderName+"/{artist} - {title}.{output-ext}'")
	_, err = cmd.Output()
	common.IsErr(err)

}
