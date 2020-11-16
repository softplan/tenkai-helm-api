package handlers

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/softplan/tenkai-helm-api/pkg/configs"
	"github.com/softplan/tenkai-helm-api/pkg/global"
	"github.com/softplan/tenkai-helm-api/pkg/model"
	"github.com/softplan/tenkai-helm-api/pkg/rabbitmq"
	helmapi "github.com/softplan/tenkai-helm-api/pkg/service/_helm"
	"github.com/softplan/tenkai-helm-api/pkg/service/core"
	"github.com/streadway/amqp"
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
			global.AppFields{global.Function: "consumeInstallQueue", "error": err.Error()},
			"error when call GetCosumer")
		panic(err)
	}

	go func() {
		for delivery := range msgs {
			global.Logger.Info(
				global.AppFields{global.Function: "consumeInstallQueue"},
				"Message Received")
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
			global.AppFields{global.Function: "consumeRepositoriesQueue", "error": err.Error()},
			"error when call GetCosumer")
		panic(err)
	}

	go func() {
		for delivery := range msgs {
			global.Logger.Info(
				global.AppFields{global.Function: "consumeRepositoriesQueue"},
				"Message Received")
			var repo model.Repository
			json.Unmarshal([]byte(delivery.Body), &repo)
			err = appContext.HelmServiceAPI.AddRepository(repo)
			if err != nil {
				global.Logger.Error(
					global.AppFields{global.Function: "consumeRepositoriesQueue"},
					"Error when try to add a new repo - "+err.Error())
			}
		}
	}()
}

func consumeDeleteRepoQueue(appContext *AppContext) {
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
			global.AppFields{global.Function: "consumeDeleteRepoQueue", "error": err.Error()},
			"error when call GetCosumer")
		panic(err)
	}

	go func() {
		for delivery := range msgs {
			global.Logger.Info(
				global.AppFields{global.Function: "consumeDeleteRepoQueue"},
				"Message Received")
			var repo string
			repo = string(delivery.Body)
			err = appContext.HelmServiceAPI.RemoveRepository(repo)
			if err != nil {
				global.Logger.Error(
					global.AppFields{global.Function: "consumeDeleteRepoQueue"},
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
