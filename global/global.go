package global

import (
	_ "os/exec"
	"panel_backend/config"
	"panel_backend/models"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 全局只存在一个Config对象
var (
	Config    *config.Config
	Log       *logrus.Logger
	DB        *gorm.DB
	Secret    string
	UserCount int
	Bash      *models.Bash
	// CMD *exec.Cmd
	TaskC       *cron.Cron
	TaskMap = make(map[string]cron.EntryID)
	TaskMu      sync.Mutex // 确保线程安全

)
