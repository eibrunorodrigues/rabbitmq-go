package router_type

import (
	"errors"
	"strings"
)

type RouterTypeEnum int

const (
	QUEUE   RouterTypeEnum = 1
	TOPIC   RouterTypeEnum = 2
	HEADERS RouterTypeEnum = 3
	FANOUT  RouterTypeEnum = 4
)

func (e RouterTypeEnum) String() (string, error) {
	switch e {
	case QUEUE:
		return "QUEUE", nil
	case TOPIC:
		return "TOPIC", nil
	case HEADERS:
		return "HEADERS", nil
	case FANOUT:
		return "FANOUT", nil
	default:
		return "", errors.New("Invalid enum " + string(rune(e)))
	}
}



func ParseRouterType(property string) (RouterTypeEnum, error) {
	switch strings.ToUpper(property) {
	case "QUEUE":
		return QUEUE, nil
	case "TOPIC":
		return TOPIC, nil
	case "HEADERS":
		return HEADERS, nil
	case "FANOUT":
		return FANOUT, nil
	default:
		return -1, errors.New("Invalid enum " + property)
	}
}

func RouterType() routerTypeStruct {
	return routerTypeStruct{
		QUEUE: QUEUE,
		TOPIC: TOPIC,
		
	}
}

type routerTypeStruct struct {
	QUEUE RouterTypeEnum
	TOPIC RouterTypeEnum
	HEADERS RouterTypeEnum
	FANOUT RouterTypeEnum
}