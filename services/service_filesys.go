package services

import (
	"bufio"
	"os"
	"panel_backend/global"
	"strings"

	"panel_backend/models"

	"github.com/gin-gonic/gin"
)

func SearchFile(c *gin.Context) {
	//搜索
	path := c.Query("path")
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		global.Log.Errorf("[%s]路径不合法:[%s]\n", path, err.Error())
		c.JSON(400, gin.H{
			"msg": "路径不存在，请重新输入",
		})
		return
	}
	files := make([]*models.File, 0)
	for _, entry := range dirEntries {
		fileinfo, err := entry.Info()
		if err != nil {
			global.Log.Errorf("[%s]读取目录信息失败:[%s]\n", entry.Name(), err.Error())
			c.JSON(400, gin.H{
				"msg": "读取文件信息失败",
			})
			return
		}
		path = strings.TrimRight(path, "/")
		file := models.File{
			Name:    fileinfo.Name(),
			Path:    path + "/" + fileinfo.Name(),
			Size:    fileinfo.Size(),
			ModTime: fileinfo.ModTime().Format("2006-01-02 15:04:05"),
			IsDir:   fileinfo.IsDir(),
		}
		files = append(files, &file)
	}
	// fmt.Printf("files: %v\n", files)
	global.Log.Debugf("返回[%s]信息成功\n", path)
	c.JSON(200, gin.H{
		"files": files,
	})
}
func AddFile(c *gin.Context) {
	var file models.File
	c.BindJSON(&file)
	path := file.Path
	name := file.Name
	isDir := file.IsDir
	// 去掉可能的最后一个斜杠
	path = strings.TrimRight(path, "/")
	if isDir {
		// 创建目录
		err := os.Mkdir(path+"/"+name, os.ModePerm)
		if err != nil {
			global.Log.Errorf("[%s]创建目录失败:[%s]\n", path+"/"+name, err.Error())
			c.JSON(500, gin.H{
				"msg": "系统错误，创建目录失败",
			})
			return
		}
		global.Log.Debugf("创建[%s]目录成功\n", path+"/"+name)
		c.JSON(200, gin.H{
			"msg": "创建成功",
		})
		return
	} else {
		// 创建文件
		_, err := os.Create(path + "/" + name)
		if err != nil {
			global.Log.Errorf("[%s]创建文件失败:[%s]\n", path+"/"+name, err.Error())
			c.JSON(500, gin.H{
				"msg": "系统错误，创建文件失败",
			})
			return
		}
		global.Log.Debugf("创建[%s]文件成功\n", path+"/"+name)
		c.JSON(200, gin.H{
			"msg": "创建成功",
		})
		return
	}

}
func DeleteFile(c *gin.Context) {
	// body, _ := io.ReadAll(c.Request.Body)
	// fmt.Printf("消息体：%s",string(body))
	if c.GetHeader("Override") != "DELETE" {
		global.Log.Errorf("删除文件请求不合法,Override请求头缺失或有误\n")
		c.JSON(400, gin.H{
			"msg": "请求不合法",
		})
		return
	}
	var deleteRequest models.DeleteRequest
	c.BindJSON(&deleteRequest)
	path := strings.TrimRight(deleteRequest.Path, "/")
	// 循环删除
	for _, name := range deleteRequest.Names {
		currentPath := path + "/" + name
		currentName := strings.Split(name, ".")
		if len(currentName) == 1 {
			err := os.RemoveAll(currentPath)
			if err != nil {
				global.Log.Errorf("[%s]删除失败:[%s]\n", currentPath, err.Error())
				c.JSON(500, gin.H{
					"msg": "系统错误，删除失败",
				})
				return
			}
			global.Log.Debugf("删除[%s]成功\n", currentPath)
		} else {
			err := os.Remove(currentPath)
			if err != nil {
				global.Log.Errorf("[%s]删除失败:[%s]\n", currentPath, err.Error())
				c.JSON(500, gin.H{
					"msg": "系统错误，删除失败",
				})
				return
			}
			global.Log.Debugf("删除[%s]成功\n", currentPath)
		}
	}
	c.JSON(200, gin.H{
		"msg": "删除成功",
	})
}
func RenameFile(c *gin.Context) {
	type RenameRequest struct {
		Path string `json:"path"`
	}
	var renameRequest RenameRequest
	c.BindJSON(&renameRequest)
	path := renameRequest.Path
	name := c.Param("oldername")
	newName := c.Param("newname")
	path = strings.TrimRight(path, "/")
	err := os.Rename(path+"/"+name, path+"/"+newName)
	if err != nil {
		global.Log.Errorf("[%s]重命名失败:[%s]\n", path+"/"+name, err.Error())
		c.JSON(500, gin.H{
			"msg": "系统错误，重命名失败",
		})
		return
	}
	global.Log.Debugf("重命名[%s]成功\n", path+"/"+name)
	c.JSON(200, gin.H{
		"msg": "重命名成功",
	})
}
func ReadFile(c *gin.Context) {
	// 读取文件内容
	path := c.Query("path")
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		global.Log.Errorf("[%s]打开文件失败:[%s]\n", path, err.Error())
		c.JSON(500, gin.H{
			"msg": "系统错误，打开文件失败",
		})
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	content := ""
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			break
		}
		content += string(buf[:n])
	}
	global.Log.Debugf("读取[%s]文件成功\n", path)
	c.JSON(200, gin.H{
		"text": content,
	})
}
type WriteFileText struct {
	Path string `json:"path"`
	Text string `json:"text"`
	Name string `json:"name"`
}
func WriteFile(c *gin.Context) {
	// 写入文件内容
	var writeFileText WriteFileText
	c.BindJSON(&writeFileText)
	path := writeFileText.Path
	path = strings.TrimRight(path, "/") + "/" + writeFileText.Name

	text := writeFileText.Text
	file, err := os.OpenFile(path, os.O_WRONLY | os.O_TRUNC, os.ModePerm)
	if err != nil {
		global.Log.Errorf("[%s]打开文件失败:[%s]\n", path, err.Error())
		c.JSON(500, gin.H{
			"msg": "系统错误，打开文件失败",
		})
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	_, err = writer.Write([]byte(text))
	if err != nil {
		global.Log.Errorf("[%s]写入文件失败:[%s]\n", path, err.Error())
		c.JSON(500, gin.H{
			"msg": "系统错误，写入文件失败",
		})
		return
	}
	// 将缓冲区中的数据刷新到磁盘
	err = writer.Flush()
	if err != nil {
		global.Log.Errorf("[%s]刷新缓冲区失败:[%s]\n", path, err.Error())
		c.JSON(500, gin.H{
			"msg": "系统错误，刷新缓冲区失败",
		})
		return
	}

}
