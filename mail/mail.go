package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"mailer-ms/config"
	"mailer-ms/queue"
	"mailer-ms/tracer"
	"mailer-ms/utils/array"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/rabbitmq/amqp091-go"
)

var utf8 = "utf-8"

type Mailer struct {
	client   SesApi
	cfg      *config.Config
	queue    *queue.Server
	validate *validator.Validate
}

func New(cfg *config.Config, queue *queue.Server) Mailer {
	return Mailer{
		cfg:      cfg,
		queue:    queue,
		validate: validator.New(),
		client:   ses.NewFromConfig(cfg.Aws.Instance),
	}
}

func (m *Mailer) handleSendEmailFailure(ctx context.Context, originalDelivery *amqp091.Delivery, failure error) {
	// TODO: finish me !
	m.queue.Publish(ctx, "", "", amqp091.Publishing{})
}

func (m *Mailer) HandleMailRequestDelivery(d *amqp091.Delivery) {
	ctx, span := tracer.NewSpan(context.TODO(), "queue", "SendEmail")
	defer span.End()

	var dto SendEmailDto

	if err := json.Unmarshal(d.Body, &dto); err != nil {
		tracer.AddSpanErrorAndFail(span, err, "failed to unmarshal send mail request")
		m.handleSendEmailFailure(ctx, d, err)
		return
	}

	if err := m.validate.Struct(dto); err != nil {
		m.handleSendEmailFailure(ctx, d, fmt.Errorf("validation error: %w", err))
		return
	}

	m.SendEmail(ctx, &dto)
}

func (m *Mailer) SendEmail(ctx context.Context, dto *SendEmailDto) error {
	ctx, span := tracer.NewSpan(ctx, "mailer", "SendEmail")
	defer span.End()

	recipients := append(dto.To, dto.Cc...)
	recipients = append(recipients, dto.Bcc...)

	for _, recipient := range array.RemoveDuplicates(recipients) {
		go m.Send(ctx, &ses.SendEmailInput{
			Source:           &m.cfg.Mail.Sender,
			ReplyToAddresses: dto.ReplyToAddresses,
			Destination: &types.Destination{
				ToAddresses:  []string{recipient},
				CcAddresses:  []string{},
				BccAddresses: []string{},
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
		})
	}

	return nil
}

func (m *Mailer) Send(ctx context.Context, emailInput *ses.SendEmailInput) {
	m.SendWithRetry(ctx, 1, emailInput)
}

func (m *Mailer) SendWithRetry(ctx context.Context, currentAttempt int, emailInput *ses.SendEmailInput) {
	ctx, span := tracer.NewSpan(ctx, "mailer", "SendWithRetry")
	defer span.End()

	span.SetAttributes(attribute.Key("recipient").String(emailInput.Destination.ToAddresses[0]))
	span.SetAttributes(attribute.Key("subject").String(*emailInput.Message.Subject.Data))

	_, sesError := m.client.SendEmail(ctx, emailInput)

	if sesError == nil {
		span.SetStatus(codes.Ok, "email sent successfully")
		return
	}

	if currentAttempt > m.cfg.Mail.MaxRetryAttempts {
		tracer.AddSpanErrorAndFail(span, sesError, fmt.Sprintf("MAX_SES_RETRY_ATTEMPTS of: %d reached", m.cfg.Mail.MaxRetryAttempts))
		return
	}

	span.RecordError(sesError)

	time.Sleep(time.Duration(m.cfg.Mail.RetryWaitTime) * time.Second)

	m.SendWithRetry(ctx, currentAttempt+1, emailInput)
}
