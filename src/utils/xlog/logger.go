/*
小米网golang日志库

小米网golang日志库支持6种日志级别：

 1）Debug
 2）Trace
 3）Notice
 4）Warn
 5）Fatal
 6）None。

支持两种输出格式：

 1）json格式
 2）自定义输出。

日志级别优先级：

 Debug < Trace < Notice < Warn < Fatal < None。

即如果定义日志级别为Debug：则Trace、Notice、Warn、Fatal等级别的日志都会输出；反之，如果定义日志级别为Trace，则Debug不会输出，其它级别的日志都会输出。当日志级别为None时，将不会有任何日志输出。

*/

package xlog

import (
	"errors"
	"fmt"
	"sync"
)

type XLoggerInstance struct {
	logger  XLogInterface
	enable  bool
	initial bool
	source  string
}

var (
	g_LoggerMgr map[string]*XLoggerInstance = make(map[string]*XLoggerInstance)
	lock        sync.RWMutex
)

func init() {
	config := make(map[string]string)
	config["level"] = "debug"
	initLogger("console", config, "auto")
}

func RegisterLogger(name string, logger XLogInterface) (err error) {

	lock.Lock()
	defer lock.Unlock()

	_, ok := g_LoggerMgr[name]
	if ok {
		err = errors.New(fmt.Sprintf("duplicate logger[%s]", name))
		return
	}

	g_LoggerMgr[name] = &XLoggerInstance{
		logger:  logger,
		enable:  false,
		initial: false,
	}

	return
}
func initLogger(name string, config map[string]string, source ...string) (err error) {

	instance, ok := g_LoggerMgr[name]
	if !ok {
		err = errors.New(fmt.Sprintf("not found logger[%s]", name))
		return
	}

	err = instance.logger.Init(config)
	if err != nil {
		return
	}

	if len(source) > 0 {
		instance.source = source[0]
	} else {
		instance.source = ""
	}

	instance.enable = true
	instance.initial = true

	return
}

func InitLogger(name string, config map[string]string) (err error) {

	lock.Lock()
	defer lock.Unlock()

	err = initLogger(name, config)
	//关闭自动注入的logger
	for _, v := range g_LoggerMgr {

		if v.logger == nil || !v.enable {
			continue
		}

		if v.source == "auto" {
			v.enable = false
		}
	}

	return
}

func EnableLogger(name string, enable bool) (err error) {

	lock.Lock()
	defer lock.Unlock()

	instance, ok := g_LoggerMgr[name]
	if !ok {
		err = errors.New(fmt.Sprintf("not found logger[%s]", name))
		return
	}

	if !instance.initial {
		instance.enable = false
		return
	}

	instance.enable = enable
	return
}

func GetLogger(name string) (logger XLogInterface, err error) {

	lock.RLock()
	defer lock.RUnlock()

	instance, ok := g_LoggerMgr[name]
	if !ok {
		err = errors.New(fmt.Sprintf("not found logger[%s]", name))
		return
	}

	logger = instance.logger
	return
}

func UnregisterLogger(name string) (err error) {

	lock.Lock()
	defer lock.Unlock()

	v, ok := g_LoggerMgr[name]
	if !ok {
		err = errors.New(fmt.Sprintf("not found logger[%s]", name))
		return
	}

	if v != nil {
		v.logger.Close()
	}

	delete(g_LoggerMgr, name)
	return
}

func ReOpen() (err error) {

	var errorMsg string

	lock.RLock()
	defer lock.RUnlock()

	for k, v := range g_LoggerMgr {

		if v.logger == nil || !v.enable {
			continue
		}

		errRet := v.logger.ReOpen()
		if errRet != nil {
			errorMsg += fmt.Sprintf("logger[%s] reload failed, err[%v]\n", k, errRet)
			continue
		}
	}

	err = errors.New(errorMsg)
	return
}

func SetLevelAll(level string) {

	lock.RLock()
	defer lock.RUnlock()

	for _, v := range g_LoggerMgr {

		if !v.enable || v.logger == nil {
			continue
		}

		v.logger.SetLevel(level)
	}
}

func SetLevel(name, level string) (err error) {

	lock.Lock()
	defer lock.Unlock()

	v, ok := g_LoggerMgr[name]
	if !ok || v.logger == nil {
		err = errors.New(fmt.Sprintf("not found logger[%s]", name))
		return
	}

	v.logger.SetLevel(level)
	return
}

func Warn(format string, a ...interface{}) (err error) {

	lock.RLock()
	defer lock.RUnlock()

	for _, v := range g_LoggerMgr {
		if v.logger == nil || !v.enable {
			continue
		}
		v.logger.Warn(format, a...)
	}

	return
}

func Fatal(format string, a ...interface{}) (err error) {

	lock.RLock()
	defer lock.RUnlock()

	for _, v := range g_LoggerMgr {
		if v.logger == nil || !v.enable {
			continue
		}
		v.logger.Fatal(format, a...)
	}

	return
}

func Notice(format string, a ...interface{}) (err error) {

	lock.RLock()
	defer lock.RUnlock()

	for _, v := range g_LoggerMgr {
		if v.logger == nil || !v.enable {
			continue
		}
		v.logger.Notice(format, a...)
	}

	return
}

func Trace(format string, a ...interface{}) (err error) {

	lock.RLock()
	defer lock.RUnlock()

	for _, v := range g_LoggerMgr {
		if v.logger == nil || !v.enable {
			continue
		}
		v.logger.Trace(format, a...)
	}

	return
}

func Debug(format string, a ...interface{}) (err error) {

	lock.RLock()
	defer lock.RUnlock()

	for _, v := range g_LoggerMgr {
		if v.logger == nil || !v.enable {
			continue
		}
		v.logger.Debug(format, a...)
	}

	return
}

func Warnx(logId, format string, a ...interface{}) (err error) {

	lock.RLock()
	defer lock.RUnlock()

	for _, v := range g_LoggerMgr {
		if v.logger == nil || !v.enable {
			continue
		}
		v.logger.Warnx(logId, format, a...)
	}

	return
}

func Fatalx(logId, format string, a ...interface{}) (err error) {

	lock.RLock()
	defer lock.RUnlock()

	for _, v := range g_LoggerMgr {
		if v.logger == nil || !v.enable {
			continue
		}
		v.logger.Fatalx(logId, format, a...)
	}

	return
}

func Noticex(logId, format string, a ...interface{}) (err error) {

	lock.RLock()
	defer lock.RUnlock()

	for _, v := range g_LoggerMgr {
		if v.logger == nil || !v.enable {
			continue
		}
		v.logger.Noticex(logId, format, a...)
	}

	return
}

func Tracex(logId, format string, a ...interface{}) (err error) {

	lock.RLock()
	defer lock.RUnlock()

	for _, v := range g_LoggerMgr {
		if v.logger == nil || !v.enable {
			continue
		}
		v.logger.Tracex(logId, format, a...)
	}

	return
}

func Debugx(logId, format string, a ...interface{}) (err error) {

	lock.RLock()
	defer lock.RUnlock()

	for _, v := range g_LoggerMgr {
		if v.logger == nil || !v.enable {
			continue
		}
		v.logger.Debugx(logId, format, a...)
	}

	return
}

func Close() {

	lock.RLock()
	defer lock.RUnlock()

	for _, v := range g_LoggerMgr {
		if v.logger == nil || !v.enable {
			continue
		}
		v.logger.Close()
	}

	return
}
