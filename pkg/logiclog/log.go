package logiclog

import (
	"github.com/kataras/iris/v12"
	"os"
	// 日志相关依赖库
	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	formatter := Formatter{
		ChildFormatter: &log.JSONFormatter{},
		Line:           true,
		Package:        false,
		File:           true,
		BaseNameOnly:   false,
	}
	log.SetFormatter(&formatter)
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	log.SetLevel(log.InfoLevel)
}

var ServerName string
var Environment string
var InstanceKey string

// InitConfig set global config for logic
func InitConfig(serverName, env, instanceKey, level string) {
	ServerName = serverName
	Environment = env
	InstanceKey = instanceKey

	formatLevel, err := log.ParseLevel(level)
	if err != nil {
		log.SetLevel(log.InfoLevel)
	}
	log.SetLevel(formatLevel)
}

// CtxLogger get logic.Entry with common fields: user， req_id，*logic.Entry returned
func CtxLogger(ctx iris.Context) *log.Entry {
	user := "-"
	if len(ctx.GetHeader("X-User-Name")) > 0 {
		user = ctx.GetHeader("X-User-Name")
	}

	reqId := ctx.GetID()

	// A common pattern is to re-use fields between logging statements by re-using
	// the logic.Entry returned from WithFields()
	contextLogger := log.WithFields(log.Fields{
		"user":         user,
		"req_id":       reqId,
		"server_name":  ServerName,
		"environment":  Environment,
		"instance_key": InstanceKey,
		"logic_type":   "logic",
	})
	return contextLogger
}

// Logger get logic.Entry with common fields: server_name，environment，instance_key，*logic.Entry returned
func Logger() *log.Entry {
	// A common pattern is to re-use fields between logging statements by re-using
	// the logic.Entry returned from WithFields()
	contextLogger := log.WithFields(log.Fields{
		"server_name":  ServerName,
		"environment":  Environment,
		"instance_key": InstanceKey,
		"logic_type":   "logic",
	})
	return contextLogger
}
