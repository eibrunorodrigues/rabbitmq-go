package rabbitmq

import (
	"errors"
	"fmt"
	"github.com/eibrunorodrigues/rabbitmq-go/constants"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/eibrunorodrigues/rabbitmq-go/enums"

	"github.com/eibrunorodrigues/rabbitmq-go/types"
	"github.com/eibrunorodrigues/rabbitmq-go/utils"

	"github.com/rabbitmq/amqp091-go"
)

// Client layer following github.com/eibrunorodrigues/rabbitmq-go/interfaces/broker.go interface
type Client struct {
	localConnection           *amqp091.Connection
	channel                   *amqp091.Channel
	reconnectRoutineIsRunning bool
	reconnectAttempts         int
	isListening               bool
	consumerTag               string
	Config                    Configs
}

// Configs are the Connection parameters
type Configs struct {
	Host             string
	Port             int
	Vhost            string
	User             string
	Pass             string
	Heartbeat        int
	SendRetryMax     int
	ReplayDelayMs    int
	Prefetch         int
	ReconnectAttemps int
}

type Actions struct {
	message      *amqp091.Delivery
	wasCompleted bool
	wasAbandoned bool
	wasRequeued  bool
}

func (r *Client) connect() (*amqp091.Connection, error) {
	if r.Config.Host == "" {
		r.Config = Configs{
			Host:             utils.GetTypedEnvVariable("BROKER_HOST", "localhost", reflect.String).(string),
			Port:             utils.GetTypedEnvVariable("BROKER_PORT", 5672, reflect.Int).(int),
			Vhost:            utils.GetTypedEnvVariable("BROKER_VHOST", "/", reflect.String).(string),
			User:             utils.GetTypedEnvVariable("BROKER_USER", "admin", reflect.String).(string),
			Pass:             utils.GetTypedEnvVariable("BROKER_PASS", "admin", reflect.String).(string),
			Heartbeat:        utils.GetTypedEnvVariable("BROKER_HEARTBEAT", 600, reflect.Int).(int),
			SendRetryMax:     utils.GetTypedEnvVariable("BROKER_SEND_RETRY_MAX", 3, reflect.Int).(int),
			ReplayDelayMs:    utils.GetTypedEnvVariable("BROKER_REPLAY_DELAY_MS", 60000, reflect.Int).(int),
			Prefetch:         utils.GetTypedEnvVariable("BROKER_PREFETCH", 100, reflect.Int).(int),
			ReconnectAttemps: utils.GetTypedEnvVariable("BROKER_RECONNECT_ATTEMPS", 5, reflect.Int).(int),
		}
	}

	if r.Config.Host == "" {
		log.Fatalf("rabbitmq: please provide a truly valid BROKER_URI in your env")
	}

	portString := strconv.Itoa(r.Config.Port)

	conn, err := amqp091.Dial("amqp://" + r.Config.User + ":" + r.Config.Pass + "@" + r.Config.Host + ":" + portString + "/" + r.Config.Vhost)
	if conn != nil {
		return conn, err
	}

	return &amqp091.Connection{}, err
}

// Connect connects or reconnects to Client
func (r *Client) Connect() *amqp091.Channel {
	if r.localConnection == nil || r.localConnection.IsClosed() {
		conn, err := r.connect()
		if err != nil {
			if r.reconnectAttempts < r.Config.ReconnectAttemps {
				r.reconnectAttempts++
				r.localConnection = nil
				fmt.Printf("\nrabbitmq: connection attempt failed... retry %d/%d: %v", r.reconnectAttempts, r.Config.ReconnectAttemps, err)
				time.Sleep(2 * time.Second)
				return r.Connect()
			}
			panic(err)
		}
		r.localConnection = conn
		r.reconnectAttempts = 0
	}

	if r.channel == nil || r.channel.IsClosed() {
		r.channel = r.makeChannel()
	}

	return r.channel
}

// IsOpen verifies if the connection and channel are open
func (r *Client) IsOpen() bool {
	return !r.localConnection.IsClosed() && !r.channel.IsClosed()
}

//AcknowledgeMessage lets Rabbit knows that you received successfully a message
//and removes it from a queue
func (r *Client) AcknowledgeMessage(messageID int) error {
	return r.Connect().Ack(uint64(messageID), false)
}

//RejectMessage lets Rabbit knows that you received successfully a message
//and removes it from a queue
func (r *Client) RejectMessage(messageID int, requeue bool) error {
	return r.Connect().Reject(uint64(messageID), requeue)
}

//CheckIfQueueExists Passive Declares a Queue. If an error with "not_found"
//is thrown, then the queue doesnt exist.
func (r *Client) CheckIfQueueExists(queueName string) bool {
	_, err := r.Connect().QueueDeclarePassive(queueName, true, false, false, false, amqp091.Table{})
	if err != nil && strings.Contains(err.Error(), "404") {
		return false
	}
	return true
}

//CheckIfRouterExists Passive Declares a Router. If an error with "not_found"
//is returned, then the router doesnt exist.
func (r *Client) CheckIfRouterExists(routerName string) bool {
	exchangeType, _ := enums.RouterType.TOPIC.String()
	err := r.Connect().ExchangeDeclarePassive(routerName, strings.ToLower(exchangeType), true, false, false, false, amqp091.Table{})
	if err != nil && strings.Contains(err.Error(), "404") {
		return false
	}
	return true
}

//CreateQueue creates a fancy queue (with dlq, exchanges and binds) and returns the name
func (r *Client) CreateQueue(queueName string, createDlq bool, exclusive bool) (string, error) {
	if err := validateQueueName(queueName); err != nil {
		return "", err
	}

	queueName = strings.ToUpper(queueName)
	routerName, err := r.CreateRouter(queueName, enums.RouterPrefix.QUEUE, enums.RouterType.DIRECT)

	if err != nil {
		return "", err
	}

	var queueArgs []types.Filters
	dlqQueue := queueName + ".delay"

	if createDlq {

		var dlqArgs = []types.Filters{
			{Key: "x-message-ttl", Value: r.Config.ReplayDelayMs},
			{Key: "x-dead-letter-exchange", Value: routerName},
			{Key: "x-dead-letter-routing-key", Value: ""},
		}

		_, err := r.Connect().QueueDeclare(dlqQueue, true, false, false, false, filtersToTable(dlqArgs))

		if err != nil {
			return "", err
		}

		if _, err := r.BindQueueToRouter(dlqQueue, routerName, "delay"); err != nil {
			return "", err
		}

		queueArgs = append(queueArgs, types.Filters{Key: "x-dead-letter-exchange", Value: routerName})
		queueArgs = append(queueArgs, types.Filters{Key: "x-dead-letter-routing-key", Value: "delay"})
	}

	_, err = r.Connect().QueueDeclare(queueName, true, false, exclusive, false, amqp091.Table{})

	if err != nil {
		if _, err := r.Connect().QueueDeclare(queueName, false, false, exclusive, false, amqp091.Table{}); err != nil {
			return "", err
		}
		fmt.Printf("\nrabbitmq: wrong durable... creating queue with flag durable: false")
	}
	if !exclusive {
		if _, err := r.BindQueueToRouter(queueName, routerName, ""); err != nil {
			return "", err
		}
	}
	return queueName, nil
}

//CreateRouter creates an Exchange and returns the formatted name
func (r *Client) CreateRouter(routerName string, prefix enums.RouterPrefixEnum, routerType enums.RouterTypeEnum) (string, error) {
	routerName, err := validateRouterName(routerName, prefix)
	if err != nil {
		return "", err
	}

	routerTypeString, err := routerType.String()
	if err != nil {
		return "", err
	}

	err = r.Connect().ExchangeDeclare(routerName, strings.ToLower(routerTypeString), true, false, false, false, amqp091.Table{})
	if err != nil {
		fmt.Printf("\nrabbitmq: exchange with wrong durable")
		if err := r.Connect().ExchangeDeclare(routerName, strings.ToLower(routerTypeString), false, false, false, false, amqp091.Table{}); err != nil {
			return "", err
		}
	}

	return routerName, nil
}

//PublishToQueue Publishes a message to a queue and return if it published successfully.
func (r *Client) PublishToQueue(message []byte, queueName string, filters []types.Filters) (bool, error) {
	if _, err := validateFiltersArg(filters); err != nil {
		return false, err
	}

	return r.publishMessage(message, "", queueName, filters, 0)
}

//PublishToRouter Publishes a message to a router and return if it published
//successfully.
func (r *Client) PublishToRouter(message []byte, routerName string, filters interface{}) (bool, error) {
	if _, err := validateFiltersArg(filters); err != nil {
		return false, err
	}

	return r.publishMessage(message, routerName, "", filters, 0)
}

//DeleteQueue deletes an existing queue and return if it the operation was
//successfully completed.
func (r *Client) DeleteQueue(queueName string) (bool, error) {
	_, err := r.Connect().QueueDelete(queueName, false, false, false)

	if err != nil {
		return false, err
	}

	return true, nil
}

//DeleteRouter deletes an existing router and return if it the operation was
//successfully completed.
func (r *Client) DeleteRouter(routerName string) (bool, error) {
	err := r.Connect().ExchangeDelete(routerName, false, false)

	if err != nil {
		return false, err
	}

	return true, nil
}

//ForceClose method closes connection.
func (r *Client) ForceClose() {
	if !r.localConnection.IsClosed() {
		_ = r.localConnection.Close()
	}
}

//Close method closes the channel.
func (r *Client) Close() error {
	err := r.channel.Close()
	if err != nil {
		return err
	}
	return nil
}

//HealthCheck method checks the current channel and connection status.
func (r *Client) HealthCheck() bool {
	return !r.localConnection.IsClosed()
}

//BindQueueToRouter binds a queue to an exchange.
func (r *Client) BindQueueToRouter(queueName string, routerName string, filters interface{}) (bool, error) {
	var err error = nil
	switch filters.(type) {
	case string:
		err = r.Connect().QueueBind(queueName, filters.(string), routerName, false, amqp091.Table{})
	case []types.Filters:
		err = r.Connect().QueueBind(queueName, "", routerName, false, filtersToTable(filters.([]types.Filters)))
	default:
		return false, errors.New("rabbitmq: invalid filters type argument")
	}
	return true, err
}

//BindRouterToRouter binds a router to an exchange.
func (r *Client) BindRouterToRouter(destination string, source string, filters interface{}) (bool, error) {
	var err error = nil
	switch filters.(type) {
	case string:
		err = r.Connect().ExchangeBind(destination, filters.(string), source, false, amqp091.Table{})
	case []types.Filters:
		err = r.Connect().ExchangeBind(destination, "", source, false, filtersToTable(filters.([]types.Filters)))
	default:
		return false, errors.New("rabbitmq: invalid filters type argument")
	}
	return true, err
}

//Listen starts consuming a queue and calls a callback function to wait for
//success on internal operation
func (r *Client) Listen(queueName string, receiverCallback types.ReceiverCallback) error {
	consumer := fmt.Sprintf("%s-%s", queueName, ksuid.New().String())
	messages, err := r.Connect().Consume(queueName, consumer, false, false, false, false, nil)
	if err != nil {
		return err
	}

	r.isListening = true

	for message := range messages {
		r.consumerTag = message.ConsumerTag

		receiverModel := GetReceiverModel(message)

		if err := receiverCallback(receiverModel); err != nil {
			return fmt.Errorf("rabbitmq: error receiving event. %v", err)
		}

	}
	return nil
}

//GetReceiverModel receives an amqp091.Delivery and returns a types.Receiver
func GetReceiverModel(message amqp091.Delivery) types.Receiver {
	receiverModel := types.Receiver{}
	receiverModel.Act = &Actions{message: &message}

	receiverModel.Filters = append(receiverModel.Filters, types.Filters{Key: "routing-key", Value: message.RoutingKey})

	if message.Headers["x-first-death-exchange"] != nil {
		receiverModel.RouterOrigin = message.Headers["x-first-death-exchange"].(string)
	} else if message.Exchange != "" {
		receiverModel.RouterOrigin = message.Exchange
	}

	receiverModel.MessageId = int(message.DeliveryTag)
	receiverModel.IsARedelivery = checkIfIsARedelivery(message)
	receiverModel.Body = message.Body

	return receiverModel
}

//StopListening stops consuming a queue
func (r *Client) StopListening() (bool, error) {
	if err := r.Connect().Cancel(r.consumerTag, true); err != nil {
		return false, err
	}
	return true, nil
}

func checkIfIsARedelivery(message amqp091.Delivery) bool {
	if message.Headers["x-death"] != nil {
		return true
	}

	if message.Redelivered {
		return true
	}

	return false
}

func validateQueueName(queueName string) error {
	isMatch, err := regexp.MatchString(constants.QueueNameRule, queueName)
	if err != nil {
		return err
	}

	_, err = utils.StrToInt(queueName)

	if queueName == "" || !isMatch || err == nil {
		return errors.New("rabbitmq: invalid queueName " + queueName + "... rule: " + constants.QueueNameRule)
	}

	return nil
}

func validateRouterName(routerName string, prefix enums.RouterPrefixEnum) (string, error) {
	if routerName == "" {
		return "", errors.New("rabbitmq: empty routerName found")
	}

	isAFullRouterName, err := regexp.MatchString(constants.FullRouterNameRule, routerName)
	if err != nil {
		return "", err
	}

	if isAFullRouterName {
		var strPrefix = strings.Split(routerName, "/")[0]
		if _, err = enums.ParseRouterPrefix(strings.ToUpper(strPrefix)); err == nil {
			return routerName, nil
		}
		return "", errors.New("rabbitmq: default prefix is not allowed. " + strPrefix)
	}

	if matched, err := regexp.MatchString(constants.RouterNameRule, routerName); err == nil && matched {
		prefixString, _ := prefix.String()
		return strings.ToUpper(prefixString) + "/" + routerName + ".master", nil
	}

	return "", errors.New("rabbitmq: invalid routerName found... rule: " + constants.RouterNameRule)
}

func validateFiltersArg(filters interface{}) (bool, error) {
	switch filters.(type) {
	case string:
		return true, nil
	case []types.Filters:
		return true, nil
	default:
		return false, errors.New("rabbitmq: invalid filter passed")
	}
}

func filtersToTable(filters []types.Filters) amqp091.Table {
	var table = amqp091.Table{}
	for _, item := range filters {
		table[item.Key] = item.Value
	}
	return table
}

func (r *Client) publishMessage(message []byte, exchange string, routingKey string, filters interface{}, retries int) (bool, error) {
	publishingMsg := amqp091.Publishing{Body: message, DeliveryMode: 2, Timestamp: time.Now()}

	switch filters.(type) {
	case string:
		routingKey = filters.(string)
		break
	case []types.Filters:
		publishingMsg.Headers = filtersToTable(filters.([]types.Filters))
		break
	default:
		return false, errors.New("rabbitmq: invalid filters type")
	}

	err := r.Connect().Publish(exchange, routingKey, false, false, publishingMsg)
	if err != nil {
		if retries <= r.Config.SendRetryMax {
			return r.publishMessage(message, exchange, routingKey, filters, retries+1)
		}
		return false, err
	}
	return true, nil
}

func (r *Client) makeChannel() *amqp091.Channel {
	ch, err := r.localConnection.Channel()

	if err != nil {
		if r.reconnectAttempts < r.Config.ReconnectAttemps {
			fmt.Printf("\nbroker: couldn't create channel... Attempt %d/%d...", r.reconnectAttempts, r.Config.ReconnectAttemps)
			r.reconnectAttempts++
			time.Sleep(2 * time.Second)
			return r.makeChannel()
		}
		panic(err)
	}
	r.reconnectAttempts = 0
	return ch
}

func (a *Actions) Complete() (bool, error) {
	err := a.message.Ack(false)
	if err != nil {
		return false, fmt.Errorf("rabbitmq: error while acking event %v", err)
	}
	a.wasCompleted = true
	return true, nil
}

func (a *Actions) Abandon() (bool, error) {
	err := a.message.Reject(false)
	if err != nil {
		return false, fmt.Errorf("rabbitmq: error while abandoning event %v", err)
	}
	a.wasAbandoned = true
	return true, nil
}

func (a *Actions) Requeue() (bool, error) {
	err := a.message.Nack(false, true)
	if err != nil {
		return false, fmt.Errorf("rabbitmq: error while requeueing event %v", err)
	}
	a.wasRequeued = true
	return true, nil
}
