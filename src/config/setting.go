package config

var (
	// C 全局配置文件，在Init调用前为nil
	C *Config
)

// Config 配置
type Config struct {
	App     app     `yaml:"app"`
	Redis   redis   `yaml:"redis"`
	MongoDB mongodb `yaml:"mongodb"`
	Web     web     `yaml:"web"`
	LogConf logConf `yaml:"logConf"`
	Debug   bool    `yaml:"debug"`
}

type app struct {
	Addr   string `yaml:"addr"`
	Prefix string `yaml:"prefix"`
}

type redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type mongodb struct {
	Addr     string `yaml:"addr"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}

type web struct {
	Addr string `yaml:"addr"`
}

type logConf struct {
	LogPath     string `yaml:"log_path"`
	LogFileName string `yaml:"log_file_name"`
}
