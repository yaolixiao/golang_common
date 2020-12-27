package lib

import (
	"os"
	"github.com/spf13/viper"
	dlog "github.com/yaolixiao/golang_common/log"
	"io/ioutil"
	"bytes"
	"strings"
)

type BaseConf struct {
	DebugMode    string
	TimeLocation string
	Log 		 LogConfig
	Base 		 struct {
		DebugMode 	 string
		TimeLocation string
	}
}

type LogConfFileWriter struct {
	On 				bool
	LogPath 		string
	RotateLogPath 	string
	WfLogPath 		string
	RotateWfLogPath string
}

type LogConfConsoleWriter struct {
	On 	  bool
	Color bool
}

type LogConfig struct {
	Level string
	FW LogConfFileWriter
	CW LogConfConsoleWriter
}

type RedisMapConf struct {
	List map[string]*RedisConf
}

type RedisConf struct {
	ProxyList 	 []string
	Password  	 string
	Db 		  	 int
	ConnTimeout  int
	ReadTimeout  int
	WriteTimeout int
}

var ConfBase *BaseConf
var ConfRedis *RedisConf
var ConfRedisMap *RedisMapConf
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