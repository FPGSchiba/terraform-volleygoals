package mail

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

func SendInvitationEmail(ctx context.Context, toEmail, inviteToken, teamName, inviterName string, expiry int) error {
	client = GetClient()
	completeInviteLink := FrontendBaseUrl + "/accept-invite?token=" + inviteToken
	result, err := client.SendEmail(ctx, &sesv2.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Content: &types.EmailContent{
			Template: &types.Template{
				TemplateArn: aws.String(InviteTemplateArn),
				TemplateData: aws.String(fmt.Sprintf(`{
					"inviterName": "%s",
					"teamName": "%s",
					"acceptLink": "%s",
					"expiryDays": "%d"
				}`, inviterName, teamName, completeInviteLink, expiry)),
			},
		},
		FromEmailAddress: aws.String(EmailSender),
	})
	if err != nil {
		return err
	}
	_ = result
	return nil
}
