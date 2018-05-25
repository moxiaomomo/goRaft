package logger

import (
	"fmt"
	"gomh/util"
)

var logLevel util.LogLevelEnum = util.LOG_INFO

func SetLogLevel(level util.LogLevelEnum) {
	logLevel = level
}

func LogDebug(a ...interface{}) (n int, err error) {
	if logLevel > util.LOG_DEBUG {
		return 0, nil
	}
	fmt.Print("[DEBUG]")
	return fmt.Println(a...)
}

func LogInfo(a ...interface{}) (n int, err error) {
	if logLevel > util.LOG_INFO {
		return 0, nil
	}
	fmt.Print("[INFO]")
	return fmt.Println(a...)
}

func LogWarn(a ...interface{}) (n int, err error) {
	if logLevel > util.LOG_WARN {
		return 0, nil
	}
	fmt.Print("[WARN]")
	return fmt.Println(a...)
}

func LogError(a ...interface{}) (n int, err error) {
	fmt.Print("[ERROR]")
	return fmt.Println(a...)
}

func LogDebugf(format string, a ...interface{}) (n int, err error) {
	if logLevel > util.LOG_DEBUG {
		return 0, nil
	}
	fmt.Print("[DEBUG]")
	return fmt.Printf(format, a...)
}

func LogInfof(format string, a ...interface{}) (n int, err error) {
	if logLevel > util.LOG_INFO {
		return 0, nil
	}
	fmt.Print("[INFO]")
	return fmt.Printf(format, a...)
}

func LogWarnf(format string, a ...interface{}) (n int, err error) {
	if logLevel > util.LOG_WARN {
		return 0, nil
	}
	fmt.Print("[WARN]")
	return fmt.Printf(format, a...)
}

func LogErrorf(format string, a ...interface{}) (n int, err error) {
	fmt.Print("[ERROR]")
	return fmt.Printf(format, a...)
}
