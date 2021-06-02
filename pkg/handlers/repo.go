package handlers

import (
	"encoding/hex"
	"encoding/json"

	"github.com/softplan/tenkai-helm-api/pkg/dbms/model"
	"github.com/softplan/tenkai-helm-api/pkg/global"
	"github.com/softplan/tenkai-helm-api/pkg/util"
)

func (appContext *AppContext) handleRepoQueue(repo model.Repository) error {
	err := appContext.HelmServiceAPI.AddRepository(repo)
	if err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: "handleRepoQueue"},
			"Error when try to add a new repo - "+err.Error())
	}
	return err
}

func (appContext *AppContext) initRepos() error {
	repos, err := appContext.Repositories.RepoDAO.All()
	passKey := appContext.Configuration.App.PassKey
	for _, repo := range repos {
		repo.Password, err = decryptRepoPassword(repo.Password, passKey)
		if err != nil {
			global.Logger.Error(
				global.AppFields{global.Function: "initRepos"},
				"Error when try to add a new repo - "+err.Error())
			continue
		}
		err = appContext.HelmServiceAPI.AddRepository(repo)
		if err != nil {
			global.Logger.Error(
				global.AppFields{global.Function: "initRepos"},
				"Error when try to add a new repo - "+err.Error())
		}
	}
	return err
}

func encryptRepoPassword(password, passKey string) string {
	secret := util.Encrypt([]byte(password), passKey)
	return hex.EncodeToString(secret)
}

func decryptRepoPassword(cryptedPassword, passKey string) (string, error) {
	data, _ := json.Marshal(cryptedPassword)
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

func consumeDeleteRepoQueue(appContext *AppContext) {
	functionName := "consumeDeleteRepoQueue"
	msgs, err := appContext.RabbitMQ.GetConsumer(
		appContext.Queues.DeleteRepoQueue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: functionName, "error": err.Error()},
			"error when call GetCosumer")
		panic(err)
	}

	go func() {
		for delivery := range msgs {
			global.Logger.Info(
				global.AppFields{global.Function: functionName},
				global.MessageReceived)
			var repo string
			repo = string(delivery.Body)
			err = appContext.HelmServiceAPI.RemoveRepository(repo)
			if err != nil {
				global.Logger.Error(
					global.AppFields{global.Function: functionName},
					"Error when try to del some repo - "+err.Error())
			}
		}
	}()
}
