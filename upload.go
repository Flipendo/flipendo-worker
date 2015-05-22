package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"
)

var S3Instance struct {
	bucket *s3.Bucket
}

func awsInit() error {
	auth, err := aws.GetAuth("", "", "", time.Time{})
	if err != nil {
		return err
	}
	client := s3.New(auth, aws.APSoutheast2)

	S3Instance.bucket = client.Bucket("flipendo")
	return nil
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

func uploadFile(dest string, filename string) {
	err := S3Instance.bucket.Put(dest, getFileContent(filename), "content-type", s3.Private, s3.Options{})
	if err != nil {
		log.Fatal(err)
	}
}

func prepareForUpload(srcFile *File) int {
	file, err := os.Open(_segmentsFileName)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	count := 0
	var files []string
	for scanner.Scan() {
		filename := scanner.Text()
		files = append(files, filename)
		go func(count int) {
			uploadFile("chunks/"+srcFile.id+"/"+filename, filename)
			msg, err := json.Marshal(map[string]interface{}{
				"action":    "transcode",
				"id":        srcFile.id,
				"chunk":     strconv.Itoa(count),
				"extension": srcFile.extension,
			})
			failOnError(err, "Failed to marshal message")
			publishToQueue(_workerQueueName, "text/json", msg)
		}(count)
		count += 1
	}
	fmt.Printf("Got %d files, returning from prepareForUpload call\n", count)
	return count
}
