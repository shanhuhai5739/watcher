package xlog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	XConsoleLogDefaultLogId = "800000001"
)

type XConsoleLog struct {
	level int

	skip     int
	hostname string
	service  string
}

type Brush func(string) string

func NewBrush(color string) Brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

var colors = []Brush{
	NewBrush("1;37"), // white
	NewBrush("1;36"), // debug			cyan
	NewBrush("1;35"), // trace   magenta
	NewBrush("1;31"), // notice      red
	NewBrush("1;33"), // warn    yellow
	NewBrush("1;32"), // fatal			green
	NewBrush("1;34"), //
	NewBrush("1;34"), //
}

func init() {
	RegisterLogger("console", NewXConsoleLog())
}

func NewXConsoleLog() XLogInterface {
	return &XConsoleLog{
		skip: XLogDefSkipNum,
	}
}

func (p *XConsoleLog) Init(config map[string]string) (err error) {

	level, ok := config["level"]
	if !ok {
		err = errors.New(fmt.Sprintf("init XConsoleLog failed, not found level"))
		return
	}

	service, _ := config["service"]
	if len(service) > 0 {
		p.service = service
	}

	skip, _ := config["skip"]
	if len(skip) > 0 {
		skipNum, err := strconv.Atoi(skip)
		if err == nil {
			p.skip = skipNum
		}
	}

	p.level = LevelFromStr(level)
	hostname, _ := os.Hostname()
	p.hostname = hostname

	return
}

//@title 设置日志级别
//@level：日志级别，如下："Debug", "Trace", "Notice", "Warn", "Fatal", "None"
func (p *XConsoleLog) SetLevel(level string) {
	p.level = LevelFromStr(level)
}

func (p *XConsoleLog) SetSkip(skip int) {
	p.skip = skip
}

func (p *XConsoleLog) ReOpen() error {
	return nil
}

func (p *XConsoleLog) Warn(format string, a ...interface{}) error {

	return p.warnx(XConsoleLogDefaultLogId, format, a...)
}

func (p *XConsoleLog) Warnx(logId, format string, a ...interface{}) error {

	return p.warnx(logId, format, a...)
}

//打印warn日志，当日志级别大于Warn时，不会输出任何日志。
func (p *XConsoleLog) warnx(logId, format string, a ...interface{}) error {

	if p.level > WarnLevel {
		return nil
	}

	logText := Format(format, a...)
	fun, filename, lineno := GetRuntimeInfo(p.skip)

	color := colors[WarnLevel]
	logText = color(fmt.Sprintf("[%s:%s:%d] %s", fun, filepath.Base(filename), lineno, logText))

	return p.write(WarnLevel, &logText, logId)
}

func (p *XConsoleLog) Fatal(format string, a ...interface{}) error {

	return p.fatalx(XConsoleLogDefaultLogId, format, a...)
}

func (p *XConsoleLog) Fatalx(logId, format string, a ...interface{}) error {

	return p.fatalx(logId, format, a...)
}

//打印fatal日志，当日志级别大于Fatal时，不会输出任何日志。
func (p *XConsoleLog) fatalx(logId, format string, a ...interface{}) error {

	if p.level > FatalLevel {
		return nil
	}

	logText := Format(format, a...)
	fun, filename, lineno := GetRuntimeInfo(p.skip)

	color := colors[FatalLevel]
	logText = color(fmt.Sprintf("[%s:%s:%d] %s", fun, filepath.Base(filename), lineno, logText))

	return p.write(FatalLevel, &logText, logId)
}

func (p *XConsoleLog) Notice(format string, a ...interface{}) error {

	return p.Noticex(XConsoleLogDefaultLogId, format, a...)
}

//打印notice日志，当日志级别大于Notice时，不会输出任何日志。
func (p *XConsoleLog) Noticex(logId, format string, a ...interface{}) error {

	if p.level > NoticeLevel {
		return nil
	}

	logText := Format(format, a...)
	return p.write(NoticeLevel, &logText, logId)
}

func (p *XConsoleLog) Trace(format string, a ...interface{}) error {

	return p.tracex(XConsoleLogDefaultLogId, format, a...)
}

func (p *XConsoleLog) Tracex(logId, format string, a ...interface{}) error {
	return p.tracex(logId, format, a...)
}

//打印trace日志，当日志级别大于Trace时，不会输出任何日志。
func (p *XConsoleLog) tracex(logId, format string, a ...interface{}) error {

	if p.level > TraceLevel {
		return nil
	}

	logText := Format(format, a...)
	fun, filename, lineno := GetRuntimeInfo(p.skip)

	color := colors[TraceLevel]
	logText = color(fmt.Sprintf("[%s:%s:%d] %s", fun, filepath.Base(filename), lineno, logText))

	return p.write(TraceLevel, &logText, logId)
}

func (p *XConsoleLog) Debug(format string, a ...interface{}) error {

	return p.debugx(XConsoleLogDefaultLogId, format, a...)
}

func (p *XConsoleLog) Debugx(logId, format string, a ...interface{}) error {
	return p.debugx(logId, format, a...)
}

//打印debug日志，当日志级别大于Debug时，不会输出任何日志。
func (p *XConsoleLog) debugx(logId, format string, a ...interface{}) error {

	if p.level > DebugLevel {
		return nil
	}

	logText := Format(format, a...)
	fun, filename, lineno := GetRuntimeInfo(p.skip)

	color := colors[DebugLevel]
	logText = color(fmt.Sprintf("[%s:%s:%d] %s", fun, filepath.Base(filename), lineno, logText))

	return p.write(DebugLevel, &logText, logId)
}

//关闭日志库。注意：如果没有调用Close()关闭日志库的话，将会造成文件句柄泄露
func (p *XConsoleLog) Close() {
}

func (p *XConsoleLog) GetHost() string {
	return p.hostname
}

func (p *XConsoleLog) write(level int, msg *string, logId string) error {

	color := colors[level]
	levelText := color(levelTextArray[level])
	time := time.Now().Format("2006-01-02 15:04:05")

	logText := FormatLog(msg, time, p.service, p.hostname, levelText, logId)
	file := os.Stdout
	if level >= WarnLevel {
		file = os.Stderr
	}

	file.Write([]byte(logText))
	return nil
}
