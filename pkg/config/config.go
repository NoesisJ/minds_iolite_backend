package config

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func LoadConfig() (*DBConfig, error) {
	// 开发环境硬编码配置
	return &DBConfig{
		Host:     "tarsgo.com",
		Port:     "3306",
		User:     "tarsgo",
		Password: "xf210398444@",
		DBName:   "tarsgo",
	}, nil
}
