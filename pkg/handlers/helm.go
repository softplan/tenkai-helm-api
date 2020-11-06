package handlers

import (
	"bytes"
	"errors"

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
