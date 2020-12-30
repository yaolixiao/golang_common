package lib

import (
	"os"
	"github.com/spf13/viper"
	dlog "github.com/yaolixiao/golang_common/log"
	"github.com/yaolixiao/gorm"
	"database/sql"
	"io/ioutil"
	"bytes"
	"strings"
)

type BaseConf struct {
	DebugMode    string `mapstructure:"debug_mode"`
	TimeLocation string `mapstructure:"time_location"`
	Log 		 LogConfig `mapstructure:"log"`
	Base 		 struct {
		DebugMode 	 string `mapstructure:"debug_mode"`
		TimeLocation string `mapstructure:"time_location"`
	} `mapstructure:"base"`
}

type LogConfFileWriter struct {
	On 				bool `mapstructure:"on"`
	LogPath 		string `mapstructure:"log_path"`
	RotateLogPath 	string `mapstructure:"rotate_log_path"`
	WfLogPath 		string `mapstructure:"wf_log_path"`
	RotateWfLogPath string `mapstructure:"rotate_wf_log_path"`
}

type LogConfConsoleWriter struct {
	On 	  bool `mapstructure:"on"`
	Color bool `mapstructure:"color"`
}

type LogConfig struct {
	Level string `mapstructure:"log_level"`
	FW LogConfFileWriter `mapstructure:"file_writer"`
	CW LogConfConsoleWriter `mapstructure:"console_writer"`
}

type RedisMapConf struct {
	List map[string]*RedisConf `mapstructure:"list"`
}

type RedisConf struct {
	ProxyList 	 []string `mapstructure:"proxy_list"`
	Password  	 string `mapstructure:"password"`
	Db 		  	 int `mapstructure:"db"`
	ConnTimeout  int `mapstructure:"conn_timeout"`
	ReadTimeout  int `mapstructure:"read_timeout"`
	WriteTimeout int `mapstructure:"write_timeout"`
}

type MysqlMapConf struct {
	List map[string]*MySQLConf `mapstructure:"list"`
}

type MySQLConf struct {
	DriverName string `mapstructure:"list"`
	DataSourceName string `mapstructure:"data_source_name"`
	MaxOpenConn int `mapstructure:"max_open_conn"`
	MaxIdleConn int `mapstructure:"max_idle_conn"`
	MaxConnLifeTime int `mapstructure:"max_conn_life_time"`
}

var ConfBase *BaseConf
var ConfRedis *RedisConf
var ConfRedisMap *RedisMapConf
var DBMapPool map[string]*sql.DB
var GORMMapPool map[string]*gorm.DB
var DBDefaultPool *sql.DB
var GORMDefaultPool *gorm.DB
var ViperConfMap map[string]*viper.Viper

// 初始化配置文件
// 设置支持 .toml配置文件
// 将配置文件内容读取到全局变量 ViperConfMap
func InitViperConf() error {
	f, err := os.Open(ConfEnvPath + "/")
	if err != nil {
		return err
	}

	fileList, err := f.Readdir(1024)
	if err !=  nil {
		return err
	}

	for _, f0 := range fileList {
		if !f0.IsDir() {
			bts, err := ioutil.ReadFile(ConfEnvPath + "/" + f0.Name())
			if err != nil {
				return err
			}
			// 使用viper读取配置内容
			v := viper.New()
			v.SetConfigType("toml")
			v.ReadConfig(bytes.NewBuffer(bts))
			if ViperConfMap == nil {
				ViperConfMap = make(map[string]*viper.Viper)
			}
			filenameArr := strings.Split(f0.Name(), ".")
			ViperConfMap[filenameArr[0]] = v
		}
	}
	return nil
}

func InitBaseConf(path string) error {
	ConfBase = &BaseConf{}
	// 将path代表的文件 反序列化为 ConfBase结构体
	err := ParseConfig(path, ConfBase)
	if err != nil {
		return err
	}

	if ConfBase.DebugMode == "" {
		if ConfBase.Base.DebugMode != "" {
			ConfBase.DebugMode = ConfBase.Base.DebugMode
		} else {
			ConfBase.DebugMode = "debug"
		}
	}
	if ConfBase.TimeLocation == "" {
		if ConfBase.Base.TimeLocation != "" {
			ConfBase.TimeLocation = ConfBase.Base.TimeLocation
		} else {
			ConfBase.TimeLocation = "Asia/Chongqing"
		}
	}
	if ConfBase.Log.Level == "" {
		ConfBase.Log.Level = "trace"
	}

	// 配置日志
	logConf := dlog.LogConfig{
		Level: ConfBase.Log.Level,
		FW: dlog.ConfFileWriter{
			On:              ConfBase.Log.FW.On,
			LogPath:         ConfBase.Log.FW.LogPath,
			RotateLogPath:   ConfBase.Log.FW.RotateLogPath,
			WfLogPath:       ConfBase.Log.FW.WfLogPath,
			RotateWfLogPath: ConfBase.Log.FW.RotateWfLogPath,
		},
		CW: dlog.ConfConsoleWriter{
			On:    ConfBase.Log.CW.On,
			Color: ConfBase.Log.CW.Color,
		},
	}
	if err := dlog.SetupDefaultLogWithConf(logConf); err != nil {
		panic(err)
	}
	dlog.SetLayout("2006-01-02T15:04:05.000")
	return nil
}

func InitRedisConf(path string) error {
	ConfRedis = &RedisConf{}
	err := ParseConfig(path, ConfRedis)
	if err != nil {
		return err
	}
	ConfRedisMap = &RedisMapConf{}
	redisMap := make(map[string]*RedisConf)
	redisMap["default"] = ConfRedis
	ConfRedisMap.List = redisMap
	return nil
}

// 获取配置信息
func GetStringConf(key string) string {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return ""
	}
	v, ok := ViperConfMap[keys[0]]
	if !ok {
		return ""
	}
	return v.GetString(strings.Join(keys[1:len(keys)], "."))
}

// 获取配置信息
func GetIntConf(key string) int {
	keys := strings.Split(key, ".")
	l := len(keys)
	if l < 2 {
		return 0
	}
	v := ViperConfMap[keys[0]]
	return v.GetInt(strings.Join(keys[1:l], "."))
}