package repository

import (
	"panel_backend/global"
	"panel_backend/models"
)
func AddTask(timingTask models.TimingTask) error {
	//添加定时任务
	err := global.DB.Create(&timingTask).Error
	return err
}
func DeleteTask(taskName string) error {
	//删除定时任务
	err := global.DB.Where("task_name = ?", taskName).Delete(&models.TimingTask{}).Error
	return err
}
func UpdateTask() {
	//更新定时任务
}
func GetTaskList() ([]models.TimingTask, error){
	//获取定时任务列表
	var timingTasks []models.TimingTask
	err := global.DB.Find(&timingTasks).Error
	return timingTasks, err
}