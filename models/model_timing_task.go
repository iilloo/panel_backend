package models
type TimingTask struct {
	Model
	TaskName string `json:"task_name" gorm:"unique;not null;size:50"`
	Timing string `json:"timing" gorm:"not null;size:50"`
	Command string `json:"command" gorm:"not null;size:100"`
	Describe string `json:"describe" gorm:"size:500"`
}