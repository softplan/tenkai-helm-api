package handlers

import (
	"testing"

	mockSvc "github.com/softplan/tenkai-helm-api/pkg/service/_helm/mocks"
	"github.com/softplan/tenkai-helm-api/pkg/service/rabbitmq"
	"github.com/softplan/tenkai-helm-api/pkg/service/rabbitmq/mocks"
	"github.com/stretchr/testify/mock"
)

func getInstallPayload() rabbitmq.Install {
	install := rabbitmq.Install{}
	install.UpgradeRequest = getUpgradeRequest("xpto")
	install.Filename = "xpto"
	return install
}

func TestHandleInstallQueueOk(t *testing.T) {
	sr := getListSearchResult()

	mockHelmSvc := mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("SearchCharts", mock.Anything, mock.Anything).Return(&sr)
	mockHelmSvc.On("Upgrade", mock.Anything, mock.Anything).Return(nil)

	mockRabbit := mocks.RabbitInterface{}
	mockRabbit.On("Publish", mock.Anything, mock.Anything, false, false, mock.Anything).Return(nil)

	appContext := AppContext{}
	appContext.HelmServiceAPI = &mockHelmSvc
	appContext.RabbitMQ = &mockRabbit
	appContext.handleInstallQueue(getInstallPayload())
}
