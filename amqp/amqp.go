package amqp

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	rabbitLib "github.com/streadway/amqp"
	"sync"
	"time"
)

var (
	errAvailable    = errors.New("rabbitMQ server is not available")
	errChannelIsNil = errors.New("errors channel is nil")
)

const (
	// DefaultConsumeIdleTimeout время ожидания сообщения из очереди после которого консьюмер прекращает работу при установленном свойстве Consumer.IsMaintain
	DefaultConsumeIdleTimeout = 10 * time.Second
	// DefaultDelayIdleTimeout задержка перед получением следующего сообщения из очереди
	DefaultDelayIdleTimeout = 1 * time.Second

	MessageHeaderCountAttempt = "x-count-attempt"
	DefaultVirtualHost        = "/"
)

// Publisher только для публикации сообщений в RabbitMQ
type Publisher interface {
	Publish(body string) error
	GetQueueName() string
}

// Client реализует клиент к RabbitMQ
type Client struct {
	sync.RWMutex
	connection      *rabbitLib.Connection // Указатель на соединение
	channel         *rabbitLib.Channel    // Указатель на канал соединения
	errorChannel    chan *rabbitLib.Error // Go канал для оповещения о разрыве соединения
	logger          zerolog.Logger        // Указатель на логер
	config          Config                // Конфиг присвоенный при инициализации
	consumers       []*Consumer           // Слайс консьюмеров слушающих очередь
	silenceMode     bool                  // Режим тишины  - при установке в true при публикации логи не пишутся
	declareEntities bool                  // Декларировать ли Queue и Exchange
}

// Consumer реализует слушатель очереди RabbitMQ
type Consumer struct {
	sync.RWMutex
	wg         *sync.WaitGroup // Для проверки когда все горутины в консьюмере закончили выполнение
	client     *Client         // Указатель на клиент к которому относится консьюмер
	tag        string          // Можно указать TAG, если осталяем пустым, rabbit сам сгенерирует
	done       chan error      // Go канал для принудительного завершения консьюмера
	handler    handler         // Функция которая выполняется для каждого сообщения в очереди
	deadline   time.Time       // Используется при установленном свойстве IsMaintain=false
	Timeout    time.Duration   // Время которое ждет консьюмер сообщений в очереди, если IsMaintain=false и очередь пуста, то консьюмер завершает работу
	IsMaintain bool            // Если true - консьюмер бесконечно ждет сообщений в очереди
	isInit     bool            // Для проверки инициализирован консьюмер или нет
	delay      time.Duration   // Задержка перед получением сообщения из очереди
}

// Config содержит конфигурация клиента
type Config struct {
	Host          string
	Port          int
	Queue         string
	Exchange      string // В данной реализации не используется, но при желании можно
	UserName      string
	Password      string
	PrefetchCount int
	Arguments     rabbitLib.Table // Указатель на дополнительные параметры очереди
	ContentType   string
	VirtualHost   string
	Properties    map[string]interface{}
}

// NewClient создает экземпляр структуры с требуемыми параметрами
func NewClient(config Config, logger zerolog.Logger) *Client {
	c := &Client{
		config: config,
		logger: logger,
	}

	return c
}

func (client *Client) DeclareEntities(b bool) *Client {
	client.declareEntities = b
	return client
}

// GetQueueName возвращает имя очереди к которой привязан клиент
func (client *Client) GetQueueName() string {
	return client.config.Queue
}

// Connect запускает процесс соединения и поддержания соединения
func (client *Client) Connect() *Client {
	err := client.connect()
	if err != nil {
		client.logger.Fatal().Dict("error connect", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Err(err)).Msg("")
	}
	go client.reConnector()
	return client
}

// reConnector получает сообщения о разрыве соединения и запуск процесса подключения
func (client *Client) reConnector() {
	select {
	case err, ok := <-client.errorChannel:
		if ok && err != nil {
			client.logger.Error().Dict("reconnecting after connection closed", zerolog.Dict().Err(err)).Msg("")
			client.connection.Close()
			client.connectLoop()
			return
		}
	}
}

// connectLoop пытается соединится, если успешно также патается восстановить консьюмеры
func (client *Client) connectLoop() {
	for {
		err := client.connect()
		if err != nil {
			client.logger.Error().Dict("connection to rabbitMQ failed. Retrying in 1 sec... ", zerolog.Dict().Err(err)).Msg("")
			time.Sleep(1 * time.Second)
		} else {
			client.logger.Info().Dict("connection to rabbitMQ successful", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port))).Msg("")
			go client.reConnector()
			client.reConsume()
			return
		}
	}
}

// connect инициализирует клиент
func (client *Client) connect() error {
	err := client.Dial()
	if err != nil {
		return err
	}

	client.errorChannel = make(chan *rabbitLib.Error)
	client.connection.NotifyClose(client.errorChannel) // Указываем канал в который будут сыпаться ошибки при разрывах
	client.logger.Info().Dict("connection to rabbitMQ successful", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port))).Msg("")
	client.OpenChannel()
	if client.declareEntities {
		client.DeclareQueue()
		client.DeclareExchange()
	}

	return nil
}

// Dial инициализирует подключение
func (client *Client) Dial() error {
	client.logger.Info().Dict("connecting to rabbitMQ", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port))).Str("queue", client.config.Queue).Msg("")

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", client.config.UserName, client.config.Password, client.config.Host, client.config.Port, client.config.VirtualHost)
	config := rabbitLib.Config{Properties: client.config.Properties}
	conn, err := rabbitLib.DialConfig(url, config)

	if err != nil {
		client.logger.Error().Dict("connection to rabbit", zerolog.Dict().Str("addr", fmt.Sprintf("amqp://%s:%s@%s:%d/%s", client.config.UserName, client.config.Password, client.config.Host, client.config.Port, client.config.VirtualHost))).Err(err).Msg("")
		return err
	}

	client.connection = conn
	return nil
}

// ClearConsumers очищает список консьюмеров и завершает их
func (client *Client) ClearConsumers() *Client {
	for _, consumer := range client.GetConsumers() {
		if consumer.IsNotShutdown() {
			consumer.Close()
		}
	}

	for {
		flag := true
		for _, consumer := range client.GetConsumers() {
			if consumer.IsNotShutdown() {
				flag = false
			}
		}

		if flag {
			break
		}
	}

	client.Lock()
	client.consumers = make([]*Consumer, 0)
	client.Unlock()
	return client
}

// GetConsumers получает зарегестрированные консьюмеры
func (client *Client) GetConsumers() []*Consumer {
	client.RLock()
	defer client.RUnlock()
	return client.consumers
}

// OpenChannel открывает канал
func (client *Client) OpenChannel() *Client {
	if client.connection != nil {
		channel, err := client.connection.Channel()
		if client.config.PrefetchCount > 0 && channel != nil {
			err := channel.Qos(client.config.PrefetchCount, 0, false)
			if err != nil {
				client.logger.Error().Dict("set prefetchCount", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Err(err)).Msg("")
			}
		}
		if err != nil {
			client.logger.Error().Dict("opening channel", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Err(err)).Msg("")
		} else {
			client.logger.Info().Dict("opening channel", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)))
		}
		client.channel = channel
	}
	return client
}

// DeclareQueue объявляет очередь
func (client *Client) DeclareQueue() *Client {
	if client.channel == nil {
		client.logger.Error().Dict("the queue cannot be declared because the channel is nil", zerolog.Dict().Err(errChannelIsNil)).Msg("")
		return client
	}
	_, err := client.channel.QueueDeclare(
		client.config.Queue,     // name
		true,                    // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		client.config.Arguments, // arguments
	)
	if err != nil {
		client.logger.Error().Dict("queue declaration", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Str("queueName", client.config.Queue).Err(err)).Msg("")
	} else {
		client.logger.Info().Dict("queue declaration", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Str("queueName", client.config.Queue)).Msg("")
	}

	return client
}

// DeclareExchange объявляет exchange
func (client *Client) DeclareExchange() *Client {
	if client.config.Exchange == "" {
		return client
	}
	if client.channel == nil {
		client.logger.Error().Dict("the queue cannot be declared because the channel is nil", zerolog.Dict().Err(errChannelIsNil)).Msg("")
		return client
	}
	err := client.channel.ExchangeDeclare(
		client.config.Exchange, // name
		rabbitLib.ExchangeDirect,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		client.logger.Error().Dict("exchange declaration", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Str("exchangeName", client.config.Exchange).Err(err)).Msg("")
	} else {
		client.logger.Info().Dict("exchange declaration", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Str("exchangeName", client.config.Exchange)).Msg("")
	}

	return client
}

// QueueBind привязывает очередь к exchange
func (client *Client) QueueBind(queue, routingKey, exchangeName string) *Client {
	if client.channel == nil {
		client.logger.Error().Dict("It is impossible to bind because no channel is advertised", zerolog.Dict().Err(errChannelIsNil)).Msg("")
		return client
	}
	err := client.channel.QueueBind(
		queue,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		client.logger.Error().Dict("queue binding", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Str("exchangeName", exchangeName).Str("queueName", queue).Str("routingKey", routingKey).Err(err)).Msg("")
	} else {
		client.logger.Info().Dict("queue binding", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Str("exchangeName", exchangeName).Str("queueName", queue).Str("routingKey", routingKey)).Msg("")
	}

	return client
}

// Close закрывает клиент
func (client *Client) Close() {
	client.logger.Info().Dict("closing connection", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port))).Msg("")
	client.ClearConsumers()

	for {
		if len(client.consumers) == 0 {
			if client.channel != nil {
				client.channel.Close()
			}
			if client.connection != nil {
				client.connection.Close()
			}
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// GetChannel возвращает указатель на канал
func (client *Client) GetChannel() *rabbitLib.Channel {
	return client.channel
}

// SetSilenceMode устанавливает режим тишины (Публикация сообщений не логируется)
func (client *Client) SetSilenceMode(mode bool) *Client {
	client.silenceMode = mode
	return client
}

// Publish публикует сообщение в очередь
func (client *Client) Publish(body string, routingKey string) error {
	if client == nil {
		return errAvailable
	}

	if client.channel == nil {
		return errChannelIsNil
	}

	contentType := "application/json"
	if client.config.ContentType != "" {
		contentType = client.config.ContentType
	}

	if routingKey == "" {
		routingKey = client.config.Queue
	}

	err := client.channel.Publish(
		client.config.Exchange, // exchange
		routingKey,             // routing key
		false,                  // mandatory
		false,                  // immediate
		rabbitLib.Publishing{
			ContentType: contentType,
			Body:        []byte(body),
		})
	if err != nil {
		client.logger.Error().Dict("error publish", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Time("time", time.Now()).Err(err)).Msg("")
		return err
	}

	if !client.silenceMode {
		client.logger.Info().Dict("publish in queue", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Time("time", time.Now()).Str("queueName", client.config.Queue).Str("event_message", body)).Msg("")
	}
	return nil
}

// Publish публикует сообщение в очередь
func (client *Client) PublishWithCount(body string, routingKey string, countAttempt int) error {
	if client == nil {
		return errAvailable
	}

	if client.channel == nil {
		return errChannelIsNil
	}

	contentType := "application/json"
	if client.config.ContentType != "" {
		contentType = client.config.ContentType
	}

	if routingKey == "" {
		routingKey = client.config.Queue
	}

	err := client.channel.Publish(
		client.config.Exchange, // exchange
		routingKey,             // routing key
		false,                  // mandatory
		false,                  // immediate
		rabbitLib.Publishing{
			ContentType: contentType,
			Body:        []byte(body),
			Headers: rabbitLib.Table{
				MessageHeaderCountAttempt: countAttempt,
			},
		})
	if err != nil {
		client.logger.Error().Dict("error publish", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Time("time", time.Now()).Err(err)).Msg("")
		return err
	}

	if !client.silenceMode {
		client.logger.Info().Dict("publish in queue", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Time("time", time.Now()).Str("queueName", client.config.Queue).Str("event_message", body)).Msg("")
	}

	return nil
}

func GetMessageCountAttempt(d *rabbitLib.Delivery) int {
	if v, ok := d.Headers[MessageHeaderCountAttempt]; ok {
		return int(v.(int32))
	}

	return 0
}

// PublishExchange публикует сообщение в exchange
func (client *Client) PublishExchange(body string, exchange string) error {
	if client == nil {
		return errAvailable
	}

	if client.channel == nil {
		return errChannelIsNil
	}

	err := client.channel.Publish(
		exchange, // exchange
		"",       // routing key
		false,    // mandatory
		false,    // immediate
		rabbitLib.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		client.logger.Error().Dict("error publish", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Str("exchangeName", exchange).Err(err)).Msg("")
		return err
	}

	if !client.silenceMode {
		client.logger.Info().Dict("publish in exchange", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)).Str("exchangeName", exchange).Str("event_message", body)).Msg("")
	}
	return nil
}

func (consumer *Consumer) SetIsInit(value bool) {
	consumer.Lock()
	defer consumer.Unlock()
	consumer.isInit = value
}

func (consumer *Consumer) IsNotShutdown() bool {
	consumer.RLock()
	defer consumer.RUnlock()
	return consumer.isInit
}

// Init ининциализирует консьюмер
func (consumer *Consumer) Init() error {
	if consumer.client.channel == nil {
		return errChannelIsNil
	}

	deliveries, err := consumer.client.channel.Consume(
		consumer.client.config.Queue, // name
		consumer.tag,                 // consumerTag,
		false,                        // noAck
		false,                        // exclusive
		false,                        // noLocal
		false,                        // noWait
		nil,                          // arguments
	)

	if err != nil {
		consumer.client.logger.Error().Dict("queue consume", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", consumer.client.config.Host, consumer.client.config.Port)).Str("queueName", consumer.client.config.Queue).Err(err)).Msg("")
	}

	consumer.SetIsInit(true)
	consumer.client.logger.Info().Dict("queue bound to exchange, starting consume", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", consumer.client.config.Host, consumer.client.config.Port)).Str("queueName", consumer.client.config.Queue).Str("consumerTag", consumer.tag)).Msg("")
	// С помощью wg отслеживаем, когда горутина обрабатывающая сообщения из очереди завершит работу
	consumer.wg.Add(1)
	go consumer.handle(deliveries, consumer.done, consumer.wg)
	// До тех пор ждем и приложение не завершает работу
	consumer.wg.Wait()
	consumer.SetIsInit(false)
	return nil
}

// reConsume переинициализирует консьюмеры клиента
func (client *Client) reConsume() {
	for _, consumer := range client.GetConsumers() {
		err := consumer.Init()
		if err != nil {
			client.logger.Error().Dict("init consumer", zerolog.Dict().Str("queueName", consumer.client.config.Queue).Err(err)).Msg("")
		}
	}
}

// NewConsumer возвращает экземпляр структуры Consumer и добавляет ее в список консьюмеров клиента
func (client *Client) NewConsumer(handle handler, tag string) *Consumer {
	consumer := &Consumer{
		handler:    handle,
		client:     client,
		tag:        tag, // Когда tag пуст. Rabbit атоматом сгенирирует tag/
		done:       make(chan error),
		deadline:   time.Now().Add(DefaultConsumeIdleTimeout),
		Timeout:    DefaultConsumeIdleTimeout,
		IsMaintain: true,
		wg:         new(sync.WaitGroup),
		delay:      0,
	}

	client.Lock()
	client.consumers = append(client.consumers, consumer)
	client.Unlock()
	return consumer
}

// SetTimeout устанавливает время жизни консьюмера без сообщений в очереди
func (consumer *Consumer) SetTimeout(timeout time.Duration) *Consumer {
	consumer.Timeout = timeout
	consumer.deadline = time.Now().Add(timeout)
	return consumer
}

// SetDelay устанавливает задержку между получениями собщений из очереди
func (consumer *Consumer) SetDelay(delay time.Duration) *Consumer {
	consumer.delay = delay
	return consumer
}

// SetMaintain  устанавливает режим, при котором консьюмер бесконечно слушает очередь
func (consumer *Consumer) SetMaintain(isMaintain bool) *Consumer {
	consumer.IsMaintain = isMaintain
	return consumer
}

// Close закрывает консьюмер, ожидая завершения все горутин которые он вызвал
func (consumer *Consumer) Close() {
	consumer.done <- nil
	for {
		if !consumer.IsNotShutdown() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// handle реализут принятие сообщения из очереди и передачу колбэку
func (consumer *Consumer) handle(deliveries <-chan rabbitLib.Delivery, done <-chan error, wgMain *sync.WaitGroup) {
	defer consumer.client.logger.Info().Dict("handle: deliveries channel closed", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", consumer.client.config.Host, consumer.client.config.Port)).Str("queueName", consumer.client.config.Queue)).Msg("")
	defer wgMain.Done()

	tickChan := time.NewTicker(DefaultDelayIdleTimeout).C

	var wg sync.WaitGroup

	for {
		select {
		case d, ok := <-deliveries: // Получаем сообщение из очереди
			if !ok {
				wg.Wait()
				return
			}
			if d.Body != nil {
				if !consumer.client.silenceMode {
					consumer.client.logger.Info().Dict("message received", zerolog.Dict().Str("addr", fmt.Sprintf("%s:%d", consumer.client.config.Host, consumer.client.config.Port)).Str("queueName", consumer.client.config.Queue).Str("event_message", string(d.Body))).Msg("")
				}
				wg.Add(1) // Указываем что запущена горутина в которой обрабатывается сообщение. Как только все горутины завершат работу, можно будет завершить метод.
				time.Sleep(consumer.delay)
				go consumer.handler.Handle(&d, &wg)
				consumer.deadline = time.Now().Add(consumer.Timeout)
			}
		case <-tickChan: // Проверяем не вышло ли время жизни консьюмера, если да то ждем завершения все горутин и выходим
			if !consumer.IsMaintain {
				if time.Now().After(consumer.deadline) {
					wg.Wait()
					return
				}
			}
		case <-done: // Принудительный выход с ожиданием при сигнале снаружи
			wg.Wait()
			return
		}
	}
}

// handler типовой обработчик в консьюмере
type handler interface {
	Handle(*rabbitLib.Delivery, *sync.WaitGroup)
}
