package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-helm-api/pkg/configs"
	"github.com/softplan/tenkai-helm-api/pkg/global"
	"github.com/softplan/tenkai-helm-api/pkg/model"
	"github.com/softplan/tenkai-helm-api/pkg/rabbitmq"
	helmapi "github.com/softplan/tenkai-helm-api/pkg/service/_helm"
	"github.com/softplan/tenkai-helm-api/pkg/service/core"
	"github.com/streadway/amqp"
	"go.elastic.co/apm/module/apmgorilla"
)

//AppContext AppContext
type AppContext struct {
	ConventionInterface core.ConventionInterface
	HelmServiceAPI      helmapi.HelmServiceInterface
	K8sConfigPath       string
	Configuration       *configs.Configuration
	ChartImageCache     sync.Map
	DockerTagsCache     sync.Map
	ConfigMapCache      sync.Map
	RabbitImpl          rabbitmq.RabbitInterface
	Mutex               sync.Mutex
}

func consumeInstallQueue(appContext *AppContext) {
	functionName := "consumeInstallQueue"
	msgs, err := appContext.RabbitImpl.GetConsumer(
		rabbitmq.InstallQueue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: functionName, "error": err.Error()},
			"error when call GetCosumer")
		panic(err)
	}

	go func() {
		for delivery := range msgs {
			global.Logger.Info(
				global.AppFields{global.Function: functionName},
				global.MessageReceived)
			out := &bytes.Buffer{}

			var payload rabbitmq.RabbitPayloadConsumer
			json.Unmarshal([]byte(delivery.Body), &payload)

			createEnvironmentFile(
				payload.Name,
				payload.Token,
				payload.Filename,
				payload.CACertificate,
				payload.ClusterURI,
				payload.Namespace,
			)

			str, err := appContext.doUpgrade(payload.UpgradeRequest, out)
			err = appContext.sendInstallResponse(str, err, payload.DeploymentID)
		}
	}()
}

func consumeRepositoriesQueue(appContext *AppContext) {
	functionName := "consumeRepositoriesQueue"
	msgs, err := appContext.RabbitImpl.GetConsumer(
		rabbitmq.RepositoriesQueue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: functionName, "error": err.Error()},
			"error when call GetCosumer")
		panic(err)
	}

	go func() {
		for delivery := range msgs {
			global.Logger.Info(
				global.AppFields{global.Function: functionName},
				global.MessageReceived)
			var repo model.Repository
			json.Unmarshal([]byte(delivery.Body), &repo)
			err = appContext.HelmServiceAPI.AddRepository(repo)
			if err != nil {
				global.Logger.Error(
					global.AppFields{global.Function: functionName},
					"Error when try to add a new repo - "+err.Error())
			}
		}
	}()
}

func consumeDeleteRepoQueue(appContext *AppContext) {
	functionName := "consumeDeleteRepoQueue"
	msgs, err := appContext.RabbitImpl.GetConsumer(
		rabbitmq.DeleteRepoQueue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: functionName, "error": err.Error()},
			"error when call GetCosumer")
		panic(err)
	}

	go func() {
		for delivery := range msgs {
			global.Logger.Info(
				global.AppFields{global.Function: functionName},
				global.MessageReceived)
			var repo string
			repo = string(delivery.Body)
			err = appContext.HelmServiceAPI.RemoveRepository(repo)
			if err != nil {
				global.Logger.Error(
					global.AppFields{global.Function: functionName},
					"Error when try to del some repo - "+err.Error())
			}
		}
	}()
}

//StartConsumer start consume from queues
func StartConsumer(appContext *AppContext) {
	consumeInstallQueue(appContext)
	consumeRepositoriesQueue(appContext)
	consumeDeleteRepoQueue(appContext)

	forever := make(chan bool)
	global.Logger.Info(
		global.AppFields{global.Function: "StartConsumer"},
		"[*] Waiting for new messages")
	<-forever
}

func (appContext *AppContext) sendInstallResponse(str string, err error, deploymentID uint) error {
	var success bool
	var errorMessage string
	if err != nil {
		success = false
		errorMessage = err.Error()
	} else {
		success = true
		errorMessage = ""
	}

	payload := rabbitmq.RabbitPayloadProducer{
		Success:      success,
		Error:        errorMessage,
		DeploymentID: deploymentID,
	}

	payloadJSON, _ := json.Marshal(payload)

	return appContext.RabbitImpl.Publish(
		"",
		rabbitmq.ResultInstallQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payloadJSON,
		},
	)
}

//StartHTTPServer StartHTTPServer
func StartHTTPServer(appContext *AppContext) {

	port := appContext.Configuration.Server.Port
	global.Logger.Info(global.AppFields{global.Function: "startHTTPServer", "port": port}, "online - listen and server")

	r := mux.NewRouter()

	defineRotes(r, appContext)

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
