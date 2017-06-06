package xlog_test

import (
	"errors"
	"mi.com/base/xlog"
	"testing"
)

func TestFileLogger(t *testing.T) {

	config := make(map[string]string)
	config["path"] = "./logs/"
	config["filename"] = "test"
	config["level"] = "debug"
	config["service"] = "xm_shopapi_service"

	xlog.InitLogger("file", config)
	xlog.Debug("this is my xlog helper print to %s %s to file", "aaaaa", "bbbb")
	xlog.Debugx("1234", "this is my xlog helper print to file")

	xlog.Fatal("this is my xlog helper print to file")
	xlog.Fatalx("1234", "this is my xlog helper print to file")

	t.Log("logger open succ")

	//.SetLevel("Notice")
	ip := "10.237.38.299"
	err := errors.New("invalid ip error")
	xlog.Warnx("1234", "get ip[%s] from server, err:%v", ip, err)
}

func TestConsoleLogger(t *testing.T) {

	config := make(map[string]string)
	config["level"] = "debug"
	config["service"] = "console_service"

	xlog.InitLogger("console", config)
	xlog.Debug("this is my xlog helper print to file")
	xlog.Debugx("1234", "this is my xlog helper print to file")

	xlog.Fatal("this is my xlog helper print to file")
	xlog.Fatalx("1234", "this is my xlog helper print to file")

	t.Log("logger open succ")

	//.SetLevel("Notice")
	ip := "10.237.38.299"
	err := errors.New("invalid ip error")
	xlog.Warnx("1234", "get ip[%s] from server, err:%v", ip, err)
}

func TestFileAndConsoleLogger(t *testing.T) {

	config := make(map[string]string)
	config["level"] = "debug"
	config["service"] = "xm_shopapi_service"

	xlog.InitLogger("console", config)

	file := make(map[string]string)
	file["path"] = "./logs/"
	file["filename"] = "test"
	file["level"] = "debug"
	file["service"] = "xm_shopapi_service"

	xlog.InitLogger("file", file)

	xlog.Debugx("1234", "this is my xlog helper print to file")
	xlog.Debug("this is my xlog helper print to file")

	xlog.Fatalx("1234", "this is my xlog helper print to file")
	xlog.Fatal("this is my xlog helper print to file")

	t.Log("logger open succ")

	//.SetLevel("Notice")
	ip := "10.237.38.299"
	err := errors.New("invalid ip error")
	xlog.Warnx("1234", "get ip[%s] from server, err:%v", ip, err)
	xlog.Warn("get ip[%s] from server, err:%v", ip, err)

}
