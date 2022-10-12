package mail

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ses"
)

type SesApi interface {
	SendEmail(ctx context.Context, params *ses.SendEmailInput, optFns ...func(*ses.Options)) (*ses.SendEmailOutput, error)
}
