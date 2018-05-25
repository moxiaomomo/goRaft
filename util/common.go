package util

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	REG_WORKER_OK     = 0
	REG_WORKER_FAILED = -1

	HEARTBEAT_INTERVAL = 10
)

type CommonResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type LBPolicyEnum int

const (
	_ LBPolicyEnum = iota
	LB_RANDOM
	LB_ROUNDROBIN
	LB_FASTRESP
)

type LogLevelEnum int

const (
	_ LogLevelEnum = iota
	LOG_DEBUG
	LOG_INFO
	LOG_WARN
	LOG_ERROR
)

func SetStructField(obj interface{}, name string, value interface{}) error {
	structObj := reflect.ValueOf(obj).Elem()
	structField := structObj.FieldByName(name)

	if !structField.IsValid() {
		return fmt.Errorf("Field of struct not found: %s", name)
	}
	if !structField.CanSet() {
		return fmt.Errorf("Field of struct cannot set: %s", name)
	}

	structFieldType := structField.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return fmt.Errorf("Field of struct type mismatch: %s", name)
	}

	structField.Set(val)
	return nil
}

func CurrentDirectory() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return strings.Replace(dir, "\\", "/", -1), nil
}
