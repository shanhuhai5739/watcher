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

type XLogInterface interface {
	Init(config map[string]string) error
	ReOpen() error
	SetLevel(level string)
	SetSkip(skip int)

	Warn(format string, a ...interface{}) error
	Fatal(format string, a ...interface{}) error
	Notice(format string, a ...interface{}) error
	Trace(format string, a ...interface{}) error
	Debug(format string, a ...interface{}) error

	Warnx(logId, format string, a ...interface{}) error
	Fatalx(logId, format string, a ...interface{}) error
	Noticex(logId, format string, a ...interface{}) error
	Tracex(logId, format string, a ...interface{}) error
	Debugx(logId, format string, a ...interface{}) error

	Close()
}
