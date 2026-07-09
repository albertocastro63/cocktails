package email

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

// SESSender delivers reset emails through Amazon SES v2, from a verified
// site-domain address.
type SESSender struct {
	client *sesv2.Client
	from   string
}

func NewSESSender(client *sesv2.Client, from string) *SESSender {
	return &SESSender{client: client, from: from}
}

func (s *SESSender) SendPasswordReset(to string, data PasswordResetData) error {
	msg := BuildResetEmail(data)
	_, err := s.client.SendEmail(context.Background(), &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.from),
		Destination:      &types.Destination{ToAddresses: []string{to}},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws.String(msg.Subject)},
				Body: &types.Body{
					Html: &types.Content{Data: aws.String(msg.HTML)},
					Text: &types.Content{Data: aws.String(msg.Text)},
				},
			},
		},
	})
	return err
}
