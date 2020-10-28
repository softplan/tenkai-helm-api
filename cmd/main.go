package main

import (
	"log"
	"os"

	"github.com/softplan/tenkai-helm-api/pkg/configs"
	"github.com/softplan/tenkai-helm-api/pkg/global"
	"github.com/softplan/tenkai-helm-api/pkg/handlers"
	"github.com/softplan/tenkai-helm-api/pkg/rabbitmq"
)

const (
	configFileName = "app"
)

func main() {
	logFields := global.AppFields{global.Function: "main"}
	_ = os.Mkdir(global.KubeConfigBasePath, 0777)

	global.Logger.Info(logFields, "loading config properties")

	config, err := configs.ReadConfig(configFileName)
	checkFatalError(err)

	appContext := &handlers.AppContext{Configuration: config}

	//RabbitMQ Connection
	rabbitMQ := initRabbit(config.App.Rabbit.URI)
	appContext.RabbitImpl = rabbitMQ
	defer rabbitMQ.Conn.Close()
	defer rabbitMQ.Channel.Close()

	handlers.StartConsumer(appContext)

}

func initRabbit(uri string) rabbitmq.RabbitImpl {
	rabbitMQ := rabbitmq.RabbitImpl{}
	rabbitMQ.Conn = rabbitMQ.GetConnection(uri)
	rabbitMQ.Channel = rabbitMQ.GetChannel()

	return rabbitMQ
}

func checkFatalError(err error) {
	if err != nil {
		global.Logger.Error(global.AppFields{global.Function: "upload", "error": err}, "erro fatal")
		log.Fatal(err)
	}
}