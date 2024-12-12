package core

import (
	"panel_backend/global"
	"panel_backend/repository"

	"github.com/robfig/cron/v3"
)
func InitTimingTask() {
	global.TaskMap = make(map[string]cron.EntryID)
	// 初始化 cron 调度器
	global.TaskC = cron.New(cron.WithSeconds()) // 支持秒级调度
	global.TaskC.Start()
	//初始化定时任务
	//从数据库中获取定时任务列表，并添加到定时任务中
	timingTasks, err := repository.GetTaskList()
	if err != nil {
		global.Log.Errorf("获取定时任务列表失败：%s", err)
		return
	}
	for _, timingTask := range timingTasks {
		id, err := global.TaskC.AddFunc(timingTask.Timing, func() {
			//执行定时任务
			global.Log.Infof("定时任务执行：%s", timingTask.TaskName)
		})
		if err != nil {
			global.Log.Errorf("定时任务添加失败：%s", err)
			continue
		}
		global.TaskMap[timingTask.TaskName] = id
	}
}