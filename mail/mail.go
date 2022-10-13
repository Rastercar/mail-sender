package mail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mailer-ms/config"
	"mailer-ms/queue"
	"mailer-ms/tracer"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/google/uuid"
	"golang.org/x/time/rate"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/rabbitmq/amqp091-go"
)

var (
	utf8             = "utf-8"
	mailUuidTag      = "mail_uuid"
	maxSesRecipients = 50
)

type Mailer struct {
	client      SesApi
	cfg         *config.Config
	queue       *queue.Server
	validate    *validator.Validate
	rateLimiter *rate.Limiter
}

func New(cfg *config.Config, queue *queue.Server) Mailer {
	requestsPerMs := 1000 / cfg.Mail.ReqPerSecLimit
	limit := rate.Every(time.Duration(requestsPerMs) * time.Millisecond)

	return Mailer{
		cfg:         cfg,
		queue:       queue,
		validate:    validator.New(),
		client:      ses.NewFromConfig(cfg.Aws.Instance),
		rateLimiter: rate.NewLimiter(limit, 1),
	}
}

func (m *Mailer) handleMailRequestResult(ctx context.Context, originalDelivery *amqp091.Delivery, failure error) {
	ctx, span := tracer.NewSpan(ctx, "mail", "handleMailRequestResult")
	defer span.End()

	successMsg := "email queued successfully"

	var resType string
	var body []byte

	if failure == nil {
		resType = "success"
		body, _ = json.Marshal(SendEmailRes{Success: true, Message: successMsg})

		span.SetStatus(codes.Ok, successMsg)

		originalDelivery.Ack(false)
	} else {
		resType = "error"
		body, _ = json.Marshal(SendEmailRes{Success: false, Message: failure.Error()})

		span.RecordError(failure)
		span.SetStatus(codes.Error, "failed to queue email")

		originalDelivery.Reject(false)
	}

	if originalDelivery.ReplyTo != "" && originalDelivery.CorrelationId != "" {
		err := m.queue.Publish(ctx, "", originalDelivery.ReplyTo, amqp091.Publishing{
			Body:          body,
			Type:          resType,
			CorrelationId: originalDelivery.CorrelationId,
		})

		if err != nil {
			tracer.AddSpanErrorAndFail(span, err, "failed to publish rpc response")
		}
	}
}

func (m *Mailer) HandleMailRequestDelivery(d *amqp091.Delivery) {
	ctx, span := tracer.NewSpan(context.TODO(), "mail", "SendEmail")
	defer span.End()

	var dto SendEmailDto

	if err := json.Unmarshal(d.Body, &dto); err != nil {
		tracer.AddSpanErrorAndFail(span, err, "failed to unmarshal send mail request")
		m.handleMailRequestResult(ctx, d, err)
		return
	}

	if _, err := uuid.Parse(dto.Uuid); err != nil {
		tracer.AddSpanErrorAndFail(span, err, "invalid email uuid")
		m.handleMailRequestResult(ctx, d, errors.New("invalid email uuid"))
		return
	}

	recipientCnt := len(dto.To) + len(dto.Cc) + len(dto.Bcc)

	if recipientCnt > maxSesRecipients {
		span.SetStatus(codes.Error, "email recipient count is over 50")
		m.handleMailRequestResult(ctx, d, errors.New("email recipient count is over 50"))
		return
	}

	if err := m.validate.Struct(dto); err != nil {
		tracer.AddSpanErrorAndFail(span, err, "invalid email request")
		m.handleMailRequestResult(ctx, d, fmt.Errorf("validation error: %w", err))
		return
	}

	input := ses.SendEmailInput{
		Source:           &m.cfg.Mail.Sender,
		ReplyToAddresses: dto.ReplyToAddresses,
		Destination: &types.Destination{
			ToAddresses:  dto.To,
			CcAddresses:  dto.Cc,
			BccAddresses: dto.Bcc,
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    &dto.SubjectText,
				Charset: &utf8,
			},
			Body: &types.Body{
				Html: &types.Content{
					Data:    &dto.BodyHtml,
					Charset: &utf8,
				},
				Text: &types.Content{
					Data:    &dto.BodyText,
					Charset: &utf8,
				},
			},
		},
		Tags: []types.MessageTag{
			{
				Name:  &mailUuidTag,
				Value: &dto.Uuid,
			},
		},
	}

	if err := m.sendWithRetry(ctx, 1, &input); err != nil {
		tracer.AddSpanErrorAndFail(span, err, "invalid email request")
		m.handleMailRequestResult(ctx, d, fmt.Errorf("validation error: %w", err))
		return
	}

	m.handleMailRequestResult(ctx, d, nil)
}

func (m *Mailer) sendWithRetry(ctx context.Context, currentAttempt int, emailInput *ses.SendEmailInput) error {
	ctx, span := tracer.NewSpan(ctx, "mail", "SendWithRetry")
	defer span.End()

	span.SetAttributes(attribute.Key("recipient").String(emailInput.Destination.ToAddresses[0]))
	span.SetAttributes(attribute.Key("subject").String(*emailInput.Message.Subject.Data))
	span.SetAttributes(attribute.Key("attempt").Int(currentAttempt))

	m.rateLimiter.Wait(ctx)

	_, sesError := m.client.SendEmail(ctx, emailInput)

	if sesError == nil {
		span.SetStatus(codes.Ok, "email sent successfully")
		return nil
	}

	if currentAttempt > m.cfg.Mail.MaxRetryAttempts {
		tracer.AddSpanErrorAndFail(span, sesError, fmt.Sprintf("MAX_SES_RETRY_ATTEMPTS of: %d reached", m.cfg.Mail.MaxRetryAttempts))
		return sesError
	}

	span.RecordError(sesError)

	time.Sleep(time.Duration(m.cfg.Mail.RetryWaitTime) * time.Second)

	return m.sendWithRetry(ctx, currentAttempt+1, emailInput)
}
