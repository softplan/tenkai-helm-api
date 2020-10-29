package rabbitmq

import (
	helmapi "github.com/softplan/tenkai-helm-api/pkg/service/_helm"
)

//RabbitPayloadConsumer consumer
type RabbitPayloadConsumer struct {
	UpgradeRequest helmapi.UpgradeRequest `json:"upgradeRequest"`
	Name           string `json:"name"`
	Token          string `json:"token"`
	Filename       string `json:"filename"`
	CACertificate  string `json:"ca_certificate"`
	ClusterURI     string `json:"cluster_uri"`
	Namespace      string `json:"namespace"`
}