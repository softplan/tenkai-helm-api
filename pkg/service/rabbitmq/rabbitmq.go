package rabbitmq

import (
	"encoding/json"
	"fmt"

	"github.com/softplan/tenkai-helm-api/pkg/dbms/model"
	"github.com/streadway/amqp"
)

//RabbitInterface interface
type RabbitInterface interface {
	GetConnection(uri string) *amqp.Connection
	GetChannel() *amqp.Channel
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	GetConsumer(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	CreateQueue(queueName string) error
	CreateFanoutExchange(name string) error
	ConsumeRepoQueue(queueName string, fn HandlerRepo, repo model.Repository) error
	Bind(queueName, routingKey, exchange string) error
}

//RabbitImpl struct
type RabbitImpl struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

//Queues struct
type Queues struct {
	InstallQueue       string
	ResultInstallQueue string
	AddRepoQueue       string
	DeleteRepoQueue    string
}

//HandlerRepo func that handles with msg RepositoriesQueue
type HandlerRepo func(model.Repository) error

//Exchanges
const (
	ExchangeAddRepo = "add.repository.fx"
)

//GetConnection to the RabbitMQ Server
func (rabbit RabbitImpl) GetConnection(uri string) *amqp.Connection {
	conn, err := amqp.Dial(uri)
	if err != nil {
		//log
		panic("Fail to connect RabbitMQ Server")
	}
	return conn
}

//GetChannel with rabbitMQ Server
func (rabbit RabbitImpl) GetChannel() *amqp.Channel {
	ch, err := rabbit.Conn.Channel()
	if err != nil {
		//log
		panic("Fail to open a channel with RabbitMQ Server")
	}
	return ch
}

//Publish a message on queue
func (rabbit RabbitImpl) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return rabbit.Channel.Publish(exchange, key, mandatory, immediate, msg)
}

//GetConsumer queue
func (rabbit RabbitImpl) GetConsumer(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return rabbit.Channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}

//CreateQueue func
func (rabbit RabbitImpl) CreateQueue(queueName string) error {
	_, err := rabbit.Channel.QueueDeclare(queueName, false, true, false, false, nil)
	return err
}

//CreateFanoutExchange func
func (rabbit RabbitImpl) CreateFanoutExchange(name string) error {
	err := rabbit.Channel.ExchangeDeclare(
		name, "fanout", false, false, false, false, nil,
	)
	return err
}

//ConsumeRepoQueue func
func (rabbit RabbitImpl) ConsumeRepoQueue(queueName string, fn HandlerRepo, repo model.Repository) error {
	msgs, err := rabbit.Channel.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		return err
	}
	for msg := range msgs {
		fmt.Println("Message Received")
		if err := json.Unmarshal(msg.Body, &repo); err == nil {
			fn(repo)
		} else {
			fmt.Println("Error", err.Error())
		}
	}
	return nil
}

//Bind func
func (rabbit RabbitImpl) Bind(queueName, routingKey, exchange string) error {
	return rabbit.Channel.QueueBind(queueName, routingKey, exchange, false, nil)
}
