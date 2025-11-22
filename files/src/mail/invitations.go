package mail

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

func SendInvitationEmail(ctx context.Context, toEmail, inviteToken, baseUrl string) error {
	client := GetClient()
	completeInviteLink := baseUrl + "/accept-invite?token=" + inviteToken
	result, err := client.SendEmail(ctx, &sesv2.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    aws.String("You're invited to join VolleyGoals!"),
					Charset: aws.String("UTF-8"),
				},
				Body: &types.Body{
					Text: &types.Content{
						Data:    aws.String("Hello,\n\nYou have been invited to join VolleyGoals. Please click the link below to accept the invitation:\n\n" + completeInviteLink + "\n\nBest regards,\nThe VolleyGoals Team"),
						Charset: aws.String("UTF-8"),
					},
				},
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
