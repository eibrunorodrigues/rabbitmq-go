package enums

import (
	"errors"
	"strings"
)

type RouterTypeEnum int

var (
	typeDirect   RouterTypeEnum = 1
	typeTopic   RouterTypeEnum = 2
	typeHeaders RouterTypeEnum = 3
	typeFanout  RouterTypeEnum = 4
)

var RouterType = struct {
	DIRECT RouterTypeEnum
	TOPIC RouterTypeEnum
	HEADERS RouterTypeEnum
	FANOUT RouterTypeEnum
}{
	DIRECT: typeDirect,
	TOPIC: typeTopic,
	HEADERS: typeHeaders,
	FANOUT: typeFanout,
}


func (e RouterTypeEnum) String() (string, error) {
	switch e {
	case RouterType.DIRECT:
		return "DIRECT", nil
	case RouterType.TOPIC:
		return "TOPIC", nil
	case RouterType.HEADERS:
		return "HEADERS", nil
	case RouterType.FANOUT:
		return "FANOUT", nil
	default:
		return "", errors.New("Invalid enum " + string(rune(e)))
	}
}

func ParseRouterType(property string) (RouterTypeEnum, error) {
	switch strings.ToUpper(property) {
	case "DIRECT":
		return RouterType.DIRECT, nil
	case "TOPIC":
		return RouterType.TOPIC, nil
	case "HEADERS":
		return RouterType.HEADERS, nil
	case "FANOUT":
		return RouterType.FANOUT, nil
	default:
		return -1, errors.New("Invalid enum " + property)
	}
}
