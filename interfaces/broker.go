package interfaces

import (
	"github.com/eibrunorodrigues/rabbitmq-go/enums"
	"github.com/eibrunorodrigues/rabbitmq-go/types"
)

//IBroker is an interface for RabbitMQ.Client usage
type IBroker interface {
	IsOpen() bool
	AcknowledgeMessage(messageID int)
	RejectMessage(messageID int, requeue bool)
	CheckIfQueueExists(queueName string) bool
	CheckIfRouterExists(routerName string) bool
	CreateQueue(queueName string, createDlq bool, exclusive bool) (string, error)
	CreateRouter(routerName string, prefix enums.RouterPrefixEnum, routerType enums.RouterTypeEnum) (string, error)
	PublishToQueue(message []byte, queueName string, filters []types.Filters) (bool, error)
	PublishToRouter(message []byte, routerName string, filters interface{}) (bool, error)
	DeleteQueue(queueName string) (bool, error)
	DeleteRouter(routerName string) (bool, error)
	Close() error
	ForceClose()
	HealthCheck() bool
	BindQueueToRouter(queueName string, routerName string, filters interface{}) (bool, error)
	BindRouterToRouter(destination string, source string, filters interface{}) (bool, error)
	Listen(queueName string, receiverCallback types.ReceiverCallback) error
	StopListening() (bool, error)
}
