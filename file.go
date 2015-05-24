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
	_amazonURL        = "https://s3-ap-southeast-2.amazonaws.com/flipendo/"
)

// File represents a file
type File struct {
	filename  string
	id        string
	extension string
}

// NewFile creates a new File
func NewFile(path string) *File {
	return &File{
		filename: path,
	}
}

// Split splits a file into chunks
func (file *File) Split() int {
	cmd, args := file.getSplitCmd()
	err := exec.Command(cmd, args...).Run()
	if err != nil {
		fmt.Println("failure in split")
		log.Fatal(err)
	}
	fmt.Println("About to prepare for upload")
	nb := prepareForUpload(file)
	return nb
}

// Concat merges chunks back into a single file
func (file *File) Concat(chunks int) {
	var msg []byte
	var err error

	message := map[string]interface{}{
		"action": "merged",
		"id":     file.id,
		"done":   true,
		"error":  nil,
	}

	cmd, args := file.getConcatCmd(chunks)
	if err = exec.Command(cmd, args...).Run(); err != nil {
		message["done"] = false
		message["error"] = "Could not merge file"
		if msg, err = json.Marshal(message); err != nil {
			log.Println("Could not marshal error message")
			return
		}
		publishToQueue(_apiQueueName, "text/json", msg)
		return
	}
	uploadFile("out/"+file.id+"."+_container, "merged"+"."+_container)
	msg, err = json.Marshal(message)
	failOnError(err, "Failed to marshal message")
	os.Remove(file.id)
	if err := publishToQueue(_apiQueueName, "text/json", msg); err != nil {
		log.Println("Failed to publish merged message")
	}
}

// Transcode transcodes a chunk into another format
func (file *File) Transcode(chunk string) {
	var msg []byte
	var err error

	log.Println("Transcoding chunk", chunk, "of file", file.id)

	message := map[string]interface{}{
		"action": "transcoded",
		"id":     file.id,
		"chunk":  chunk,
		"done":   true,
		"error":  nil,
	}
	os.Mkdir(file.id, 0777)
	cmd, args := file.getTranscodeCmd(chunk)
	err = exec.Command(cmd, args...).Run()
	if err != nil {
		message["done"] = false
		message["error"] = "Could not transcode chunk number" + chunk
		if msg, err = json.Marshal(message); err != nil {
			log.Println("Could not marshal error message")
			return
		}
		if err = publishToQueue(_apiQueueName, "text/json", msg); err != nil {
			log.Println("Failed to publish transcoding error message")
		}
		return
	}
	uploadFile("chunks/"+file.id+"/out/"+chunk+"."+_container, file.id+"/"+chunk+"."+_container)
	msg, err = json.Marshal(message)
	failOnError(err, "Failed to marshal message")
	if err = publishToQueue(_apiQueueName, "text/json", msg); err != nil {
		log.Println("Failed to publish transcoded message")
		return
	}
	log.Println("Transcoded chunk", chunk, "of file", file.id)
}

func (file *File) getSplitCmd() (string, []string) {
	var args []string

	if _overwrite {
		args = append(args, "-y")
	}
	args = append(args, "-i")
	args = append(args, _amazonURL+"files/"+file.filename)
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

	return _baseCmd, args
}

func (file *File) getTranscodeCmd(chunk string) (string, []string) {
	var args []string

	if _overwrite {
		args = append(args, "-y")
	}
	args = append(args, "-i")
	args = append(args, _amazonURL+"chunks/"+file.id+"/"+chunk+file.extension)
	args = append(args, "-c:v")
	args = append(args, _videoCodec)
	args = append(args, "-c:a")
	args = append(args, _audioCodec)
	if _experimental {
		args = append(args, "-strict")
		args = append(args, "-2")
	}
	args = append(args, file.id+"/"+chunk+".mkv")
	return _baseCmd, args
}

func (file *File) getConcatCmd(chunks int) (string, []string) {
	list := getConcatList(file.id, chunks)

	var args []string

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

	return _baseCmd, args
}

func getConcatList(fileID string, chunks int) string {
	file, err := os.Create("files.list")
	failOnError(err, "Cannot create concat list")
	for i := 0; i < chunks; i++ {
		file.Write([]byte("file '" + _amazonURL + "chunks/" + fileID + "/out/" + strconv.Itoa(i) + "." + _container + "'\n"))
	}
	return "files.list"
}
