package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"
)

var config struct {
	baseCmd          string
	srcFileName      string
	segmentsFileName string
	chunkDuration    int
	overwrite        bool
	experimental     bool
}

func generateConfig(filename string) {
	config.baseCmd = "ffmpeg"
	config.srcFileName = filename
	config.segmentsFileName = "segments.list"
	config.chunkDuration = 20
	config.overwrite = true
	config.experimental = true
}

func AwsInit() (*s3.Bucket, error) {
	auth, err := aws.GetAuth("", "", "", time.Time{})
	if err != nil {
		return nil, err
	}
	client := s3.New(auth, aws.APSoutheast2)

	return client.Bucket("flipendo"), nil
}

func getFileContent(filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	var data []byte
	total := 0
	buf := make([]byte, 1024)
	count, error := file.Read(buf)
	for error != io.EOF {
		total += count
		if error != nil {
			log.Fatal(error)
		}
		data = append(data, buf[:count]...)
		count, error = file.Read(buf)
	}
	return data
}

func checkForNewSegment() {

}

func uploadAsync() {
	file, err := os.Open(config.segmentsFileName)
	if err != nil {
		log.Fatal(err)
	}

	bucket, error := AwsInit()
	if error != nil {
		log.Fatal(error)
	}

	scanner := bufio.NewScanner(file)
	var wg sync.WaitGroup
	for scanner.Scan() {
		filename := scanner.Text()
		wg.Add(1)
		go func() {
			error = bucket.Put(filename, getFileContent(filename), "content-type", s3.Private, s3.Options{})
			if error != nil {
				log.Fatal(error)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func getFileDuration() int {
	fullCmd := strings.Join([]string{"ffprobe -i ", config.srcFileName, " -show_entries format=duration -v quiet -of csv=p=0"}, "")
	parts := strings.Fields(fullCmd)
	head := parts[0]
	tail := parts[1:len(parts)]
	fmt.Println(head)
	fmt.Println(tail)
	out, err := exec.Command(head, tail...).Output()
	if err != nil {
		log.Fatal(err)
	}
	s := string(out[0 : len(out)-1])
	fmt.Println(s)
	ret, error := strconv.ParseFloat(s, 64)
	if error != nil {
		log.Fatal(err)
	}
	return int(ret)
}

func getExecCmd() (string, []string) {
	args := []string{}

	if config.overwrite {
		args = append(args, "-y")
	}
	args = append(args, "-i")
	args = append(args, config.srcFileName)
	args = append(args, "-f")
	args = append(args, "segment")
	args = append(args, "-segment_time")
	args = append(args, strconv.Itoa(config.chunkDuration))
	args = append(args, "-segment_list")
	args = append(args, config.segmentsFileName)
	args = append(args, "-c")
	args = append(args, "copy")
	if config.experimental {
		args = append(args, "-strict")
		args = append(args, "-2")
	}
	args = append(args, strings.Join([]string{"fromawstest",
		"%d",
		".mp4"}, ""))
	fmt.Println(args)

	return config.baseCmd, args
}

func split(srcFileName string) {
	cmd, args := getExecCmd()
	err := exec.Command(cmd, args...).Run()
	if err != nil {
		fmt.Println("failure in split")
		log.Fatal(err)
	}
	uploadAsync()
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
