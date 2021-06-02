package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/softplan/tenkai-helm-api/pkg/configs"
	"github.com/softplan/tenkai-helm-api/pkg/dbms"
	"github.com/softplan/tenkai-helm-api/pkg/dbms/repository"
	"github.com/softplan/tenkai-helm-api/pkg/global"
	"github.com/softplan/tenkai-helm-api/pkg/handlers"
	helmapi "github.com/softplan/tenkai-helm-api/pkg/service/_helm"
	"github.com/softplan/tenkai-helm-api/pkg/service/rabbitmq"
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

	queues := getQueues()
	//RabbitMQ Connection
	rabbitMQ, err := rabbitmq.InitRabbit(config.App.Rabbit.URI, queues)
	checkFatalError(err)

	appContext.RabbitMQ = rabbitMQ
	defer rabbitMQ.Conn.Close()
	defer rabbitMQ.Channel.Close()

	appContext.K8sConfigPath = global.KubeConfigBasePath
	appContext.HelmServiceAPI = helmapi.HelmServiceBuilder()
	initializeHelm(appContext)

	handlers.StartConsumer(appContext)
	handlers.StartHTTPServer(appContext)
}

func initializeHelm(appContext *handlers.AppContext) {
	if _, err := os.Stat(global.HelmDir + "/repository/repositories.yaml"); os.IsNotExist(err) {
		appContext.HelmServiceAPI.InitializeHelm()
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

func getQueues() rabbitmq.Queues {
	queues := rabbitmq.Queues{
		InstallQueue:       "InstallQueue",
		ResultInstallQueue: "ResultInstallQueue",
		DeleteRepoQueue:    "DeleteRepoQueue",
	}
	queues.AddRepoQueue = "RepositoriesQueue" + getRandomSufix()
	return queues
}

func getRandomSufix() string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	sufix := r.Intn(1000)
	return strconv.Itoa(sufix)
}
