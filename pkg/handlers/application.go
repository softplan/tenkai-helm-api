package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/softplan/tenkai-helm-api/pkg/dbms/model"
	"github.com/softplan/tenkai-helm-api/pkg/service/rabbitmq"
	"github.com/softplan/tenkai-helm-api/pkg/util"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-helm-api/pkg/configs"
	"github.com/softplan/tenkai-helm-api/pkg/dbms"
	"github.com/softplan/tenkai-helm-api/pkg/dbms/repository"
	"github.com/softplan/tenkai-helm-api/pkg/global"
	helmapi "github.com/softplan/tenkai-helm-api/pkg/service/_helm"
	"github.com/softplan/tenkai-helm-api/pkg/service/core"
	"github.com/streadway/amqp"
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
	RabbitImpl          rabbitmq.RabbitInterface
	Mutex               sync.Mutex
	Database            dbms.Database
	Repositories        Repositories
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
			appContext.addRepositoryToDB(repo)
			err = appContext.HelmServiceAPI.AddRepository(repo)
			if err != nil {
				global.Logger.Error(
					global.AppFields{global.Function: functionName},
					"Error when try to add a new repo - "+err.Error())
			}
		}
	}()
}

func (appContext *AppContext) addRepositoryToDB(repo model.Repository) error {
	passKey := appContext.Configuration.App.PassKey
	repo.Password = encryptRepoPassword(repo.Password, passKey)
	err := appContext.Repositories.RepoDAO.CreateOrUpdate(repo)
	if err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: "addRepository"},
			"Error when try to add a new repo on database - "+err.Error())
	}
	return err
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

	appContext.initRepos()

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

func (appContext *AppContext) initRepos() error {
	repos, err := appContext.Repositories.RepoDAO.All()
	passKey := appContext.Configuration.App.PassKey
	for _, repo := range repos {
		repo.Password, err = decryptRepoPassword(repo.Password, passKey)
		if err != nil {
			global.Logger.Error(
				global.AppFields{global.Function: "initRepos"},
				"Error when try to add a new repo - "+err.Error())
			continue
		}
		err = appContext.HelmServiceAPI.AddRepository(repo)
		if err != nil {
			global.Logger.Error(
				global.AppFields{global.Function: "initRepos"},
				"Error when try to add a new repo - "+err.Error())
		}
	}
	return err
}

func encryptRepoPassword(password, passKey string) string {
	secret := util.Encrypt([]byte(password), passKey)
	return hex.EncodeToString(secret)
}

func decryptRepoPassword(cryptedPassword, passKey string) (string, error) {
	data, _ := json.Marshal(cryptedPassword)
	decryptedPassword, err := util.Decrypt(data, passKey)
	if err != nil {
		return "", err
	}
	return string(decryptedPassword), err
}
