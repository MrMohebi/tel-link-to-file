package spoty

import (
	"fmt"
	"github.com/MrMohebi/tel-link-to-file/common"
	"os"
	"os/exec"
)

func SaveAndSend(link string) error {
	folderName := common.RandStr(5)
	cmdMkdir := exec.Command("mkdir", folderName)
	_, err := cmdMkdir.Output()

	cmd := exec.Command("spotdl", "download", link, "--output", folderName+"/{artist} - {title}.{output-ext}'")
	_, err = cmd.Output()

	entries, err := os.ReadDir(folderName)
	common.IsErr(err)

	for _, e := range entries {
		fmt.Println(e.Name())
	}

	return err
}
