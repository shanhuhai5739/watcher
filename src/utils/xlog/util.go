package xlog

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

const (
	DebugLevel = iota
	TraceLevel
	NoticeLevel
	WarnLevel
	FatalLevel
	NoneLevel
)

const (
	XLogDefSkipNum = 4
)

var (
	levelTextArray = []string{
		DebugLevel:  "DEBUG",
		TraceLevel:  "TRACE",
		NoticeLevel: "NOTICE",
		WarnLevel:   "WARN",
		FatalLevel:  "FATAL",
	}
)

func LevelFromStr(level string) int {

	resultLevel := DebugLevel
	levelLower := strings.ToLower(level)
	switch levelLower {
	case "debug":
		resultLevel = DebugLevel
	case "trace":
		resultLevel = TraceLevel
	case "notice":
		resultLevel = NoticeLevel
	case "warn":
		resultLevel = WarnLevel
	case "fatal":
		resultLevel = FatalLevel
	case "none":
		resultLevel = NoneLevel
	default:
		resultLevel = NoticeLevel
	}

	return resultLevel
}

func GetRuntimeInfo(skip int) (function, filename string, lineno int) {

	function = "???"
	pc, filename, lineno, ok := runtime.Caller(skip)
	if ok {
		function = runtime.FuncForPC(pc).Name()
	}

	return
}

func FormatLog(body *string, fields ...string) string {

	var buffer bytes.Buffer
	for _, v := range fields {
		buffer.WriteString("[")
		buffer.WriteString(v)
		buffer.WriteString("] ")
	}

	buffer.WriteString(*body)
	buffer.WriteString("\n")

	return buffer.String()
}

func Format(format string, a ...interface{}) (result string) {

	if len(a) == 0 {
		result = format
		return
	}

	result = fmt.Sprintf(format, a...)
	return
}
