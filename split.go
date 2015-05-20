package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var config struct {
	baseCmd       string
	srcFileName   string
	chunkDuration int
	overwrite     bool
	experimental  bool
}

func generateConfig(filename string) {
	config.baseCmd = "ffmpeg"
	config.srcFileName = filename
	config.chunkDuration = 20
	config.overwrite = true
	config.experimental = true
}

func getExecCmd() (string, []string) {
	startTime := 0
	args := []string{}

	if config.overwrite {
		args = append(args, "-y")
	}
	args = append(args, "-i")
	args = append(args, config.srcFileName)
	args = append(args, "-ss")
	args = append(args, strconv.Itoa(startTime))
	args = append(args, "-t")
	args = append(args, strconv.Itoa(config.chunkDuration))
	args = append(args, "-c")
	args = append(args, "copy")
	if config.experimental {
		args = append(args, "-strict")
		args = append(args, "-2")
	}
	args = append(args, strings.Join([]string{"test",
		strconv.Itoa(startTime),
		".mp4"}, ""))
	fmt.Println(args)

	return config.baseCmd, args
}

func split(srcFileName string) {
	cmd, args := getExecCmd()
	fmt.Println("Splitting file")
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("exec output is: %s\n", out)
}

func main() {
	var srcFileName string

	if len(os.Args) != 2 {
		fmt.Println("Usage: [input filename]")
		return
	}
	generateConfig(os.Args[1])
	split(srcFileName)
	return
}
