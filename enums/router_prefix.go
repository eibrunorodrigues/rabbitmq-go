package enums

import (
	"errors"
	"strings"
)

type RouterPrefixEnum int

var (
	prefixQueue   RouterPrefixEnum = 1
	prefixTopic   RouterPrefixEnum = 2
	prefixHeaders RouterPrefixEnum = 3
)

var RouterPrefix = struct {
	QUEUE   RouterPrefixEnum
	TOPIC   RouterPrefixEnum
	HEADERS RouterPrefixEnum
}{
	QUEUE:   prefixQueue,
	TOPIC:   prefixTopic,
	HEADERS: prefixHeaders,
}

func (e RouterPrefixEnum) String() (string, error) {
	switch e {
	case RouterPrefix.QUEUE:
		return "QUEUE", nil
	case RouterPrefix.TOPIC:
		return "TOPIC", nil
	case RouterPrefix.HEADERS:
		return "HEADERS", nil
	default:
		return "", errors.New("Invalid enum " + string(rune(e)))
	}
}

func ParseRouterPrefix(property string) (RouterPrefixEnum, error) {
	switch strings.ToUpper(property) {
	case "QUEUE":
		return RouterPrefix.QUEUE, nil
	case "TOPIC":
		return RouterPrefix.TOPIC, nil
	case "HEADERS":
		return RouterPrefix.HEADERS, nil
	default:
		return -1, errors.New("Invalid enum " + property)
	}
}
