package helmapi

import (
	"fmt"
	model2 "github.com/softplan/tenkai-helm-api/pkg/dbms/model"
	"strings"

	"github.com/softplan/tenkai-helm-api/pkg/global"

	"github.com/Masterminds/semver"
	"github.com/gosuri/uitable"
	"github.com/helm/helm/cmd/helm/search"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
)

// searchMaxScore suggests that any score higher than this is not considered a match.
const searchMaxScore = 25

type searchCmd struct {
	out      []*search.Result
	helmhome helmpath.Home

	versions bool
	regexp   bool
	version  string
	colWidth uint
}

//SearchCharts Methotd
func (svc HelmServiceImpl) SearchCharts(searchTerms []string, allVersions bool) *[]model2.SearchResult {

	logFields := global.AppFields{global.Function: "SearchCharts"}

	global.Logger.Info(logFields, "Starting searchCharts")

	sc := &searchCmd{}

	var z helmpath.Home = global.HelmDir
	sc.helmhome = z
	sc.versions = allVersions

	error := sc.run(searchTerms)

	if error != nil {

		global.Logger.Error(logFields, "Error listing charts"+error.Error())
	}

	global.Logger.Info(logFields, "Filling model")
	res := sc.out

	var sr []model2.SearchResult

	for _, r := range res {
		item := &model2.SearchResult{Name: r.Name, ChartVersion: r.Chart.Version, AppVersion: r.Chart.AppVersion, Description: r.Chart.Description}
		sr = append(sr, *item)
	}

	global.Logger.Info(logFields, "Returning model")

	return &sr

}

func (s *searchCmd) run(args []string) error {
	index, err := s.buildIndex()
	if err != nil {
		return err
	}

	var res []*search.Result
	if len(args) == 0 {
		res = index.All()
	} else {
		q := strings.Join(args, " ")
		res, err = index.Search(q, searchMaxScore, s.regexp)
		if err != nil {
			return err
		}
	}

	search.SortScore(res)
	data, err := s.applyConstraint(res)
	if err != nil {
		return err
	}
	s.out = data
	return nil
}

func (s *searchCmd) applyConstraint(res []*search.Result) ([]*search.Result, error) {
	if len(s.version) == 0 {
		return res, nil
	}

	constraint, err := semver.NewConstraint(s.version)
	if err != nil {
		return res, fmt.Errorf("an invalid version/constraint format: %s", err)
	}

	data := res[:0]
	foundNames := map[string]bool{}
	for _, r := range res {
		if _, found := foundNames[r.Name]; found {
			continue
		}
		v, err := semver.NewVersion(r.Chart.Version)
		if err != nil || constraint.Check(v) {
			data = append(data, r)
			if !s.versions {
				foundNames[r.Name] = true // If user hasn't requested all versions, only show the latest that matches
			}
		}
	}

	return data, nil
}

func (s *searchCmd) formatSearchResults(res []*search.Result, colWidth uint) string {
	if len(res) == 0 {
		return "No results found"
	}
	table := uitable.New()
	table.MaxColWidth = colWidth
	table.AddRow("NAME", "CHART VERSION", "APP VERSION", "DESCRIPTION")
	for _, r := range res {
		table.AddRow(r.Name, r.Chart.Version, r.Chart.AppVersion, r.Chart.Description)
	}
	return table.String()
}

func (s *searchCmd) buildIndex() (*search.Index, error) {
	// Load the repositories.yaml
	rf, err := repo.LoadRepositoriesFile(s.helmhome.RepositoryFile())
	if err != nil {
		return nil, err
	}

	i := search.NewIndex()
	for _, re := range rf.Repositories {
		n := re.Name
		f := s.helmhome.CacheIndex(n)
		ind, err := repo.LoadIndexFile(f)
		if err != nil {
			//fmt.Fprintf(s.out, "WARNING: Repo %q is corrupt or missing. Try 'helm repo update'.\n", n)
			continue
		}

		i.AddRepo(n, ind, s.versions || len(s.version) > 0)
	}
	return i, nil
}
