package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"
)

func awsInit() (*s3.Bucket, error) {
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

func uploadFile(bucket *s3.Bucket, dest string, filename string) {
	err := bucket.Put(dest, getFileContent(filename), "content-type", s3.Private, s3.Options{})
	if err != nil {
		log.Fatal(err)
	}
}

func prepareForUpload(srcFile *File) int {
	bucket, err := awsInit()
	if err != nil {
		log.Fatal(err)
	}
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
			uploadFile(bucket, "chunks/"+srcFile.id+"/"+filename, filename)
			msg, err := json.Marshal(map[string]interface{}{
				"action":    "transcode",
				"id":        srcFile.id,
				"chunk":     count,
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
