package queue

import (
	"context"
	"log"
	"mailer-ms/config"
	"mailer-ms/queue/interfaces"
	"mailer-ms/tracer"

	amqp "github.com/rabbitmq/amqp091-go"
)

//go:generate mockgen -destination=../mocks/amqp.go -package=mocks github.com/rabbitmq/amqp091-go Acknowledger

type Server struct {
	interfaces.Connector
	interfaces.Publisher

	cfg         config.RmqConfig
	conn        interfaces.AmqpConnection
	channel     interfaces.AmqpChannel
	deliveries  <-chan amqp.Delivery
	notifyClose chan *amqp.Error

	ConsumerFn func(deliver *amqp.Delivery)
}

func New(cfg config.RmqConfig) Server {
	return Server{
		cfg:       cfg,
		Connector: &Connector{},
		Publisher: &Publisher{},
	}
}

func (s *Server) Start() {
	go func() {
		for {
			s.connect()

			connectionError, chanClosed := <-s.notifyClose

			// connection error is nil and chanClosed is false when
			// the connection was closed manually with client code
			if connectionError != nil {
				log.Printf("[ RMQ ] connection error: %v \n", connectionError)
			}

			if !chanClosed {
				return
			}
		}
	}()
}

func (s *Server) Stop() error {
	log.Printf("[ RMQ ] closing connections")

	if s.conn != nil {
		return s.conn.Close()
	}

	return nil
}

func (s *Server) Publish(ctx context.Context, exchange, key string, publishing amqp.Publishing) error {
	ctx, span := tracer.NewSpan(ctx, "queue", "Publish")
	defer span.End()

	return s.Publisher.PublishWithContext(ctx, s.channel, exchange, key, publishing)
}
