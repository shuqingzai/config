package config

type Config struct {
	Server `ini:"server"`
	Mysql  `ini:"mysql"`
}

type Server struct {
	Ip   string `ini:"ip"`
	Port uint   `ini:"port"`
}

type Mysql struct {
	Username string  `ini:"username"`
	Password string  `ini:"password"`
	Database string  `ini:"database"`
	Host     string  `ini:"host"`
	Port     uint    `ini:"port"`
	Timeout  float32 `ini:"timeout"`
}
