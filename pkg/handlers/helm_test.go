package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-helm-api/pkg/dbms/model"
	helmapi "github.com/softplan/tenkai-helm-api/pkg/service/_helm"
	mockSvc "github.com/softplan/tenkai-helm-api/pkg/service/_helm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getListSearchResult() []model.SearchResult {
	return []model.SearchResult{
		{
			Name:         "xptoName",
			ChartVersion: "xptoChart",
			AppVersion:   "1.1.1",
			Description:  "my description...",
		},
	}
}

func getMockHelm(searchTerms string) *mockSvc.HelmServiceInterface {
	sr := getListSearchResult()

	mockHelm := mockSvc.HelmServiceInterface{}
	mockHelm.On("SearchCharts", []string{searchTerms}, false).Return(&sr)
	return &mockHelm
}

func TestListChartsOk(t *testing.T) {
	appContext := AppContext{}
	searchTerms := "xpto"
	mockHelm := getMockHelm(searchTerms)
	appContext.HelmServiceAPI = mockHelm

	endpoint := fmt.Sprintf("/charts/%s", searchTerms)
	req, err := http.NewRequest("GET", endpoint, nil)
	assert.NoError(t, err)

	req = mux.SetURLVars(req, map[string]string{"repo": searchTerms})
	query := req.URL.Query()
	query.Add("all", "false")
	req.URL.RawQuery = query.Encode()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listCharts)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not OK")
}

func getUpgradeRequest(chart string) helmapi.UpgradeRequest {
	return helmapi.UpgradeRequest{
		Chart: chart,
	}
}

func TestDoUpgrade(t *testing.T) {
	searchTerm := "xpto"
	appContext := AppContext{}

	mockHelm := getMockHelm(searchTerm)
	mockHelm.On("Upgrade", mock.Anything, mock.Anything).Return(nil)
	appContext.HelmServiceAPI = mockHelm

	str, err := appContext.doUpgrade(getUpgradeRequest(searchTerm), nil)
	assert.Equal(t, "", str, "Response should be \"\"")
	assert.NoError(t, err)
}

func TestRepoUpdate(t *testing.T) {
	mockHelm := mockSvc.HelmServiceInterface{}
	mockHelm.On("RepoUpdate").Return(nil)

	appContext := AppContext{}
	appContext.HelmServiceAPI = &mockHelm

	req, err := http.NewRequest("GET", "/repoUpdate", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.repoUpdate)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not OK")
}
