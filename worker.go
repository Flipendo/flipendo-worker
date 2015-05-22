package main

import (
	"fmt"
	"os"

	"github.com/streadway/amqp"
)

const (
	_apiChannelName    = "apiChannel"
	_workerChannelName = "workerChannel"
)

var MQInstance struct {
	server        *amqp.Connection
	apiChannel    *amqp.Channel
	workerChannel *amqp.Channel
}

func failOnError(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func connectToBroker() {
	var err error
	MQInstance.server, err = amqp.Dial(os.Getenv("RABBITMQ_PORT_5672_TCP_ADDR"))
	failOnError(err, "Failed to connect to RabbitMQ")
	fmt.Println("Successfully connected to RabbitMQ")
}

func disconnectFromBroker() {
	MQInstance.server.Close()
}

func main() {
	connectToBroker()
	defer disconnectFromBroker()
	testFile := NewFile("test.test")
	fmt.Println("file %s set", testFile.filename)
}
