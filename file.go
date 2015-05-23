package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
)

const (
	_baseCmd          = "ffmpeg"
	_segmentsFileName = "segments.list"
	_chunkDuration    = 20
	_overwrite        = true
	_experimental     = true
	_videoCodec       = "h264"
	_audioCodec       = "aac"
	_container        = "mkv"
	_amazonUrl        = "https://s3-ap-southeast-2.amazonaws.com/flipendo/"
)

type File struct {
	filename  string
	id        string
	extension string
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
	fmt.Println("About to prepare for upload")
	nb := prepareForUpload(file)
	return nb
}

func (file *File) Concat(chunks int) {
	cmd, args := file.GetConcatCmd(chunks)
	err := exec.Command(cmd, args...).Run()
	if err != nil {
		fmt.Println("failure in concat")
		log.Fatal(err)
	}
	uploadFile("out/"+file.id+"."+_container, "merged"+"."+_container)
	msg, err := json.Marshal(map[string]interface{}{
		"action": "merged",
		"id":     file.id,
		"done":   true,
		"error":  false,
	})
	failOnError(err, "Failed to marshal message")
	publishToQueue(_apiQueueName, "text/json", msg)
}

func (file *File) Transcode(chunk string) {
	cmd, args := file.GetTranscodeCmd(chunk)
	err := exec.Command(cmd, args...).Run()
	if err != nil {
		fmt.Println("failure in transcode")
		log.Fatal(err)
	}
	uploadFile("chunks/"+file.id+"/out/"+chunk+"."+_container, chunk+"."+_container)
	msg, err := json.Marshal(map[string]interface{}{
		"action": "transcoded",
		"id":     file.id,
		"chunk":  chunk,
		"done":   true,
		"error":  false,
	})
	failOnError(err, "Failed to marshal message")
	publishToQueue(_apiQueueName, "text/json", msg)
}

func (*File) Upload() {

}

func (file *File) GetSplitCmd() (string, []string) {
	args := []string{}

	if _overwrite {
		args = append(args, "-y")
	}
	args = append(args, "-i")
	args = append(args, _amazonUrl+"files/"+file.filename)
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
	args = append(args, "%d"+file.extension)
	fmt.Println(args)

	return _baseCmd, args
}

func (file *File) GetTranscodeCmd(chunk string) (string, []string) {
	args := []string{}

	if _overwrite {
		args = append(args, "-y")
	}
	args = append(args, "-i")
	args = append(args, _amazonUrl+"chunks/"+file.id+"/"+chunk+file.extension)
	args = append(args, "-c:v")
	args = append(args, _videoCodec)
	args = append(args, "-c:a")
	args = append(args, _audioCodec)
	if _experimental {
		args = append(args, "-strict")
		args = append(args, "-2")
	}
	args = append(args, chunk+".mkv")
	fmt.Println(args)

	return _baseCmd, args
}

func (file *File) GetConcatCmd(chunks int) (string, []string) {
	list := getConcatList(file.id, chunks)

	args := []string{}

	if _overwrite {
		args = append(args, "-y")
	}
	args = append(args, "-f")
	args = append(args, "concat")
	args = append(args, "-i")
	args = append(args, list)
	args = append(args, "-c")
	args = append(args, "copy")
	args = append(args, "merged"+"."+_container)
	fmt.Println(args)

	return _baseCmd, args
}

func getConcatList(fileId string, chunks int) string {
	file, err := os.Create("files.list")
	failOnError(err, "Cannot create concat list")
	for i := 0; i < chunks; i++ {
		file.Write([]byte("file '" + _amazonUrl + "chunks/" + fileId + "/out/" + strconv.Itoa(i) + "." + _container + "'\n"))
	}
	return "files.list"
}
