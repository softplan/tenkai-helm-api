package model

import "github.com/jinzhu/gorm"

//Repository struct
type Repository struct {
	gorm.Model
	Name     string `json:"name"`
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

//RepositoryResult struct
type RepositoryResult struct {
	Repositories []Repository `json:"repositories"`
}

//DefaultRepoRequest struct
type DefaultRepoRequest struct {
	Reponame string `json:"reponame"`
}
