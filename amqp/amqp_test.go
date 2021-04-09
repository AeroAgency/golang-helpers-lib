package amqp

import (
	"github.com/rs/zerolog"
	rabbitLib "github.com/streadway/amqp"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})
	if reflect.TypeOf(client).String() != "*amqp.Client" {
		t.Fatalf("the return result is of type '%s' expected '*amqp.Client'", reflect.TypeOf(client).String())
	}
}

func TestConsumer_SetTimeout(t *testing.T) {
	testTimeout := 10 * time.Second
	consumer := Consumer{}
	consumer.SetTimeout(testTimeout)

	if consumer.Timeout != testTimeout {
		t.Fatalf("Timeout property is not set to %s", testTimeout)
	}
	if !consumer.deadline.After(time.Now()) {
		t.Fatalf("deadline property must be greater than the current timestamp - %s ", consumer.deadline)
	}
}

func TestClient_Connect(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})
	err := client.connect()
	if err == nil {
		t.Fatalf("an error is expected as a result")
	}
}

type MockHandle struct {
}

func (h *MockHandle) Handle(*rabbitLib.Delivery, *sync.WaitGroup) {
	tempCounter++
}

var tempCounter int

func TestClient_NewConsumer(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})

	consumer := client.NewConsumer(&MockHandle{})
	if reflect.TypeOf(consumer).String() != "*amqp.Consumer" {
		t.Fatalf("the return result is of type '%s' expected '*amqp.Consumer'", reflect.TypeOf(consumer).String())
	}
}

func TestConsumer_IsInit(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})
	consumer := client.NewConsumer(&MockHandle{})

	consumer.SetIsInit(true)
	if !consumer.isInit {
		t.Fatal("set init is not affecting consumer")
	}

	consumer.SetIsInit(false)
	if consumer.isInit {
		t.Fatal("set init is not affecting consumer")
	}
}
func TestClient_SilenceMode(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})

	client.SetSilenceMode(true)
	if !client.silenceMode {
		t.Fatal("set init is not affecting client")
	}

	client.SetSilenceMode(false)
	if client.silenceMode {
		t.Fatal("set init is not affecting client")
	}
}

func TestClient_SetMaintain(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})
	consumer := client.NewConsumer(&MockHandle{})

	consumer.SetMaintain(true)
	if !consumer.IsMaintain {
		t.Fatal("set maintain is not affecting consumer")
	}

	consumer.SetMaintain(false)
	if consumer.IsMaintain {
		t.Fatal("set maintain is not affecting consumer")
	}
}

func TestClient_SetDelay(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})
	consumer := client.NewConsumer(&MockHandle{})

	consumer.SetDelay(time.Second)
	if consumer.delay != time.Second {
		t.Fatal("set delay is not affecting consumer")
	}
}

func TestClient_GetConsumers(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})

	consumers := client.GetConsumers()
	if len(consumers) > 0 {
		t.Fatalf("consumer cards have not been added, the list is not empty")
	}

	client.NewConsumer(&MockHandle{})

	consumers = client.GetConsumers()
	if len(consumers) != 1 {
		t.Fatalf("one consumer has been added, but the length of the list is not equal to 1")
	}
}

func TestClient_ClearConsumers(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})
	client.NewConsumer(&MockHandle{})
	client.ClearConsumers()

	consumers := client.GetConsumers()
	if len(consumers) > 0 {
		t.Fatalf("list after cleaning should be empty")
	}
}

func TestClient_OpenChannel(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})

	client.OpenChannel()
}

func TestClient_DeclareQueue(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})

	client.DeclareQueue()
}

func TestClient_DeclareExchange(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})

	client.DeclareExchange()
}

func TestClient_Close(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})

	client.Close()
}

func TestClient_Publish(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})

	client.Publish("test", "")
}

func TestClient_PublishExchange(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})

	client.PublishExchange("test", "exchange")
}

func TestConsumer_Init(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})

	consumer := client.NewConsumer(&MockHandle{})
	if reflect.TypeOf(consumer).String() != "*amqp.Consumer" {
		t.Fatalf("the return result is of type '%s' expected '*amqp.Consumer'", reflect.TypeOf(consumer).String())
	}

	consumer.Init()
}

func TestClient_QueueBind(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})
	client.QueueBind("test", "test", "test")
}

func TestClient_ReConsume(t *testing.T) {
	client := NewClient(Config{}, zerolog.Logger{})
	client.NewConsumer(&MockHandle{})
	client.reConsume()
}
