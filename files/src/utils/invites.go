package utils

import (
	"crypto/sha1"
	"encoding/base64"

	"github.com/fpgschiba/volleygoals/models"
)

func GenerateInviteToken(teamId, email string, role models.TeamMemberRole) string {
	bv := []byte(teamId + "|" + email + "|" + string(role))
	hasher := sha1.New()
	hasher.Write(bv)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}
