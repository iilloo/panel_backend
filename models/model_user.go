package models

type User struct {
	Model
	Username string `json:"username" gorm:"unique;not null;size:20"`
	Password string `json:"password" gorm:"not null;size:20"`
}
