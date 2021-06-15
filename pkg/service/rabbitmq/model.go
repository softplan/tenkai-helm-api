package rabbitmq

import (
	helmapi "github.com/softplan/tenkai-helm-api/pkg/service/_helm"
)

//Install consumer
type Install struct {
	UpgradeRequest helmapi.UpgradeRequest `json:"upgradeRequest"`
	Name           string                 `json:"name"`
	Token          string                 `json:"token"`
	Filename       string                 `json:"filename"`
	CACertificate  string                 `json:"ca_certificate"`
	ClusterURI     string                 `json:"cluster_uri"`
	Namespace      string                 `json:"namespace"`
	DeploymentID   uint                   `json:"deployment_id"`
}

//RabbitPayloadProducer producer
type RabbitPayloadProducer struct {
	Success      bool   `json:"sucess"`
	Error        string `json:"error"`
	DeploymentID uint   `json:"deployment_id"`
}
