package main

import (
	"fmt"
	"os"
	"log"
	"os/exec"
)

func split(srcFileName string) {
	fmt.Println("Splitting file")
	out, err := exec.Command("date").Output()
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
	srcFileName = os.Args[1]
	split(srcFileName)
	return
}
