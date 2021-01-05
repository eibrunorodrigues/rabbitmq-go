package utils

import (
	"os"
	"reflect"
)

func GetTypedEnvVariable(env string, defaultValue interface{}, envType interface{}) interface{} {
	variable := os.Getenv(env)
	if variable == "" {
		return defaultValue
	}

	switch envType {
	case reflect.Int:
		return convertType(variable, StrToInt)
	case reflect.Bool:
		return convertType(variable, StrToBool)
	case reflect.String:
		return variable
	case reflect.Float32:
		return convertType(variable, StrToFloat32)
	case reflect.Float64:
		return convertType(variable, StrToFloat64)
	default:
		panic("type not allowed")
	}
}

func convertType(value string, callFunc Fn) interface{} {
	convertedValue, err := callFunc(value)
	if err != nil {
		panic(err)
	}
	return convertedValue
}
