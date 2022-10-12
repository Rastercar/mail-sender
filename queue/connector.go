package queue

import (
	"log"
	"mailer-ms/queue/interfaces"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpConnectionWrapper struct {
	conn *amqp.Connection
}

func (w AmqpConnectionWrapper) Close() error {
	return w.conn.Close()
}

func (w AmqpConnectionWrapper) Channel() (interfaces.AmqpChannel, error) {
	return w.conn.Channel()
}

type Connector struct{}

func (c *Connector) Connect(url string) (interfaces.AmqpConnection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return AmqpConnectionWrapper{conn}, nil
}

func (s *Server) connect() {
	currentAttempt := 1
	sleepTime := time.Second * time.Duration(s.cfg.ReconnectWaitTime)

	for {
		log.Printf("[ RMQ ] trying to connect, attempt: %d", currentAttempt)

		con, err := s.Connector.Connect(s.cfg.Url)

		if err != nil || con == nil {
			log.Printf("[ RMQ ] connection failed %v", err)

			currentAttempt++
			time.Sleep(sleepTime)

			continue
		}

		channel, err := con.Channel()
		if err != nil {
			log.Printf("[ RMQ ] connection channel failed %v", err)

			currentAttempt++
			time.Sleep(sleepTime)

			continue
		}

		// TODO: DECLARE AND BIND QUEUES HERE
		err = channel.ExchangeDeclare(
			s.cfg.Exchange, // name
			"topic",        // kind
			true,           // durable
			false,          // auto-deleted
			false,          // internal
			false,          // no-wait
			nil,            // args
		)

		// Its intentional to panic here as we NEED the queues and exchanges
		// to be successfully declared for this service to function correctly
		if err != nil {
			log.Fatalf("[ RMQ ] failed to declare exchange: %v ", err)
		}

		s.conn = con
		s.channel = channel

		s.notifyClose = make(chan *amqp.Error, 1024)
		s.channel.NotifyClose(s.notifyClose)

		s.startConsumer()

		log.Printf("[ RMQ ] connected")
		return
	}
}
