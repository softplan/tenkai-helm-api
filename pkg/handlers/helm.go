package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-helm-api/pkg/global"
	"github.com/softplan/tenkai-helm-api/pkg/model"
	helmapi "github.com/softplan/tenkai-helm-api/pkg/service/_helm"
)

func (appContext *AppContext) doUpgrade(upgradeRequest helmapi.UpgradeRequest, out *bytes.Buffer) (string, error) {
	var err error
	upgradeRequest.Chart, err = appContext.getChartName(upgradeRequest.Chart)
	if err != nil {
		return "", err
	}
	err = appContext.HelmServiceAPI.Upgrade(upgradeRequest, out)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (appContext *AppContext) getChartName(name string) (string, error) {

	searchTerms := []string{name}
	searchResult := appContext.HelmServiceAPI.SearchCharts(searchTerms, false)

	if len(*searchResult) > 0 {
		r := *searchResult
		return r[0].Name, nil
	}
	return "", errors.New("Chart does not exists")
}

func (appContext *AppContext) listCharts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	vars := mux.Vars(r)
	repo := vars["repo"]

	all, ok := r.URL.Query()["all"]
	allVersions := true
	if ok && len(all[0]) > 0 {
		allVersions = all[0] == "true"
	}

	searchTerms := []string{repo}
	searchResult := appContext.HelmServiceAPI.SearchCharts(searchTerms, allVersions)
	result := &model.ChartsResult{Charts: *searchResult}

	data, _ := json.Marshal(result)

	w.WriteHeader(http.StatusOK)
	w.Write(data)

}