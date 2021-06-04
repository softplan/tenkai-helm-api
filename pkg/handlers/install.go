package handlers

import (
	"bytes"
	"encoding/json"

	"github.com/softplan/tenkai-helm-api/pkg/global"
	"github.com/softplan/tenkai-helm-api/pkg/service/rabbitmq"

	"github.com/streadway/amqp"
)

func (appContext *AppContext) handleInstallQueue(payload rabbitmq.Install) error {
	global.Logger.Info(
		global.AppFields{global.Function: "handleInstallQueue"},
		global.MessageReceived)

	createEnvironmentFile(
		payload.Name,
		payload.Token,
		payload.Filename,
		payload.CACertificate,
		payload.ClusterURI,
		payload.Namespace,
	)

	out := &bytes.Buffer{}
	str, err := appContext.doUpgrade(payload.UpgradeRequest, out)
	err = appContext.sendInstallResponse(str, err, payload.DeploymentID)

	return nil
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
