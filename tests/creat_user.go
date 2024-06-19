package tests

import (
	"panel_backend/models"
	"panel_backend/repository"
)

func CreateUserTest() error{
	//创建用户
	user := models.User{
		Username: "admin",
		Password: "admin",
	}
	err := repository.CreateUser(user)
	return err
}