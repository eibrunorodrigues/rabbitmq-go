package utils

import (
	"strconv"
)

type Fn func(string) (interface{}, error)

func StrToInt(value string) (interface{}, error) {
	if value == "" {
		return 0, nil
	}

	number, err := strconv.Atoi(value)
	if err != nil {
		return -1, err
	}
	return number, nil
}

func StrToBool(value string) (interface{}, error) {
	if value == "" {
		return false, nil
	}

	response, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}
	return response, nil
}

func StrToFloat32(value string) (interface{}, error) {
	if value == "" {
		return false, nil
	}

	response, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return false, err
	}
	return response, nil
}

func StrToFloat64(value string) (interface{}, error) {
	if value == "" {
		return false, nil
	}

	response, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false, err
	}
	return response, nil
}
