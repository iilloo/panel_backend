package core

import (
	"os"
	"panel_backend/config"
	"panel_backend/global"

	"gopkg.in/yaml.v2"
)

const yamlPath = "settings.yaml"

func InitConfig() *config.Config {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		//打印错误并退出
		global.Log.Error("读取配置文件失败")
		panic(err)
	}
	c := &config.Config{}
	//读取yaml文件，生成一个Config对象，让c指向这个对象
	err = yaml.Unmarshal(data, c)
	if err != nil {
		//解析yaml失败
		global.Log.Error("解析配置文件失败")
		panic(err)
	}

	return c
}
