package handlers

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/softplan/tenkai-helm-api/pkg/dbms/model"
	"github.com/softplan/tenkai-helm-api/pkg/service/rabbitmq"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-helm-api/pkg/configs"
	"github.com/softplan/tenkai-helm-api/pkg/dbms"
	"github.com/softplan/tenkai-helm-api/pkg/dbms/repository"
	"github.com/softplan/tenkai-helm-api/pkg/global"
	helmapi "github.com/softplan/tenkai-helm-api/pkg/service/_helm"
	"github.com/softplan/tenkai-helm-api/pkg/service/core"
	"go.elastic.co/apm/module/apmgorilla"
)

//Repositories struct
type Repositories struct {
	RepoDAO repository.RepoDAOInterface
}

//AppContext AppContext
type AppContext struct {
	ConventionInterface core.ConventionInterface
	HelmServiceAPI      helmapi.HelmServiceInterface
	K8sConfigPath       string
	Configuration       *configs.Configuration
	ChartImageCache     sync.Map
	DockerTagsCache     sync.Map
	ConfigMapCache      sync.Map
	RabbitMQ            rabbitmq.RabbitInterface
	Mutex               sync.Mutex
	Database            dbms.Database
	Repositories        Repositories
	Queues              rabbitmq.Queues
}

//StartConsumer start consume from queues
func StartConsumer(appContext *AppContext) {
	go appContext.RabbitMQ.ConsumeRepoQueue(appContext.Queues.AddRepoQueue, appContext.handleRepoQueue, model.Repository{})
	//go consumeInstallQueue(appContext)
	//go consumeRepositoriesQueue(appContext)
	//go consumeDeleteRepoQueue(appContext)

	forever1 := make(chan bool)
	global.Logger.Info(
		global.AppFields{global.Function: "StartConsumer"},
		"[*] Waiting for new messages")
	<-forever1
}

//StartHTTPServer StartHTTPServer
func StartHTTPServer(appContext *AppContext) {

	port := appContext.Configuration.Server.Port
	global.Logger.Info(global.AppFields{global.Function: "startHTTPServer", "port": port}, "online - listen and server")

	r := mux.NewRouter()

	defineRotes(r, appContext)

	appContext.initRepos()

	portInt, _ := strconv.Atoi(port)
	port = strconv.Itoa(portInt + 1)
	log.Fatal(http.ListenAndServe(":"+port, commonHandler(r)))
}

func defineRotes(r *mux.Router, appContext *AppContext) {
	r.Use(apmgorilla.Middleware())
	r.HandleFunc("/health", appContext.health).Methods("GET")
	r.HandleFunc("/charts/{repo}", appContext.listCharts).Methods("GET")
	r.HandleFunc("/repoUpdate", appContext.repoUpdate).Methods("GET")
}

func commonHandler(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}
