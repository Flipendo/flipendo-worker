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
	experimental bool
}

func generateConfig(filename string) {
	config.baseCmd = "ffmpeg"
	config.srcFileName = filename
	config.videoCodec = "h264"
	config.audioCodec = "aac"
	config.experimental = true
}

func getExecCmd() (string, []string) {
	args := []string{}

	args = append(args, "-i")
	args = append(args, config.srcFileName)
	args = append(args, "-c:v")
	args = append(args, config.videoCodec)
	args = append(args, "-c:a")
	args = append(args, config.audioCodec)
	if config.experimental {
		args = append(args, "-strict")
		args = append(args, "-2")
	}
	args = append(args, strings.Join([]string{"transcodedOutput",
		".mkv"}, ""))
	fmt.Println(args)

	return config.baseCmd, args
}

func transcode() {
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
	transcode()
}
