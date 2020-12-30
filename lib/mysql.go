package lib

import (
	"fmt"
	"time"
	"github.com/yaolixiao/gorm"
	_ "github.com/yaolixiao/gorm/dialects/mysql"
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"unicode"
)

func InitDBPool(path string) error {
	DbConfMap := &MysqlMapConf{}
	err := ParseConfig(path, DbConfMap)
	if err != nil {
		return err
	}
	if len(DbConfMap.List) == 0 {
		fmt.Printf("[WARN] %s %s\n", time.Now().Format(TimeFormat), " empty mysql config.")
	}

	DBMapPool = map[string]*sql.DB{}
	GORMMapPool = map[string]*gorm.DB{}
	for confName, DBConf := range DbConfMap.List {
		// 普通的DB方式
		db, err := sql.Open("mysql", DBConf.DataSourceName)
		if err != nil {
			return err
		}
		db.SetMaxOpenConns(DBConf.MaxOpenConn)
		db.SetMaxIdleConns(DBConf.MaxIdleConn)
		db.SetConnMaxLifetime(time.Duration(DBConf.MaxConnLifeTime) * time.Second)
		// 检查数据库连接是否仍然有效
		err = db.Ping()
		if err != nil {
			return err
		}

		// gorm的DB方式
		dbgorm, err := gorm.Open("mysql", DBConf.DataSourceName)
		if err != nil {
			return err
		}

		// gorm默认的结构体映射是复数形式，比如你的博客表为blog，对应的结构体名就会是blogs
		// 若表名为多个单词，对应的model结构体名字必须是驼峰式，首字母也必须大写
		// 配置 DB.SingularTable(true) 以实现结构体名为非复数形式, 禁用表名复数
		// 如果只是部分表需要使用源表名，请在实体类中声明TableName的构造函数
		// func (实体名) TableName() string {
		// 		return "数据库表名"
		// }
		dbgorm.SingularTable(true)
		err = dbgorm.DB().Ping()
		if err != nil {
			return err
		}

		dbgorm.LogMode(true)
		dbgorm.LogCtx(true)
		dbgorm.SetLogger(&MysqlGormLogger{Trace: NewTrace()})
		dbgorm.DB().SetMaxIdleConns(DBConf.MaxIdleConn)
		dbgorm.DB().SetMaxOpenConns(DBConf.MaxOpenConn)
		dbgorm.DB().SetConnMaxLifetime(time.Duration(DBConf.MaxConnLifeTime) * time.Second)
		DBMapPool[confName] = db
		GORMMapPool[confName] = dbgorm
	}
	
	//手动配置连接
	if dbpool, err := GetDBPool("default"); err == nil {
		DBDefaultPool = dbpool
	}
	if dbpool, err := GetGormPool("default"); err == nil {
		GORMDefaultPool = dbpool
	}
	return nil
}

func GetDBPool(name string) (*sql.DB, error) {
	if dbpool, ok := DBMapPool[name]; ok {
		return dbpool, nil
	}
	return nil, errors.New("get pool error")
}

func GetGormPool(name string) (*gorm.DB, error) {
	if dbpool, ok := GORMMapPool[name]; ok {
		return dbpool, nil
	}
	return nil, errors.New("get pool error")
}

func CloseDB() error {
	for _, dbpool := range DBMapPool {
		dbpool.Close()
	}
	for _, dbpool := range GORMMapPool {
		dbpool.Close()
	}
	return nil
}

func DBPoolLogQuery(trace *TraceContext, sqlDb *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	startExecTime := time.Now()
	rows, err := sqlDb.Query(query, args...)
	endExecTime := time.Now()
	if err != nil {
		Log.TagError(trace, "_com_mysql_success", map[string]interface{}{
			"sql":       query,
			"bind":      args,
			"proc_time": fmt.Sprintf("%f", endExecTime.Sub(startExecTime).Seconds()),
		})
	} else {
		Log.TagInfo(trace, "_com_mysql_success", map[string]interface{}{
			"sql":       query,
			"bind":      args,
			"proc_time": fmt.Sprintf("%f", endExecTime.Sub(startExecTime).Seconds()),
		})
	}
	return rows, err
}

//mysql日志打印类
// Logger default logger
type MysqlGormLogger struct {
	gorm.Logger
	Trace *TraceContext
}

// Print format & print log
func (logger *MysqlGormLogger) Print(values ...interface{}) {
	message := logger.LogFormatter(values...)
	if message["level"] == "sql" {
		Log.TagInfo(logger.Trace, "_com_mysql_success", message)
	} else {
		Log.TagInfo(logger.Trace, "_com_mysql_failure", message)
	}
}

// LogCtx(true) 时会执行改方法
func (logger *MysqlGormLogger) CtxPrint(s *gorm.DB,values ...interface{}) {
	ctx,ok:=s.GetCtx()
	trace:=NewTrace()
	if ok{
		trace=ctx.(*TraceContext)
	}
	message := logger.LogFormatter(values...)
	if message["level"] == "sql" {
		Log.TagInfo(trace, "_com_mysql_success", message)
	} else {
		Log.TagInfo(trace, "_com_mysql_failure", message)
	}
}

func (logger *MysqlGormLogger) LogFormatter(values ...interface{}) (messages map[string]interface{}) {
	if len(values) > 1 {
		var (
			sql             string
			formattedValues []string
			level           = values[0]
			currentTime     = logger.NowFunc().Format("2006-01-02 15:04:05")
			source          = fmt.Sprintf("%v", values[1])
		)

		messages = map[string]interface{}{"level": level, "source": source, "current_time": currentTime}

		if level == "sql" {
			// duration
			//messages = append(messages, fmt.Sprintf("%.2fms", float64(values[2].(time.Duration).Nanoseconds() / 1e4) / 100.0))
			messages["proc_time"] = fmt.Sprintf("%fs", values[2].(time.Duration).Seconds())
			// sql

			for _, value := range values[4].([]interface{}) {
				indirectValue := reflect.Indirect(reflect.ValueOf(value))
				if indirectValue.IsValid() {
					value = indirectValue.Interface()
					if t, ok := value.(time.Time); ok {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
					} else if b, ok := value.([]byte); ok {
						if str := string(b); logger.isPrintable(str) {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
						} else {
							formattedValues = append(formattedValues, "'<binary>'")
						}
					} else if r, ok := value.(driver.Valuer); ok {
						if value, err := r.Value(); err == nil && value != nil {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
						} else {
							formattedValues = append(formattedValues, "NULL")
						}
					} else {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
					}
				} else {
					formattedValues = append(formattedValues, "NULL")
				}
			}

			// differentiate between $n placeholders or else treat like ?
			if regexp.MustCompile(`\$\d+`).MatchString(values[3].(string)) {
				sql = values[3].(string)
				for index, value := range formattedValues {
					placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
					sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
				}
			} else {
				formattedValuesLength := len(formattedValues)
				for index, value := range regexp.MustCompile(`\?`).Split(values[3].(string), -1) {
					sql += value
					if index < formattedValuesLength {
						sql += formattedValues[index]
					}
				}
			}

			messages["sql"] = sql
			if len(values) > 5 {
				messages["affected_row"] = strconv.FormatInt(values[5].(int64), 10)
			}
		} else {
			messages["ext"] = values
		}
	}
	return
}

func (logger *MysqlGormLogger) NowFunc() time.Time {
	return time.Now()
}

func (logger *MysqlGormLogger) isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}