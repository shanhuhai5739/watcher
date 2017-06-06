package xlog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type XFileLog struct {
	filename string
	path     string
	level    int

	skip     int
	file     *os.File
	errFile  *os.File
	hostname string
	service  string
	split    sync.Once
	mu       sync.Mutex
}

const (
	XFileLogDefaultLogId = "900000001"
	SpliterDelay         = 5
	CleanDays            = -3
	SplitPath            = "/home/work/logs/applogs/"
)

func init() {
	RegisterLogger("file", NewXFileLog())
}

//生成一个日志实例，service用来标识业务的服务名。
//比如：logger := xlog.NewXFileLog("shopapi")
func NewXFileLog() XLogInterface {
	return &XFileLog{
		skip: XLogDefSkipNum,
	}
}

func (p *XFileLog) Init(config map[string]string) (err error) {

	path, ok := config["path"]
	if !ok {
		err = errors.New(fmt.Sprintf("init XFileLog failed, not found path"))
		return
	}

	filename, ok := config["filename"]
	if !ok {
		err = errors.New(fmt.Sprintf("init XFileLog failed, not found filename"))
		return
	}

	level, ok := config["level"]
	if !ok {
		err = errors.New(fmt.Sprintf("init XFileLog failed, not found level"))
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

	isDir, err := isDir(path)
	if err != nil || !isDir {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("Mkdir failed, err:%v", err)
		}
	}

	p.path = path
	p.filename = filename
	p.level = LevelFromStr(level)

	hostname, _ := os.Hostname()
	p.hostname = hostname
	body := func() {
		go p.Spliter()
	}
	doSplit, ok := config["dosplit"]
	if !ok {
		doSplit = "true"
	}
	if doSplit == "true" {
		p.split.Do(body)
	}
	return p.ReOpen()
}

func (p *XFileLog) Spliter() {
	preHour := time.Now().Hour()
	splitTime := time.Now().Format("2006010215")
	defer p.Close()
	for {
		time.Sleep(time.Second * SpliterDelay)
		if time.Now().Hour() != preHour {
			p.Clean()
			p.ReName(splitTime)
			preHour = time.Now().Hour()
			splitTime = time.Now().Format("2006010215")
		}
	}
}

//@title 设置日志级别
//@level：日志级别，如下："Debug", "Trace", "Notice", "Warn", "Fatal", "None"
func (p *XFileLog) SetLevel(level string) {
	p.level = LevelFromStr(level)
}

func (p *XFileLog) SetSkip(skip int) {
	p.skip = skip
}

func (p *XFileLog) openFile(filename string) (*os.File, error) {

	file, err := os.OpenFile(filename,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)

	if err != nil {
		return nil, fmt.Errorf("open %s failed, err:%v", filename, err)
	}

	return file, err
}

func delayClose(fp *os.File) {
	if fp == nil {
		return
	}
	time.Sleep(1000 * time.Millisecond)
	fp.Close()
}
func (p *XFileLog) Clean() (err error) {
	deadline := time.Now().AddDate(0, 0, CleanDays)
	var path string
	if strings.HasPrefix(p.path, SplitPath) {
		path = SplitPath
	} else {
		path = p.path
	}
	var files []string
	files, err = filepath.Glob(fmt.Sprintf("%s/%s.log*", path, p.filename))
	if err != nil {
		return
	}
	var fileInfo os.FileInfo
	for _, file := range files {
		if filepath.Base(file) == fmt.Sprintf("%s.log", p.filename) {
			continue
		}
		if filepath.Base(file) == fmt.Sprintf("%s.log.wf", p.filename) {
			continue
		}
		if fileInfo, err = os.Stat(file); err == nil {
			if fileInfo.ModTime().Before(deadline) {
				os.Remove(file)
			} else if fileInfo.Size() == 0 {
				os.Remove(file)
			}
		}
	}
	return
}
func (p *XFileLog) ReName(shuffix string) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	defer p.ReOpen()
	if p.file == nil {
		return
	}
	var fileInfo os.FileInfo
	var path string
	normalLog := p.path + "/" + p.filename + ".log"
	warnLog := normalLog + ".wf"

	if strings.HasPrefix(p.path, SplitPath) {
		path = SplitPath
	} else {
		path = p.path
	}
	newLog := fmt.Sprintf("%s/%s.log-%s", path, p.filename, shuffix)
	newWarnLog := fmt.Sprintf("%s/%s.log.wf-%s", path, p.filename, shuffix)
	if fileInfo, err = os.Stat(normalLog); err == nil && fileInfo.Size() == 0 {
		return
	}
	if _, err = os.Stat(newLog); err == nil {
		return
	}
	if err = os.Rename(normalLog, newLog); err != nil {
		return
	}
	if fileInfo, err = os.Stat(warnLog); err == nil && fileInfo.Size() == 0 {
		return
	}
	if _, err = os.Stat(newWarnLog); err == nil {
		return
	}
	if err = os.Rename(warnLog, newWarnLog); err != nil {
		return
	}
	return
}

func (p *XFileLog) ReOpen() error {

	go delayClose(p.file)
	go delayClose(p.errFile)

	normalLog := p.path + "/" + p.filename + ".log"
	file, err := p.openFile(normalLog)
	if err != nil {
		return err
	}

	p.file = file
	warnLog := normalLog + ".wf"
	p.errFile, err = p.openFile(warnLog)
	if err != nil {
		p.file.Close()
		p.file = nil
		return err
	}

	return nil
}

//打印warn日志，当日志级别大于Warn时，不会输出任何日志。
func (p *XFileLog) Warn(format string, a ...interface{}) error {

	if p.level > WarnLevel {
		return nil
	}

	return p.warnx(XFileLogDefaultLogId, format, a...)
}

//打印warn日志，当日志级别大于Warn时，不会输出任何日志。
func (p *XFileLog) Warnx(logId, format string, a ...interface{}) error {

	if p.level > WarnLevel {
		return nil
	}

	return p.warnx(logId, format, a...)
}

//打印warn日志，当日志级别大于Warn时，不会输出任何日志。
func (p *XFileLog) warnx(logId, format string, a ...interface{}) error {

	logText := Format(format, a...)
	fun, filename, lineno := GetRuntimeInfo(p.skip)
	logText = fmt.Sprintf("[%s:%s:%d] %s", fun, filepath.Base(filename), lineno, logText)

	return p.write(WarnLevel, &logText, logId)
}

//打印fatal日志，当日志级别大于Fatal时，不会输出任何日志。
func (p *XFileLog) Fatal(format string, a ...interface{}) error {

	if p.level > FatalLevel {
		return nil
	}

	return p.fatalx(XFileLogDefaultLogId, format, a...)
}

//打印fatal日志，当日志级别大于Fatal时，不会输出任何日志。
func (p *XFileLog) Fatalx(logId, format string, a ...interface{}) error {

	if p.level > FatalLevel {
		return nil
	}

	return p.fatalx(logId, format, a...)
}

func (p *XFileLog) fatalx(logId, format string, a ...interface{}) error {

	logText := Format(format, a...)
	fun, filename, lineno := GetRuntimeInfo(p.skip)
	logText = fmt.Sprintf("[%s:%s:%d] %s", fun, filepath.Base(filename), lineno, logText)

	return p.write(FatalLevel, &logText, logId)
}

//打印notice日志，当日志级别大于Notice时，不会输出任何日志。
func (p *XFileLog) Notice(format string, a ...interface{}) error {

	return p.Noticex(XFileLogDefaultLogId, format, a...)
}

//打印notice日志，当日志级别大于Notice时，不会输出任何日志。
func (p *XFileLog) Noticex(logId, format string, a ...interface{}) error {

	if p.level > NoticeLevel {
		return nil
	}

	logText := Format(format, a...)
	return p.write(NoticeLevel, &logText, logId)
}

//打印trace日志，当日志级别大于Trace时，不会输出任何日志。
func (p *XFileLog) Trace(format string, a ...interface{}) error {

	return p.tracex(XFileLogDefaultLogId, format, a...)
}

//打印trace日志，当日志级别大于Trace时，不会输出任何日志。
func (p *XFileLog) Tracex(logId, format string, a ...interface{}) error {

	return p.tracex(logId, format, a...)
}

//打印trace日志，当日志级别大于Trace时，不会输出任何日志。
func (p *XFileLog) tracex(logId, format string, a ...interface{}) error {

	if p.level > TraceLevel {
		return nil
	}

	logText := Format(format, a...)
	fun, filename, lineno := GetRuntimeInfo(p.skip)
	logText = fmt.Sprintf("[%s:%s:%d] %s", fun, filepath.Base(filename), lineno, logText)

	return p.write(TraceLevel, &logText, logId)
}

//打印debug日志，当日志级别大于Debug时，不会输出任何日志。
func (p *XFileLog) Debug(format string, a ...interface{}) error {

	return p.debugx(XFileLogDefaultLogId, format, a...)
}

//打印warn日志，当日志级别大于Warn时，不会输出任何日志。
func (p *XFileLog) debugx(logId, format string, a ...interface{}) error {

	if p.level > DebugLevel {
		return nil
	}

	logText := Format(format, a...)
	fun, filename, lineno := GetRuntimeInfo(p.skip)
	logText = fmt.Sprintf("[%s:%s:%d] %s", fun, filepath.Base(filename), lineno, logText)

	return p.write(DebugLevel, &logText, logId)
}

//打印debug日志，当日志级别大于Debug时，不会输出任何日志。
func (p *XFileLog) Debugx(logId, format string, a ...interface{}) error {

	return p.debugx(logId, format, a...)
}

//关闭日志库。注意：如果没有调用Close()关闭日志库的话，将会造成文件句柄泄露
func (p *XFileLog) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.file != nil {
		p.file.Close()
		p.file = nil
	}

	if p.errFile != nil {
		p.errFile.Close()
		p.errFile = nil
	}
}

func (p *XFileLog) GetHost() string {
	return p.hostname
}

func (p *XFileLog) write(level int, msg *string, logId string) error {

	levelText := levelTextArray[level]
	time := time.Now().Format("2006-01-02 15:04:05")

	logText := FormatLog(msg, time, p.service, p.hostname, levelText, logId)
	file := p.file
	if level >= WarnLevel {
		file = p.errFile
	}

	file.Write([]byte(logText))
	return nil
}

func isDir(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}
