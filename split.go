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

func getFileDuration() int {
	fullCmd := "ffprobe -i lily.mp4 -show_entries format=duration -v quiet -of csv=p=0"
	parts := strings.Fields(fullCmd)
	head := parts[0]
	tail := parts[1:len(parts)]
	fmt.Println(tail)
	out, err := exec.Command(head, tail...).Output()
	fmt.Printf("output is: %s\n", out)
	if err != nil {
		fmt.Println("failure in getFileDuration")
		log.Fatal(err)
	}
	s := string(out[0 : len(out)-1])
	fmt.Println(s)
	ret, error := strconv.ParseFloat(s, 64)
	if error != nil {
		fmt.Println("failure in getFileDuration conversion")
		log.Fatal(err)
	}
	return int(ret)
}

func getExecCmd(baseTime int) (string, []string) {
	args := []string{}

	if config.overwrite {
		args = append(args, "-y")
	}
	args = append(args, "-i")
	args = append(args, config.srcFileName)
	args = append(args, "-ss")
	args = append(args, strconv.Itoa(baseTime))
	args = append(args, "-t")
	args = append(args, strconv.Itoa(config.chunkDuration))
	args = append(args, "-c")
	args = append(args, "copy")
	if config.experimental {
		args = append(args, "-strict")
		args = append(args, "-2")
	}
	args = append(args, strings.Join([]string{"test",
		strconv.Itoa(baseTime),
		".mp4"}, ""))
	fmt.Println(args)

	return config.baseCmd, args
}

func split(srcFileName string) {
	duration := getFileDuration()

	for baseTime := 0; baseTime < duration; baseTime += config.chunkDuration {
		cmd, args := getExecCmd(baseTime)
		fmt.Println("Splitting file")
		out, err := exec.Command(cmd, args...).Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("exec output is: %s\nDuration is: %d\n", out, duration)
	}
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
