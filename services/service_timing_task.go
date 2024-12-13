package services

import (
	"errors"
	"os"
	"os/exec"
	"panel_backend/global"
	"panel_backend/models"
	"panel_backend/repository"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// 判断字符串是否是文件路径
func isFilePath(input string) bool {
	absPath, err := filepath.Abs(input)
	if err != nil {
		return false
	}
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return false
	}
	return !fileInfo.IsDir()
}

// 判断字符串是否是终端命令
func isCommand(input string) bool {
	_, err := exec.LookPath(input)
	return err == nil
}

// 解析字符串获取命令和参数
func parseCommand(input string) (string, string, []string) {
	// 使用空格分割命令和参数
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", "", nil
	}
	command := parts[0]  // 第一个部分是命令
	args := parts[1:]    // 后续部分是参数
	return input, command, args
}
func RunCommand(input string) error {
	// 解析命令
	_, command, args := parseCommand(input)
	if command == "" {
		return nil
	}
	if isFilePath(command) {
		// 如果是文件路径，需要添加执行权限
		err := os.Chmod(command, 0755)
		if err != nil {
			return err
		}
	} else if !isFilePath(command) && !isCommand(command) {
		return errors.New("filePath not found")
	}
	if !isCommand(command) {
		return errors.New("command not found")
	}
	// 执行命令
	cmd := exec.Command(command, args...)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// err := cmd.Run()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	global.Log.Infof("执行%v命令成功：%s", command,string(output))
	return nil
}

func AddTask(c *gin.Context) {
	//添加定时任务
	var timingTask models.TimingTask
	c.ShouldBindJSON(&timingTask)
	global.TaskMu.Lock()
	defer global.TaskMu.Unlock()
	//添加定时任务到定时任务列表
	id, err := global.TaskC.AddFunc(timingTask.Timing, func() {
		//执行定时任务
		err := RunCommand(timingTask.Command)
		if err != nil {
			global.Log.Errorf("%s定时任务执行失败：%s", timingTask.Command, err)
		}
	})

	if err != nil {
		global.Log.Errorf("定时任务添加失败：%s", err)
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "定时任务添加失败",
		})
		return
	}
	global.TaskMap[timingTask.TaskName] = id
	//添加定时任务到数据库
	err = repository.AddTask(timingTask)
	if err != nil {
		global.Log.Errorf("定时任务添加到数据库失败：%s", err)
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "定时任务添加到数据库失败:任务名重复",
		})
		global.TaskC.Remove(id)
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "定时任务添加成功",
	})

}

type DeleteTaskRequest struct {
	TaskNames []string `json:"task_names"`
}

func DeleteTask(c *gin.Context) {
	var timingTask DeleteTaskRequest
	c.ShouldBindJSON(&timingTask)
	global.Log.Infof("删除定时任务：%v", timingTask.TaskNames)
	global.TaskMu.Lock()
	defer global.TaskMu.Unlock()
	//删除定时任务
	for _, taskName := range timingTask.TaskNames {
		id, ok := global.TaskMap[taskName]
		if !ok {
			global.Log.Errorf("定时任务不存在：%s", taskName)
			c.JSON(500, gin.H{
				"code": 500,
				"msg":  taskName + "定时任务不存在",
			})
			return
		}

		//删除数据库中的定时任务
		err := repository.DeleteTask(taskName)
		if err != nil {
			global.Log.Errorf("定时任务删除失败：%s", err)
			c.JSON(500, gin.H{
				"code": 500,
				"msg":  taskName + "定时任务删除失败",
			})
			return
		}
		global.TaskC.Remove(id)
		delete(global.TaskMap, taskName)
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "定时任务删除成功",
	})
}

type UpdateTaskRequest struct {
	OldTaskName string `json:"old_task_name"`
	NewTaskName string `json:"new_task_name"`
	Timing      string `json:"timing"`
	Command     string `json:"command"`
	Describe    string `json:"describe"`
}

func UpdateTask(c *gin.Context) {
	//更新定时任务
	var updateTaskRequest UpdateTaskRequest
	c.ShouldBindJSON(&updateTaskRequest)
	global.TaskMu.Lock()
	defer global.TaskMu.Unlock()
	//删除旧的定时任务
	id, ok := global.TaskMap[updateTaskRequest.OldTaskName]
	if !ok {
		global.Log.Infof("定时任务不存在：%s", updateTaskRequest.OldTaskName)
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "定时任务不存在",
		})
		return
	}
	//删除数据库中的定时任务
	err := repository.DeleteTask(updateTaskRequest.OldTaskName)
	if err != nil {
		global.Log.Errorf("定时任务删除失败：%s", err)
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "定时任务删除失败",
		})
		return
	}
	global.TaskC.Remove(id)
	delete(global.TaskMap, updateTaskRequest.OldTaskName)
	//添加新的定时任务
	id, err = global.TaskC.AddFunc(updateTaskRequest.Timing, func() {
		//执行定时任务
		err := RunCommand(updateTaskRequest.Command)
		if err != nil {
			global.Log.Errorf("%s定时任务执行失败：%s", updateTaskRequest.Command, err)
		}
	})
	if err != nil {
		global.Log.Errorf("定时任务添加失败：%s", err)
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "定时任务添加失败",
		})
		return
	}
	global.TaskMap[updateTaskRequest.NewTaskName] = id
	//添加定时任务到数据库
	err = repository.AddTask(models.TimingTask{
		TaskName: updateTaskRequest.NewTaskName,
		Timing:   updateTaskRequest.Timing,
		Command:  updateTaskRequest.Command,
		Describe: updateTaskRequest.Describe,
	})
	if err != nil {
		global.Log.Errorf("定时任务添加到数据库失败：%s", err)
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "定时任务添加到数据库失败",
		})
		global.TaskC.Remove(id)
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "定时任务更新成功",
	})
}
func GetTaskList(c *gin.Context) {
	//获取定时任务列表
	tasks, err := repository.GetTaskList()
	if err != nil {
		global.Log.Errorf("获取定时任务列表失败：%s", err)
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "获取定时任务列表失败",
		})
		return
	}
	// global.Log.Infof("获取定时任务列表成功:%v", tasks)
	c.JSON(200, gin.H{
		"code":  200,
		"tasks": tasks,
		"msg":   "获取定时任务列表成功",
	})
}
