package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func prepareForUpload() int {
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
		count += 1
	}
	fmt.Printf("Got %d files, returning from prepareForUpload call\n", count)
	return count
}
