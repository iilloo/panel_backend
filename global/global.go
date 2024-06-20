package global

import (
	"os/exec"
	_ "os/exec"
	"panel_backend/config"
	"panel_backend/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 全局只存在一个Config对象
var (
	Config *config.Config
	Log *logrus.Logger
	DB *gorm.DB
	Secret string
	UserCount int
	Bash *models.Bash
	CMD *exec.Cmd
)
	
