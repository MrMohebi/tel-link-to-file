package ytdlp

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func GetAudio(link string) (io.Reader, string, error) {
	folderName, err := os.MkdirTemp("", "yt-*")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp folder: %w", err)
	}
	defer os.RemoveAll(folderName)

	cmd := exec.Command(
		"yt-dlp",
		"-x",
		"--audio-format", "mp3",
		"--audio-quality", "10",
		"--no-overwrites",
		"--no-warnings",
		"--ignore-errors",
		"--cookies", "YTM_cookies.txt",
		"-o", folderName+"/%(artist)s - %(title)s - %(abr)s.%(ext)s",
		link,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, "", fmt.Errorf("yt-dlp error: %w, output: %s", err, string(output))
	}

	entries, err := os.ReadDir(folderName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read temp folder: %w", err)
	}

	if len(entries) == 0 {
		return nil, "", fmt.Errorf("no audio files downloaded")
	}

	var filePath string
	for _, e := range entries {
		if !e.IsDir() {
			filePath = filepath.Join(folderName, e.Name())
			break
		}
	}

	if filePath == "" {
		return nil, "", fmt.Errorf("no audio file found")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read audio file: %w", err)
	}

	return bytes.NewReader(data), filepath.Base(filePath), nil
}
