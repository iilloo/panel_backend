package services

import (
	"fmt"
	"math"
	"os"
	"panel_backend/global"
	"panel_backend/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/net"
)

func getCpuTime() (uint64, uint64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		global.Log.Errorf("读取cpuStat信息失败:[%s]\n", err.Error())
		return 0, 0, err
	}
	lines := strings.Split(string(data), "\n")

	var cpuTotalTime uint64 = 0
	var cpuFreeTime uint64 = 0
	for _, line := range lines {
		//将每行的内容按空格分割存入fields切片
		fields := strings.Fields(line)

		if len(fields) < 5 || fields[0] != "cpu" {
			continue
		}

		var times []uint64
		//把cpu的各项时间存入times切片
		for _, field := range fields[1:] {
			value, err := strconv.ParseUint(field, 10, 64)
			if err != nil {
				global.Log.Errorf("解析cpuStat信息失败:[%s]\n", err.Error())
				return 0, 0, err
			}
			times = append(times, value)
		}
		//计算cpu总时间和空闲时间
		cpuFreeTime = times[3]
		for _, time := range times {
			cpuTotalTime += time
		}
		break
	}
	return cpuFreeTime, cpuTotalTime, nil
}
func keepTwoDecimals(num float64) float64 {
	return math.Floor(num*100+0.5) / 100
}
func GetCpuInfo() interface{} {
	//获取cpu核心数
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		global.Log.Errorf("读取cpuInfo信息失败:[%s]\n", err.Error())
		return nil
	}
	content := string(data)
	lines := strings.Split(content, "\n")
	coreCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "processor") {
			coreCount++
		}
	}
	//获取cpu使用率
	cpuFreeTime_1, cpuTotalTime_1, err := getCpuTime()
	if err != nil {
		global.Log.Errorf("获取cpu使用率失败:[%s]\n", err.Error())
		return nil
	}
	//等待1秒
	time.Sleep(1 * time.Second)
	cpuFreeTime_2, cpuTotalTime_2, err := getCpuTime()
	if err != nil {
		global.Log.Errorf("获取cpu使用率失败:[%s]\n", err.Error())
		return nil
	}
	cpuFreeSub := cpuFreeTime_2 - cpuFreeTime_1
	cpuTotalSub := cpuTotalTime_2 - cpuTotalTime_1
	cpuUsage := (1 - (float64(cpuFreeSub) / float64(cpuTotalSub)))
	cpuCurrent := float64(coreCount) * cpuUsage
	cpuCurrent = keepTwoDecimals(cpuCurrent)
	cpuUsage = math.Round(cpuUsage * 100)
	return &models.HostItemStatu{
		Name:       "cpu",
		Percentage: int8(cpuUsage),
		Current:    cpuCurrent,
		Sum:        float64(coreCount),
		Suffix:     "核",
	}
}

func GetMemInfo() interface{} {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		global.Log.Errorf("读取memInfo信息失败:[%s]\n", err.Error())
		return nil
	}
	lines := strings.Split(string(data), "\n")
	var totalMemory uint64 = 0
	var freeMemory uint64 = 0
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			totalMemory, _ = strconv.ParseUint(fields[1], 10, 64)
		case "MemAvailable:":
			freeMemory, _ = strconv.ParseUint(fields[1], 10, 64)
		}
	}
	memoryUsage := (1 - (float64(freeMemory) / float64(totalMemory)))
	memoryUsage = math.Round(memoryUsage * 100)
	return &models.HostItemStatu{
		Name:       "内存",
		Percentage: int8(memoryUsage),
		Current:    keepTwoDecimals(float64(totalMemory-freeMemory) / 1024 / 1024),
		Sum:        keepTwoDecimals(float64(totalMemory) / 1024 / 1024),
		Suffix:     "GB",
	}
}

func GetSwapInfo() interface{} {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		global.Log.Errorf("读取memInfo信息失败:[%s]\n", err.Error())
		return nil
	}
	lines := strings.Split(string(data), "\n")
	var totalMemory uint64 = 0
	var freeMemory uint64 = 0
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "SwapTotal:":
			totalMemory, _ = strconv.ParseUint(fields[1], 10, 64)
		case "SwapFree:":
			freeMemory, _ = strconv.ParseUint(fields[1], 10, 64)
		}
	}
	memoryUsage := (1 - (float64(freeMemory) / float64(totalMemory)))
	memoryUsage = math.Round(memoryUsage * 100)
	return &models.HostItemStatu{
		Name:       "Swap",
		Percentage: int8(memoryUsage),
		Current:    keepTwoDecimals(float64(totalMemory-freeMemory) / 1024 / 1024),
		Sum:        keepTwoDecimals(float64(totalMemory) / 1024 / 1024),
		Suffix:     "GB",
	}
}

func GetDiskInfo() interface{} {
	usage, err := disk.Usage("/")
	if err != nil {
		global.Log.Errorf("获取磁盘信息失败:[%s]\n", err.Error())
		return nil
	}
	return &models.HostItemStatu{
		Name:       "磁盘",
		Percentage: int8(math.Round(usage.UsedPercent)),
		Current:    keepTwoDecimals(float64(usage.Used) / 1024 / 1024 / 1024),
		Sum:        keepTwoDecimals(float64(usage.Total) / 1024 / 1024 / 1024),
		Suffix:     "GB",
	}
}

// 根据数据的大小使用对应的单位
func bytesToHumanReadable(bytes uint64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	var i int
	value := float64(bytes)
	for value >= 1024 && i < len(units)-1 {
		value /= 1024
		i++
	}
	return fmt.Sprintf("%.2f %s", value, units[i])
}

// 判断是否是物理网卡
func isPhysicalInterface(name string) bool {
	// Add more conditions if necessary
	invalidPrefixes := []string{"lo", "docker", "br-", "veth", "vmnet", "virbr", "vboxnet"}
	for _, prefix := range invalidPrefixes {
		if strings.HasPrefix(name, prefix) {
			return false
		}
	}
	return true
}

func GetNetInfo() interface{} {

	netStat, err := net.IOCounters(true)
	if err != nil {
		global.Log.Errorf("获取网络信息失败:[%s]\n", err.Error())
		return nil
	}
	//等待1秒
	time.Sleep(1 * time.Second)
	netStat_2, err := net.IOCounters(true)
	if err != nil {
		global.Log.Errorf("获取网络信息失败:[%s]\n", err.Error())
		return nil
	}
	//过滤掉虚拟网卡
	var keys []int = make([]int, 0)
	for i := range netStat {
		if isPhysicalInterface(netStat[i].Name) {
			keys = append(keys, i)
		}
	}
	var netSumRecv_1 uint64 = 0
	var netSumSent_1 uint64 = 0
	var netSumRecv_2 uint64 = 0
	var netSumSent_2 uint64 = 0
	//计算所有物理网卡总的接收和发送字节数
	for _, key := range keys {
		netSumRecv_1 += netStat[key].BytesRecv
		netSumSent_1 += netStat[key].BytesSent
		netSumRecv_2 += netStat_2[key].BytesRecv
		netSumSent_2 += netStat_2[key].BytesSent
	}
	return &models.NetStatus{
		DownloadSpeed: bytesToHumanReadable(netSumRecv_2-netSumRecv_1) + "/s",
		UploadSpeed:   bytesToHumanReadable(netSumSent_2-netSumSent_1) + "/s",
		DownloadTotal: bytesToHumanReadable(netSumRecv_2),
		UploadTotal:   bytesToHumanReadable(netSumSent_2),
	}
}

func HostBasicInfos(c *gin.Context) {
	var SysBasicInfo []models.SysBasicInfo
	//发行版本信息
	platform, _, version, err := host.PlatformInformation()
	if err != nil {
		global.Log.Errorf("获取系统架构失败:[%s]\n", err.Error())
		c.JSON(500, gin.H{
			"msg": "获取系统架构失败",
		})
		return
	}
	SysBasicInfo = append(SysBasicInfo, models.SysBasicInfo{
		Name:  "发行版本",
		Value: platform + " " + version,
	})

	//获取内核版本
	kernelVersion, err := host.KernelVersion()
	if err != nil {
		global.Log.Errorf("获取内核版本失败:[%s]\n", err.Error())
		c.JSON(500, gin.H{
			"msg": "获取内核版本失败",
		})
		return
	}
	SysBasicInfo = append(SysBasicInfo, models.SysBasicInfo{
		Name:  "内核版本",
		Value: kernelVersion,
	})
	//系统架构
	arch, err := host.KernelArch()
	if err != nil {
		global.Log.Errorf("获取系统架构失败:[%s]\n", err.Error())
		c.JSON(500, gin.H{
			"msg": "获取系统架构失败",
		})
		return
	}
	SysBasicInfo = append(SysBasicInfo, models.SysBasicInfo{
		Name:  "系统架构",
		Value: arch,
	})

	//最近开机时间
	hostInfo, err := host.Info()
	if err != nil {
		global.Log.Errorf("获取主机信息失败:[%s]\n", err.Error())
		c.JSON(500, gin.H{
			"msg": "获取主机信息失败",
		})
		return
	}
	bootTime := time.Unix(int64(hostInfo.BootTime), 0)
	bootTimeString := bootTime.Format("2006-01-02 15:04:05")
	SysBasicInfo = append(SysBasicInfo, models.SysBasicInfo{
		Name:  "开机时间",
		Value: bootTimeString,
	})
	//系统已运行时间
	uptime := time.Duration(hostInfo.Uptime) * time.Second
	uptimeDays := int(uptime.Hours() / 24)
	uptimeHours := int(uptime.Hours()) % 24
	SysBasicInfo = append(SysBasicInfo, models.SysBasicInfo{
		Name:  "运行时间",
		Value: fmt.Sprintf("%d天%d小时", uptimeDays, uptimeHours),
	})
	c.JSON(200, gin.H{
		"hostBasicInfos": SysBasicInfo,
	})
}
