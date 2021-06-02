package handlers

import (
	"bytes"
	"encoding/json"

	"github.com/softplan/tenkai-helm-api/pkg/global"
	"github.com/softplan/tenkai-helm-api/pkg/service/rabbitmq"
	"github.com/streadway/amqp"
)

func consumeInstallQueue(appContext *AppContext) {
	functionName := "consumeInstallQueue"
	msgs, err := appContext.RabbitMQ.GetConsumer(
		appContext.Queues.InstallQueue,
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

	return appContext.RabbitMQ.Publish(
		"",
		appContext.Queues.ResultInstallQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payloadJSON,
		},
	)
}
