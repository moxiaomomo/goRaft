package logger

import (
	"fmt"

	"github.com/moxiaomomo/goRaft/util"
)

var logLevel util.LogLevelEnum = util.LOG_INFO

// SetLogLevel SetLogLevel
func SetLogLevel(level util.LogLevelEnum) {
	logLevel = level
}

// Debug Debug level log
func Debug(a ...interface{}) (n int, err error) {
	if logLevel > util.LOG_DEBUG {
		return 0, nil
	}
	str := []interface{}{"[DEBUG]"}
	return fmt.Println(append(str, a)...)
}

// Info Info level log
func Info(a ...interface{}) (n int, err error) {
	if logLevel > util.LOG_INFO {
		return 0, nil
	}
	str := []interface{}{"[INFO]"}
	return fmt.Println(append(str, a)...)
}

// Warn Warn level log
func Warn(a ...interface{}) (n int, err error) {
	if logLevel > util.LOG_WARN {
		return 0, nil
	}
	str := []interface{}{"[WARN]"}
	return fmt.Println(append(str, a)...)
}

// Error Error level log
func Error(a ...interface{}) (n int, err error) {
	str := []interface{}{"[ERROR]"}
	return fmt.Println(append(str, a)...)
}

// Debugf Debugf level log
func Debugf(format string, a ...interface{}) (n int, err error) {
	if logLevel > util.LOG_DEBUG {
		return 0, nil
	}

	return fmt.Printf("[DEBUG] "+format, a...)
}

// Infof Infof level log
func Infof(format string, a ...interface{}) (n int, err error) {
	if logLevel > util.LOG_INFO {
		return 0, nil
	}
	return fmt.Printf("[INFO] "+format, a...)
}

// Warnf Warnf level log
func Warnf(format string, a ...interface{}) (n int, err error) {
	if logLevel > util.LOG_WARN {
		return 0, nil
	}
	return fmt.Printf("[WARN] "+format, a...)
}

// Errorf Errorf  level log
func Errorf(format string, a ...interface{}) (n int, err error) {
	return fmt.Printf("[ERROR] "+format, a...)
}
