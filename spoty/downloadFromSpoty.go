package spoty

import (
	"github.com/MrMohebi/tel-link-to-file/common"
	"os/exec"
)

func DownloadAndSave(link string) {
	folderName := common.RandStr(5)
	cmd := exec.Command("spotdl", "download", link, "--output", "'./"+folderName+"/{artist} - {title}.{output-ext}'")
	_, err := cmd.Output()
	common.IsErr(err)

}
