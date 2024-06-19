package config

// 定义一个System结构体，用于存储系统配置
type System struct {
	IP  string `yaml:"ip"`
	Port string `yaml:"port"`
	ENV string `yaml:"env"`
}

func (s *System) GetAddr() string {
	return s.IP + ":" + s.Port
}