package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var config struct {
	baseCmd      string
	srcFileName  string
	videoCodec   string
	audioCodec   string
	overwrite    bool
	experimental bool
}

func generateConfig(filename string) {
	config.baseCmd = "ffmpeg"
	config.srcFileName = filename
	config.videoCodec = "h264"
	config.audioCodec = "aac"
	config.overwrite = true
	config.experimental = true
}

func getExecCmd() (string, []string) {
	args := []string{}

	if config.overwrite {
		args = append(args, "-y")
	}
	args = append(args, "-f")
	args = append(args, "concat")
	args = append(args, "-i")
	args = append(args, config.srcFileName)
	args = append(args, "-c")
	args = append(args, "copy")
	args = append(args, strings.Join([]string{"merged",
		".mkv"}, ""))
	fmt.Println(args)

	return config.baseCmd, args
}

func concat() {
	cmd, args := getExecCmd()
	err := exec.Command(cmd, args...).Run()
	if err != nil {
		fmt.Println("failure in transcode")
		log.Fatal(err)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: [input filename]")
		return
	}
	generateConfig(os.Args[1])
	concat()
}
