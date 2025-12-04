package mail

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/fpgschiba/volleygoals/models"
)

func SendInvitationEmail(ctx context.Context, toEmail, inviteToken, teamName, inviterName, message string, expiry int) error {
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
					"expiryDays": "%d",
					"personalMessage": "%s"
				}`, inviterName, teamName, completeInviteLink, expiry, message)),
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

func ResendInvitationEmail(ctx context.Context, invite *models.Invite, inviter *models.User, team *models.Team) error {
	var message string
	if invite.Message != nil {
		message = *invite.Message
	} else {
		message = "Welcome to the team!"
	}
	expiresInDays := int(invite.ExpiresAt.Sub(time.Now()).Hours() / 24)
	var inviterName string
	if inviter != nil && inviter.Name != nil {
		inviterName = *inviter.Name
	} else if inviter != nil {
		inviterName = inviter.Email
	} else {
		inviterName = "A team admin"
	}
	return SendInvitationEmail(ctx, invite.Email, invite.Token, team.Name, inviterName, message, expiresInDays)
}
