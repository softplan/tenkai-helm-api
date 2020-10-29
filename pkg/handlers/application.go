package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/softplan/tenkai-helm-api/pkg/configs"
	"github.com/softplan/tenkai-helm-api/pkg/rabbitmq"
	helmapi "github.com/softplan/tenkai-helm-api/pkg/service/_helm"
	"github.com/softplan/tenkai-helm-api/pkg/service/core"
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
	handeInstallQueue(appContext)
}

func handeInstallQueue(appContext *AppContext) {
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
		fmt.Println(err)
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
			fmt.Println(str,err)
		}
	}()

	fmt.Println(" [*] - waiting for messages")
	<-forever
}