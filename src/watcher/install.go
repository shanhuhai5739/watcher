package watcher

import (
	"utils/conf"
	"utils/xlog"
)

func init() {
	var err error

	// init xlog
	xlogConfig := make(map[string]string)
	logs, err := conf.GetSect("logs")
	if err != nil {
		panic(err)
	}
	xlogConfig["path"] = logs["path"]
	xlogConfig["filename"] = logs["filename"]
	xlogConfig["level"] = logs["level"]
	xlogConfig["service"] = logs["name"]
	err = xlog.InitLogger("file", xlogConfig)
	if err != nil {
		panic(err)
	}
}
