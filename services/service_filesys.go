package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"panel_backend/global"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
			"code": 400,
			"msg":  "路径不存在，请重新输入",
		})
		return
	}
	files := make([]*models.File, 0)
	for _, entry := range dirEntries {
		fileinfo, err := entry.Info()
		if err != nil {
			global.Log.Errorf("[%s]读取目录信息失败:[%s]\n", entry.Name(), err.Error())
			c.JSON(400, gin.H{
				"code": 400,
				"msg":  "读取文件信息失败",
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
		"code":  200,
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
				"code": 500,
				"msg":  "系统错误，创建目录失败",
			})
			return
		}
		global.Log.Debugf("创建[%s]目录成功\n", path+"/"+name)
		c.JSON(200, gin.H{
			"code": 200,
			"msg":  "创建成功",
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
			"code": 200,
			"msg":  "创建成功",
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
			"code": 400,
			"msg":  "请求不合法",
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
					"code": 500,
					"msg":  "系统错误，删除失败",
				})
				return
			}
			global.Log.Debugf("删除[%s]成功\n", currentPath)
		} else {
			err := os.Remove(currentPath)
			if err != nil {
				global.Log.Errorf("[%s]删除失败:[%s]\n", currentPath, err.Error())
				c.JSON(500, gin.H{
					"code": 500,
					"msg":  "系统错误，删除失败",
				})
				return
			}
			global.Log.Debugf("删除[%s]成功\n", currentPath)
		}
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "删除成功",
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
			"code": 500,
			"msg":  "系统错误，重命名失败",
		})
		return
	}
	global.Log.Debugf("重命名[%s]成功\n", path+"/"+name)
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "重命名成功",
	})
}
func ReadFile(c *gin.Context) {
	// 读取文件内容
	path := c.Query("path")
	global.Log.Debugf("读取[%s]文件\n", path)
	done := make(chan bool, 1)
	cancel := make(chan bool, 1)
	go func() {
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
	lable1:
		for {
			select {
			case <-cancel:
				global.Log.Errorf("读取[%s]文件取消\n", path)
				return
			default:
				n, err := reader.Read(buf)
				if err != nil {
					break lable1
				}
				content += string(buf[:n])
			}
		}
		global.Log.Debugf("读取[%s]文件成功\n", path)
		c.JSON(200, gin.H{
			"text": content,
		})
		done <- true
	}()
	select {
	case <-done:
		return
	case <-time.After(3 * time.Second):
		global.Log.Errorf("读取[%s]文件超时\n", path)
		// 通知协程取消
		cancel <- true
		c.JSON(500, gin.H{
			"msg": "系统错误或文件过大，读取文件超时",
		})
		return
	}
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
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
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
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "写入成功",
	})

}

type PasteRequest struct {
	OldPath     string   `json:"oldPath"`
	NewPath     string   `json:"newPath"`
	Names       []string `json:"names"`
	DeleteNames []string `json:"deleteNames"`
}

func CutPasteFile(c *gin.Context) {
	// 粘贴文件
	var pasteRequest PasteRequest
	c.BindJSON(&pasteRequest)
	oldPath := pasteRequest.OldPath
	newPath := pasteRequest.NewPath
	names := pasteRequest.Names
	// deleteNames := pasteRequest.DeleteNames
	global.Log.Infof("oldPath: %s, newPath: %s\n, names: %v\n", oldPath, newPath, names)
	oldPath = strings.TrimRight(oldPath, "/")
	newPath = strings.TrimRight(newPath, "/")
	// 删除文件
	// for _, name := range deleteNames {
	// 	currentPath := filepath.Join(newPath, name)
	// 	fileInfo, err := os.Stat(currentPath)
	// 	if err != nil {
	// 		global.Log.Errorf("[%s]获取文件信息失败:[%s]\n", currentPath, err.Error())
	// 		c.JSON(500, gin.H{
	// 			"code": 500,
	// 			"msg":  "系统错误，获取欲覆盖文件信息失败",
	// 		})
	// 		return
	// 	}
	// 	if fileInfo.IsDir() {
	// 		err := os.RemoveAll(currentPath)
	// 		if err != nil {
	// 			global.Log.Errorf("[%s]删除失败:[%s]\n", currentPath, err.Error())
	// 			c.JSON(500, gin.H{
	// 				"code": 500,
	// 				"msg":  "系统错误，删除失败",
	// 			})
	// 			return
	// 		}
	// 	} else {
	// 		err := os.Remove(currentPath)
	// 		if err != nil {
	// 			global.Log.Errorf("[%s]删除失败:[%s]\n", currentPath, err.Error())
	// 			c.JSON(500, gin.H{
	// 				"code": 500,
	// 				"msg":  "系统错误，删除失败",
	// 			})
	// 			return
	// 		}
	// 	}
	// }

	//移动文件
	for _, name := range names {
		err := os.Rename(oldPath+"/"+name, newPath+"/"+name)
		if err != nil {
			global.Log.Errorf("[%s]移动失败:[%s]\n", oldPath+"/"+name, err.Error())
			c.JSON(500, gin.H{
				"code": 500,
				"msg":  "系统错误，移动失败",
			})
			return
		}
		global.Log.Debugf("移动[%s]成功\n", oldPath+"/"+name)
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "移动成功",
	})
}

// FolderSize 计算文件夹大小
//
//	func FolderSize(path string) int64 {
//		var size int64 = 0
//		entries, err := os.ReadDir(path)
//		if err != nil {
//			global.Log.Errorf("[%s]路径不合法:[%s]\n", path, err.Error())
//			return -1
//		}
//		for _, entry := range entries {
//			fileInfo, err := entry.Info()
//			if err != nil {
//				global.Log.Errorf("[%s]获取文件信息失败:[%s]\n", entry.Name(), err.Error())
//				return -1
//			}
//			if fileInfo.IsDir() {
//				size += FolderSize(path + "/" + entry.Name())
//			} else {
//				size += fileInfo.Size()
//			}
//		}
//		return size
//	}
func FolderSize(root string) (uint64, error) {
	var size uint64

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// 如果遇到权限问题或其他错误，记录错误但继续遍历
			global.Log.Errorf("访问路径出错 %s: %v\n", path, err)
			return filepath.SkipDir
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				global.Log.Errorf("无法获取文件信息 %s: %v\n", path, err)
				return nil
			}
			atomic.AddUint64(&size, uint64(info.Size()))
		}

		return nil
	})

	if err != nil {
		return size, fmt.Errorf("遍历目录时发生错误: %w", err)
	}

	return size, nil
}

// AllFilesSize 计算所有文件大小
func AllFilesSize(path string, names []string) uint64 {
	var size uint64
	for _, name := range names {
		fileFullPath := path + "/" + name
		fileInfo, err := os.Stat(fileFullPath)
		if err != nil {
			global.Log.Errorf("[%s]获取文件信息失败:[%s]\n", fileFullPath, err.Error())
			continue
		}
		if fileInfo.IsDir() {
			ssize, _ := FolderSize(fileFullPath)
			size += ssize
		} else {
			size += uint64(fileInfo.Size())
		}
	}
	return size
}

// copyFileWithProgress 拷贝文件并发送进度
func copyFileWithProgress(source, destination string, progressChan chan<- uint64, c *gin.Context) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		global.Log.Errorf("[%s]打开文件失败:[%s]\n", source, err.Error())
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destination)
	if err != nil {
		global.Log.Errorf("[%s]创建文件失败:[%s]\n", destination, err.Error())
		return err
	}
	defer destinationFile.Close()
	fmt.Fprintf(c.Writer, "data: Copying %s to %s\n\n", source, destination)
	fmt.Fprintf(c.Writer, "data: SrcFileName: %s\n\n", source)
	fmt.Fprintf(c.Writer, "data: DestFileName: %s\n\n", destination)
	c.Writer.(http.Flusher).Flush()
	global.Log.Infof("当前正在进行Copying %s to %s\n", source, destination)
	copiedBytes := uint64(0)
	buffer := make([]byte, 1024*1024) // 1MB buffer

	for {
		n, err := sourceFile.Read(buffer)
		if err != nil && err != io.EOF {
			global.Log.Errorf("[%s]读取文件失败:[%s]\n", source, err.Error())
			return err
		}
		if n == 0 {
			break
		}

		_, err = destinationFile.Write(buffer[:n])
		if err != nil {
			global.Log.Errorf("[%s]写入文件失败:[%s]\n", destination, err.Error())
			return err
		}

		copiedBytes = uint64(n)
		progressChan <- copiedBytes
	}

	return nil
}

// copyDirWithProgress 递归拷贝目录并发送进度
func copyDirWithProgress(source, destination string, progressChan chan<- uint64, c *gin.Context) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		global.Log.Errorf("[%s]读取目录失败:[%s]\n", source, err.Error())
		return err
	}

	if err := os.MkdirAll(destination, os.ModePerm); err != nil {
		global.Log.Errorf("[%s]创建目录失败:[%s]\n", destination, err.Error())
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destinationPath := filepath.Join(destination, entry.Name())

		var err error
		if entry.IsDir() {
			err = copyDirWithProgress(sourcePath, destinationPath, progressChan, c)
		} else {
			err = copyFileWithProgress(sourcePath, destinationPath, progressChan, c)
		}

		if err != nil {
			global.Log.Errorf("[%s]拷贝失败:[%s]\n", sourcePath, err.Error())
			return err
		}
	}
	return nil
}

func CopyPasteFile(c *gin.Context) {
	//设置SSE http长连接响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	// 粘贴文件
	oldPath := c.Query("oldPath")
	newPath := c.Query("newPath")
	names := c.Query("names")
	var fileNames []string
	if err := json.Unmarshal([]byte(names), &fileNames); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "存在不合法的文件名",
		})
		return
	}

	global.Log.Infof("oldPath: %s, newPath: %s\n, names: %v\n", oldPath, newPath, fileNames)
	oldPath = strings.TrimRight(oldPath, "/")
	newPath = strings.TrimRight(newPath, "/")
	// 计算所有文件大小
	totalBytes := AllFilesSize(oldPath, fileNames)
	global.Log.Infof("totalBytes: %d\n", totalBytes)
	// 通知前端文件大小
	fmt.Fprintf(c.Writer, "data: TotalBytes: %d\n\n", totalBytes)
	c.Writer.(http.Flusher).Flush()

	progressChan := make(chan uint64)
	doneChan := make(chan bool)
	// 开启协程监测进度
	go func() {
		var copiedBytes uint64 = 0
		var preProgressPercentage int = -1
		for progress := range progressChan {
			copiedBytes += progress
			progressPercentage := int(float64(copiedBytes) / float64(totalBytes) * 100)
			if progressPercentage != preProgressPercentage {
				preProgressPercentage = progressPercentage
				global.Log.Infof("copied: %d当前progressPercentage: %d\n", copiedBytes, progressPercentage)
				fmt.Fprintf(c.Writer, "data: Percent: %d\n\n", progressPercentage)
				// 刷新缓冲区，确保数据立即发送
				if flusher, ok := c.Writer.(http.Flusher); ok {
					if flusher != nil {
						flusher.Flush()
					} else {
						global.Log.Errorf("flusher is nil")
					}
				} else {
					global.Log.Errorf("c.Writer does not implement http.Flusher")
				}
			} else {
				continue
			}
			if copiedBytes == totalBytes {
				doneChan <- true
				break
			}
		}
	}()
	// 复制文件
	for _, name := range fileNames {
		destPath := filepath.Join(newPath, name)
		srcPath := filepath.Join(oldPath, name)
		var err error
		fileInfo, err := os.Stat(srcPath)
		if err != nil {
			global.Log.Errorf("[%s]获取文件信息失败:[%s]\n", srcPath, err.Error())
			return
		}
		if fileInfo.IsDir() {
			err = copyDirWithProgress(srcPath, destPath, progressChan, c)
		} else {
			err = copyFileWithProgress(srcPath, destPath, progressChan, c)
		}

		if err != nil {
			global.Log.Errorf("[%s]拷贝失败:[%s]\n", srcPath, err.Error())
			return
		}
	}
	close(progressChan)
	// 等待拷贝完成
	if <-doneChan {
		fmt.Fprintf(c.Writer, "data: Copy operation completed!\n\n")
		c.Writer.(http.Flusher).Flush()
	}
}

//	type UploadProgress struct {
//		copied map[string]int
//		mu    sync.Mutex
//	}
//
//	func init() {
//		uprogress := &UploadProgress{
//			copied: make(map[string]int),
//		}
//	}
var uploadCopied sync.Map
var uploadTotal sync.Map
var uploadFileName sync.Map

func UploadFileWithProgress(resFile *multipart.FileHeader, destPath string, index string) error {
	// 上传文件
	src, err := resFile.Open()
	if err != nil {
		global.Log.Errorf("[%s]打开文件失败:[%s]\n", resFile.Filename, err.Error())
		return err
	}
	if ch, ok := uploadFileName.Load(index); ok {
        // 将interface{}断言为chan string类型
        chanStr := ch.(chan string)
        chanStr <- resFile.Filename
    }
	defer src.Close()
	dst, err := os.Create(destPath)
	if err != nil {
		global.Log.Errorf("[%s]创建文件失败:[%s]\n", destPath, err.Error())
		return err
	}
	defer dst.Close()
	buf := make([]byte, 1024*1024) // 1MB buffer
	for {
		n, err := src.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			global.Log.Errorf("[%s]读取文件失败:[%s]\n", resFile.Filename, err.Error())
			return err
		}
		_, err = dst.Write(buf[:n])
		if err != nil {
			global.Log.Errorf("[%s]写入文件失败:[%s]\n", destPath, err.Error())
			return err
		}
		// 记录上传进度
		if copied, ok := uploadCopied.Load(index); ok {
			uploadCopied.Store(index, uint64(n)+copied.(uint64))
		} else {
			uploadCopied.Store(index, n)
		}
	}
	return nil
}

func UploadFile(c *gin.Context) {
	form, _ := c.MultipartForm()
	files := form.File["files"]
	path := c.PostForm("path")
	index := c.PostForm("index")
	//声明一个管道，用于存储当前上传文件的名称
	var fileNamesChan = make(chan string, len(files))
	uploadFileName.Store(index, fileNamesChan)
	// 去掉可能的最后一个斜杠
	// path = strings.TrimRight(path, "/")
	totalSize := uint64(0)
	for _, file := range files {
		totalSize += uint64(file.Size)
	}
	uploadTotal.Store(index, totalSize)
	for _, file := range files {
		// 使用filepath.Clean()去掉路径中的多余斜杠
		dst := filepath.Join(filepath.Clean(path), file.Filename)
		// if err := c.SaveUploadedFile(file, dst); err != nil {
		// 	global.Log.Errorf("[%s]上传失败:[%s]\n", dst, err.Error())
		// 	c.JSON(500, gin.H{
		// 		"msg": "系统错误，上传失败",
		// 	})
		// 	return
		// }
		// global.Log.Debugf("上传[%s]成功\n", dst)
		err := UploadFileWithProgress(file, dst, index)
		if err != nil {
			c.JSON(500, gin.H{
				"code": 500,
				"msg":  fmt.Sprintf("系统错误，%s上传失败", file.Filename),
			})
		}
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "上传成功",
	})
}

func UploadFileProgress(c *gin.Context) {
	//设置SSE http长连接响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	index := c.Query("TimeIndex")
	// 返回总大小
	var tSize uint64
	for {
		if totalSize, ok := uploadTotal.Load(index); ok {
			global.Log.Infof("index:%s-totalSize: %d\n", index, totalSize.(uint64))
			fmt.Fprintf(c.Writer, "data: TotalSize: %d\n\n", totalSize.(uint64))
			// 刷新缓冲区，确保数据立即发送
			c.Writer.(http.Flusher).Flush()
			uploadTotal.Delete(index)
			tSize = totalSize.(uint64)
			break
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}
	global.Log.Infof("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n")
	// 开启一个协程返回当前上传文件名
	go func() {
		if ch, ok := uploadFileName.Load(index); ok {
			// 将interface{}断言为chan string类型
			chanStr := ch.(chan string)
			for name := range chanStr {
				fmt.Fprintf(c.Writer, "data: FileName: %s\n\n", name)
				// 刷新缓冲区，确保数据立即发送
				c.Writer.(http.Flusher).Flush()
			}
		}
	}()
	// 返回上传进度
	var preCopied uint64 = 0
	var preProgressPercentage int = 0
	for {
		if copied, ok := uploadCopied.Load(index); ok {
			if copied.(uint64) == preCopied {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			preCopied = copied.(uint64)
			progressPercentage := int(float64(copied.(uint64)) / float64(tSize) * 100)
			if progressPercentage != preProgressPercentage {
				preProgressPercentage = progressPercentage
				global.Log.Infof("percent: %d\n", progressPercentage)
				fmt.Fprintf(c.Writer, "data: Percent: %d\n\n", progressPercentage)
				// 刷新缓冲区，确保数据立即发送
				c.Writer.(http.Flusher).Flush()
			}
			// 上传完成
			if copied.(uint64) == tSize {
				fmt.Fprintf(c.Writer, "data: Upload operation completed!\n\n")
				c.Writer.(http.Flusher).Flush()
				// 关闭对应的fileName管道，删除map中对应的数据
				if ch, ok := uploadFileName.Load(index); ok {
					chanStr := ch.(chan string)
					close(chanStr)
					uploadFileName.Delete(index)
				}
				break
			}
		}
	}
}
