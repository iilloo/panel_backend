package repository

import (
	"panel_backend/global"
	"panel_backend/models"
)

func CreateUser(user models.User) error{
	//创建用户
	err := global.DB.Create(&user).Error
	return err
}

func GetUserByUsername(username string) (models.User, error){
	//通过用户名获取用户
	var user models.User
	err := global.DB.Where("username = ?", username).First(&user).Error
	return user, err
}

func GetUserById(id uint) (models.User, error){
	//通过id获取用户
	var user models.User
	err := global.DB.Where("id = ?", id).First(&user).Error
	return user, err
}

func UpdateUser(user models.User) error{
	//更新用户
	err := global.DB.Save(&user).Error
	return err
}



func GetUserList() ([]models.User, error){
	//获取用户列表
	var users []models.User
	err := global.DB.Find(&users).Error
	return users, err
}

func DeleteAllUser() error{
	//删除所有用户
	err := global.DB.Delete(&models.User{}).Error
	return err 
}

func DeleteUserByName(username string) error{
	//通过用户名删除用户
	err := global.DB.Where("username = ?", username).Delete(&models.User{}).Error
	return err
}
