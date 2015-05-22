package main

import (
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

var config struct {
	serverUrl         string
	apiChannelName    string
	workerChannelName string
}

var instance struct {
	server        *amqp.Connection
	apiChannel    *amqp.Channel
	workerChannel *amqp.Channel
}

func generateConfig() {
	config.serverUrl = os.Getenv("RABBITMQ_PORT_5672_TCP_ADDR")
	config.apiChannelName = "apiChannel"
	config.workerChannelName = "workerChannel"
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func connectToBroker() {
	var err error
	instance.server, err = amqp.Dial(config.serverUrl)
	failOnError(err, "Failed to connect to RabbitMQ")
	fmt.Println("Successfully connected to RabbitMQ")
}

func disconnectFromBroker() {
	instance.server.Close()
}

func main() {
	generateConfig()
	connectToBroker()
	defer disconnectFromBroker()
	testFile := NewFile("test.test")
	fmt.Println("file %s set", testFile.filename)
}
