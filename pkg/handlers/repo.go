package handlers

import (
	"encoding/hex"

	"github.com/softplan/tenkai-helm-api/pkg/dbms/model"
	"github.com/softplan/tenkai-helm-api/pkg/global"
	"github.com/softplan/tenkai-helm-api/pkg/util"
)

func logMessageReceived(function string) {
	global.Logger.Info(
		global.AppFields{global.Function: function},
		"Message Received ")
}

func (appContext *AppContext) handleRepoQueue(repo model.Repository) error {
	logMessageReceived("handleRepoQueue")
	if err := appContext.HelmServiceAPI.AddRepository(repo); err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: "handleRepoQueue"},
			"Error when try to add a new repo - "+err.Error())
		return err
	}
	appContext.addRepositoryToDB(repo)
	return nil
}

func (appContext *AppContext) initRepos() error {
	repos, err := appContext.Repositories.RepoDAO.All()
	passKey := appContext.Configuration.App.PassKey
	for _, repo := range repos {
		repo.Password, err = decryptRepoPassword(repo.Password, passKey)
		if err != nil {
			global.Logger.Error(
				global.AppFields{global.Function: "initRepos"},
				"Error trying to decrypt a password - "+err.Error())
			continue
		}
		err = appContext.HelmServiceAPI.AddRepository(repo)
		if err != nil {
			global.Logger.Error(
				global.AppFields{global.Function: "initRepos"},
				"Error trying to add a new repo - "+err.Error())
		}
	}
	return err
}

func encryptRepoPassword(password, passKey string) string {
	secret := util.Encrypt([]byte(password), passKey)
	return hex.EncodeToString(secret)
}

func decryptRepoPassword(cryptedPassword, passKey string) (string, error) {

	data, _ := hex.DecodeString(cryptedPassword)
	decryptedPassword, err := util.Decrypt(data, passKey)
	if err != nil {
		return "", err
	}
	return string(decryptedPassword), err
}

func (appContext *AppContext) addRepositoryToDB(repo model.Repository) error {
	passKey := appContext.Configuration.App.PassKey
	repo.Password = encryptRepoPassword(repo.Password, passKey)
	err := appContext.Repositories.RepoDAO.CreateOrUpdate(repo)
	if err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: "addRepository"},
			"Error when try to add a new repo on database - "+err.Error())
	}
	return err
}

func (appContext *AppContext) handleDeleteRepoQueue(repo string) error {
	logMessageReceived("handleDeleteRepoQueue")
	if err := appContext.HelmServiceAPI.RemoveRepository(repo); err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: "consumeDeleteRepoQueue"},
			"Error when try to del some repo - "+err.Error())
		return err
	}
	if err := appContext.Repositories.RepoDAO.Delete(repo); err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: "consumeDeleteRepoQueue"},
			"Error when try to del some repo of database - "+err.Error())
		return err
	}
	return nil
}

func (appContext *AppContext) handleUpdateRepoQueue() error {
	logMessageReceived("handleUpdateRepoQueue")
	appContext.HelmServiceAPI.RepoUpdate()
	return nil

}
