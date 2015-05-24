package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/streadway/amqp"
)

const (
	_apiQueueName    = "flipendo-api"
	_workerQueueName = "flipendo-worker"
)

var mqInstance struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func failOnError(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func createQueues() {
	_, err := mqInstance.channel.QueueDeclare(
		_apiQueueName,
		false,
		true,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare api queue")

	_, err = mqInstance.channel.QueueDeclare(
		_workerQueueName,
		false,
		true,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare worker queue")
}

func connectToBroker() {
	var err error
	for i := 0; i < 10; i++ {
		log.Printf("Trying to connect to: %s\n", "amqp://"+os.Getenv("RABBITMQ_PORT_5672_TCP_ADDR")+
			":"+os.Getenv("RABBITMQ_PORT_5672_TCP_PORT"))
		mqInstance.connection, err = amqp.Dial("amqp://" + os.Getenv("RABBITMQ_PORT_5672_TCP_ADDR") +
			":" + os.Getenv("RABBITMQ_PORT_5672_TCP_PORT"))
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	failOnError(err, "Failed to connect to RabbitMQ")
	log.Println("Successfully connected to RabbitMQ")
	mqInstance.channel, err = mqInstance.connection.Channel()
	failOnError(err, "Failed to open a channel")
}

func disconnectFromBroker() {
	log.Println("Disconnecting from Message Broker...")
	mqInstance.connection.Close()
}

func listenToWQueue() {
	msgs, err := mqInstance.channel.Consume(
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
		var data map[string]interface{}
		err = json.Unmarshal(d.Body, &data)
		failOnError(err, "Failed to unmarshal message")
		switch data["action"].(string) {
		case "split":
			log.Println("Received split message")
			file := NewFile(data["id"].(string) + data["extension"].(string))
			file.id = data["id"].(string)
			file.extension = data["extension"].(string)
			nb := file.Split()
			msg, err := json.Marshal(map[string]interface{}{
				"action": "split",
				"id":     data["id"].(string),
				"chunks": nb,
			})
			failOnError(err, "Failed to marshal message")
			log.Printf("publishing: %d chunks to %s\n", nb, _apiQueueName)
			publishToQueue(_apiQueueName, "text/json", msg)
			log.Println("Successfully published message to api channel")
		case "transcode":
			log.Println("Received transcode message")
			file := NewFile(data["id"].(string) + data["extension"].(string))
			file.id = data["id"].(string)
			file.extension = data["extension"].(string)
			file.Transcode(data["chunk"].(string))
		case "merge":
			log.Println("Received merge message")
			file := NewFile(data["id"].(string))
			file.id = data["id"].(string)
			file.Concat(int(data["chunks"].(float64)))
		default:
			log.Fatal("Unrecognized action")
		}
	}
}

func publishToQueue(queueName string, contentType string, body []byte) error {
	err := mqInstance.channel.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: contentType,
			Body:        body,
		},
	)
	return err
}

func main() {
	err := awsInit()
	failOnError(err, "aws init failed")
	connectToBroker()
	defer disconnectFromBroker()
	createQueues()
	go listenToWQueue()

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-termChan
}
