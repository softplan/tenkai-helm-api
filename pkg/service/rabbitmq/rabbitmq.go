package rabbitmq

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/softplan/tenkai-helm-api/pkg/dbms/model"
	"github.com/streadway/amqp"
)

//RabbitInterface interface
type RabbitInterface interface {
	GetConnection(uri string) *amqp.Connection
	GetChannel() *amqp.Channel
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	GetConsumer(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	CreateQueue(queueName string, exclusive bool) error
	CreateFanoutExchange(name string) error
	Bind(queueName, routingKey, exchange string) error
	ConsumeRepoQueue(fn HandlerRepo, repo model.Repository) error
	ConsumeInstallQueue(fn HandlerInstall, install Install) error
	ConsumeDeleteRepoQueue(fn HandlerDeleteRepoQueue) error
	ConsumeUpdateRepoQueue(fn HandlerUpdateRepoQueue) error
}

//RabbitImpl struct
type RabbitImpl struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queues  Queues
}

//Queues struct
type Queues struct {
	InstallQueue       string
	ResultInstallQueue string
	AddRepoQueue       string
	DeleteRepoQueue    string
	UpdateRepoQueue    string
}

//HandlerRepo func that handles with msg RepositoriesQueue
type HandlerRepo func(model.Repository) error

//HandlerInstall func
type HandlerInstall func(Install) error

//HandlerDeleteRepoQueue func
type HandlerDeleteRepoQueue func(string) error

//HandlerUpdateRepoQueue func
type HandlerUpdateRepoQueue func() error

//Exchanges
const (
	ExchangeAddRepo    = "add.repository.fx"
	ExchangeDelRepo    = "del.repository.fx"
	ExchangeUpdateRepo = "update.repository.fx"
)

//GetConnection to the RabbitMQ Server
func (rabbit RabbitImpl) GetConnection(uri string) *amqp.Connection {
	conn, err := amqp.Dial(uri)
	if err != nil {
		panic("Fail to connect RabbitMQ Server " + err.Error())
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
func (rabbit RabbitImpl) CreateQueue(queueName string, exclusive bool) error {
	_, err := rabbit.Channel.QueueDeclare(queueName, false, false, exclusive, false, nil)
	return err
}

//CreateFanoutExchange func
func (rabbit RabbitImpl) CreateFanoutExchange(name string) error {
	err := rabbit.Channel.ExchangeDeclare(
		name, "fanout", false, true, false, false, nil,
	)
	return err
}

//ConsumeRepoQueue func
func (rabbit RabbitImpl) ConsumeRepoQueue(fn HandlerRepo, repo model.Repository) error {
	msgs, err := rabbit.Channel.Consume(rabbit.Queues.AddRepoQueue, "", true, true, false, false, nil)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	for msg := range msgs {
		if err := json.Unmarshal(msg.Body, &repo); err == nil {
			fn(repo)
		} else {
			fmt.Println("Error", err.Error())
		}
	}
	return nil
}

//ConsumeInstallQueue func
func (rabbit RabbitImpl) ConsumeInstallQueue(fn HandlerInstall, install Install) error {
	msgs, err := rabbit.Channel.Consume(rabbit.Queues.InstallQueue, "", true, false, false, false, nil)
	if err != nil {
		return err
	}
	for msg := range msgs {
		if err := json.Unmarshal(msg.Body, &install); err == nil {
			fn(install)
		} else {
			fmt.Println("Error", err.Error())
		}
	}
	return nil
}

//ConsumeDeleteRepoQueue func
func (rabbit RabbitImpl) ConsumeDeleteRepoQueue(fn HandlerDeleteRepoQueue) error {
	msgs, err := rabbit.Channel.Consume(rabbit.Queues.DeleteRepoQueue, "", true, true, false, false, nil)
	if err != nil {
		return err
	}
	for msg := range msgs {
		repo := string(msg.Body)
		repo = strings.Replace(repo, "\"", "", -1)
		fn(repo)
	}
	return nil
}

//ConsumeUpdateRepoQueue func
func (rabbit RabbitImpl) ConsumeUpdateRepoQueue(fn HandlerUpdateRepoQueue) error {
	msgs, err := rabbit.Channel.Consume(rabbit.Queues.UpdateRepoQueue, "", true, true, false, false, nil)
	if err != nil {
		return err
	}
	for msg := range msgs {
		fmt.Println(msg.AppId)
		fn()
	}
	return nil
}

//Bind func
func (rabbit RabbitImpl) Bind(queueName, routingKey, exchange string) error {
	return rabbit.Channel.QueueBind(queueName, routingKey, exchange, false, nil)
}
