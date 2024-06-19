package tests

import (
	"panel_backend/global"
	"panel_backend/models"
)

func CreateUserTable() {
	//创建表
	global.DB.AutoMigrate(&models.User{})
}
