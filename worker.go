package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/streadway/amqp"
)

const (
	_apiQueueName    = "flipendo-api"
	_workerQueueName = "flipendo-worker"
)

var MQInstance struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func failOnError(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func createQueues() {
	_, err := MQInstance.channel.QueueDeclare(
		_apiQueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare api queue")

	_, err = MQInstance.channel.QueueDeclare(
		_workerQueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare worker queue")
}

func connectToBroker() {
	var err error
	MQInstance.connection, err = amqp.Dial(os.Getenv("RABBITMQ_PORT_5672_TCP_ADDR"))
	failOnError(err, "Failed to connect to RabbitMQ")
	fmt.Println("Successfully connected to RabbitMQ")
	MQInstance.channel, err = MQInstance.connection.Channel()
	failOnError(err, "Failed to open a channel")
}

func disconnectFromBroker() {
	fmt.Println("Disconnecting from Message Broker...")
	MQInstance.connection.Close()
}

func listenToWQueue() {
	msgs, err := MQInstance.channel.Consume(
		_workerQueueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register worker consumer")

	for d := range msgs {
		log.Printf("Received a message: %s", d.Body)
	}
}

func main() {
	connectToBroker()
	defer disconnectFromBroker()
	createQueues()
	go listenToWQueue()

	testFile := NewFile("test.test")
	fmt.Println("file %s set", testFile.filename)

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-termChan
}
