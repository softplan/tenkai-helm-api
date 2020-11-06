package main

import (
	"log"
	"os"

	"github.com/softplan/tenkai-helm-api/pkg/configs"
	"github.com/softplan/tenkai-helm-api/pkg/global"
	"github.com/softplan/tenkai-helm-api/pkg/handlers"
	"github.com/softplan/tenkai-helm-api/pkg/rabbitmq"
	helmapi "github.com/softplan/tenkai-helm-api/pkg/service/_helm"
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

	createQueues(rabbitMQ)

	appContext.K8sConfigPath = global.KubeConfigBasePath
	appContext.HelmServiceAPI = helmapi.HelmServiceBuilder()
	initializeHelm(appContext)

	handlers.StartConsumer(appContext)
}

func initializeHelm(appContext *handlers.AppContext) {
	if _, err := os.Stat(global.HelmDir + "/repository/repositories.yaml"); os.IsNotExist(err) {
		appContext.HelmServiceAPI.InitializeHelm()
	}
}

func initRabbit(uri string) rabbitmq.RabbitImpl {
	rabbitMQ := rabbitmq.RabbitImpl{}
	rabbitMQ.Conn = rabbitMQ.GetConnection(uri)
	rabbitMQ.Channel = rabbitMQ.GetChannel()

	return rabbitMQ
}

func createQueues(rabbitMQ rabbitmq.RabbitImpl) {
	createInstallQueue(rabbitMQ)
	createResultInstallQueue(rabbitMQ)
}

func createInstallQueue(rabbitMQ rabbitmq.RabbitImpl) {
	_, err := rabbitMQ.Channel.QueueDeclare("InstallQueue", true, false, false, false, nil)
	if err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: "createInstallQueue"},
			"Could not declare InstallQueue - "+err.Error())
	}
}

func createResultInstallQueue(rabbitMQ rabbitmq.RabbitImpl) {
	_, err := rabbitMQ.Channel.QueueDeclare("ResultInstallQueue", true, false, false, false, nil)
	if err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: "createResultInstallQueue"},
			"Could not declare ResultInstallQueue - "+err.Error())
	}
}

func checkFatalError(err error) {
	if err != nil {
		global.Logger.Error(global.AppFields{global.Function: "upload", "error": err}, "erro fatal")
		log.Fatal(err)
	}
}
