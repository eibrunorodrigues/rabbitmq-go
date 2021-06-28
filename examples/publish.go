package examples

import (
	"fmt"

	"github.com/eibrunorodrigues/rabbitmq-go/interfaces"
	"github.com/eibrunorodrigues/rabbitmq-go/rabbitmq"
	"github.com/eibrunorodrigues/rabbitmq-go/types"
)

func createQueue() {
	broker := rabbitmq.Client{}
	broker.Connect()
	result, err := broker.CreateQueue("TESTING_GO", true, false)
	fmt.Println(result, err)
	for item := 1; item <= 100; item++ {
		publisher(&broker, "TESTING_GO")
	}
	broker.Listen("TESTING_GO", receiveMessage)
}

func publisher(br interfaces.IBroker, queue string) {
	br.PublishToQueue([]byte(queue), queue, []types.Filters{types.Filters{Key: "key", Value: "#"}})
}

func receiveMessage(receiver types.Receiver) error {
	fmt.Println(receiver.Body)
	_, _ = receiver.Act.Complete()
	return nil
}
