package main

import (
	"log"
	"os"

	"github.com/softplan/tenkai-helm-api/pkg/configs"
	"github.com/softplan/tenkai-helm-api/pkg/dbms"
	"github.com/softplan/tenkai-helm-api/pkg/dbms/repository"
	"github.com/softplan/tenkai-helm-api/pkg/global"
	"github.com/softplan/tenkai-helm-api/pkg/handlers"
	"github.com/softplan/tenkai-helm-api/pkg/rabbitmq"
	helmapi "github.com/softplan/tenkai-helm-api/pkg/service/_helm"
)

const (
	configFileName = "app-helm"
)

func main() {
	logFields := global.AppFields{global.Function: "main"}
	_ = os.Mkdir(global.KubeConfigBasePath, 0777)

	global.Logger.Info(logFields, "loading config properties")

	config, err := configs.ReadConfig(configFileName)
	checkFatalError(err)

	appContext := &handlers.AppContext{Configuration: config}

	dbmsURI := config.App.Dbms.URI
	appContext.Database.Connect(dbmsURI, dbmsURI == "")
	appContext.Repositories = initRepository(&appContext.Database)
	defer appContext.Database.Db.Close()

	//RabbitMQ Connection
	rabbitMQ := initRabbit(config.App.Rabbit.URI)
	appContext.RabbitImpl = rabbitMQ
	defer rabbitMQ.Conn.Close()
	defer rabbitMQ.Channel.Close()

	createQueues(rabbitMQ)

	appContext.K8sConfigPath = global.KubeConfigBasePath
	appContext.HelmServiceAPI = helmapi.HelmServiceBuilder()
	initializeHelm(appContext)

	go handlers.StartConsumer(appContext)
	handlers.StartHTTPServer(appContext)
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
	createQueue(rabbitmq.InstallQueue, rabbitMQ)
	createQueue(rabbitmq.ResultInstallQueue, rabbitMQ)
	createQueue(rabbitmq.RepositoriesQueue, rabbitMQ)
}

func createQueue(queueName string, rabbitMQ rabbitmq.RabbitImpl) {
	_, err := rabbitMQ.Channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: queueName},
			"Could not declare "+queueName+" - "+err.Error())
	}
}

func initRepository(database *dbms.Database) handlers.Repositories {
	repositories := handlers.Repositories{}
	repositories.RepoDAO = &repository.RepoDAOImpl{Db: database.Db}
	return repositories
}

func checkFatalError(err error) {
	if err != nil {
		global.Logger.Error(global.AppFields{global.Function: "upload", "error": err}, "erro fatal")
		log.Fatal(err)
	}
}
