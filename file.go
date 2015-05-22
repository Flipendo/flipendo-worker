package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

const (
	_baseCmd          = "ffmpeg"
	_segmentsFileName = "segments.list"
	_chunkDuration    = 20
	_overwrite        = true
	_experimental     = false
	_videoCodec       = "h264"
	_audioCodec       = "aac"
	_amazonUrl        = "https://s3-ap-southeast-2.amazonaws.com/flipendo/files/"
)

type File struct {
	filename string
}

func NewFile(path string) *File {
	return &File{
		filename: path,
	}
}

func (file *File) Split() int {
	cmd, args := file.GetSplitCmd()
	err := exec.Command(cmd, args...).Run()
	if err != nil {
		fmt.Println("failure in split")
		log.Fatal(err)
	}
	nb := prepareForUpload()
	return nb
}

func (file *File) Concat() {
	cmd, args := file.GetConcatCmd()
	err := exec.Command(cmd, args...).Run()
	if err != nil {
		fmt.Println("failure in concat")
		log.Fatal(err)
	}
}

func (file *File) Transcode() {
	cmd, args := file.GetTranscodeCmd()
	err := exec.Command(cmd, args...).Run()
	if err != nil {
		fmt.Println("failure in transcode")
		log.Fatal(err)
	}
}

func (*File) Upload() {

}

func (file *File) GetSplitCmd() (string, []string) {
	args := []string{}

	if _overwrite {
		args = append(args, "-y")
	}
	args = append(args, "-i")
	args = append(args, _amazonUrl+file.filename)
	args = append(args, "-f")
	args = append(args, "segment")
	args = append(args, "-segment_time")
	args = append(args, strconv.Itoa(_chunkDuration))
	args = append(args, "-segment_list")
	args = append(args, _segmentsFileName)
	args = append(args, "-c")
	args = append(args, "copy")
	if _experimental {
		args = append(args, "-strict")
		args = append(args, "-2")
	}
	args = append(args, strings.Join([]string{"fromawstest",
		"%d",
		".mp4"}, ""))
	fmt.Println(args)

	return _baseCmd, args
}

func (file *File) GetTranscodeCmd() (string, []string) {
	args := []string{}

	args = append(args, "-i")
	args = append(args, file.filename)
	args = append(args, "-c:v")
	args = append(args, _videoCodec)
	args = append(args, "-c:a")
	args = append(args, _audioCodec)
	if _experimental {
		args = append(args, "-strict")
		args = append(args, "-2")
	}
	args = append(args, strings.Join([]string{"transcodedOutput",
		".mkv"}, ""))
	fmt.Println(args)

	return _baseCmd, args
}

func (file *File) GetConcatCmd() (string, []string) {
	args := []string{}

	if _overwrite {
		args = append(args, "-y")
	}
	args = append(args, "-f")
	args = append(args, "concat")
	args = append(args, "-i")
	args = append(args, file.filename)
	args = append(args, "-c")
	args = append(args, "copy")
	args = append(args, strings.Join([]string{"merged",
		".mkv"}, ""))
	fmt.Println(args)

	return _baseCmd, args
}
