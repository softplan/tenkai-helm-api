package handlers

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/softplan/tenkai-helm-api/pkg/configs"
	"github.com/softplan/tenkai-helm-api/pkg/global"
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

//StartConsumer start consume from queues
func StartConsumer(appContext *AppContext) {
	msgs, err := appContext.RabbitImpl.GetConsumer(
		"InstallQueue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: "StartConsumer", "error": err.Error()},
			 "error when call GetCosumer")
		panic(err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			out := &bytes.Buffer{}

			var payload rabbitmq.RabbitPayloadConsumer
			json.Unmarshal([]byte(d.Body), &payload)

			createEnvironmentFile(
				payload.Name,
				payload.Token,
				payload.Filename,
				payload.CACertificate,
				payload.ClusterURI,
				payload.Namespace,
			)

			str, err := appContext.doUpgrade(payload.UpgradeRequest, out)
			err = appContext.sendResponse(str, err, payload.DeploymentID)
		}
	}()

	global.Logger.Info(
		global.AppFields{global.Function: "StartConsumer"},
		 "[*] Waiting for new messages")
	<-forever
}

func (appContext *AppContext) sendResponse(str string, err error, deploymentID uint) (error) {
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
		Success: success,
		Error: errorMessage,
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
			Body: payloadJSON,
		},
	)
}