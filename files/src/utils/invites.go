package utils

import (
	"crypto/sha1"
	"encoding/base64"
	"math/rand"
	"time"

	"github.com/fpgschiba/volleygoals/models"
)

func GenerateInviteToken(teamId, email string, role models.TeamMemberRole) string {
	bv := []byte(teamId + "|" + email + "|" + string(role))
	hasher := sha1.New()
	hasher.Write(bv)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

func ValidateInviteToken(token string, email, teamId string, role models.TeamMemberRole) bool {
	expectedToken := GenerateInviteToken(teamId, email, role)
	return expectedToken == token
}

func GeneratePassword(passwordLength int) string {
	// Character sets for generating passwords
	lowerCase := "abcdefghijklmnopqrstuvwxyz" // lowercase
	upperCase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" // uppercase
	numbers := "0123456789"                   // numbers
	specialChar := "!@#$%&*_-+=?"             // special characters

	// Variable for storing password
	password := ""

	// Initialize the random number generator
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Generate password character by character
	for n := 0; n < passwordLength; n++ {
		// Generate a random number to choose a character set
		randNum := rng.Intn(4)

		switch randNum {
		case 0:
			randCharNum := rng.Intn(len(lowerCase))
			password += string(lowerCase[randCharNum])
		case 1:
			randCharNum := rng.Intn(len(upperCase))
			password += string(upperCase[randCharNum])
		case 2:
			randCharNum := rng.Intn(len(numbers))
			password += string(numbers[randCharNum])
		case 3:
			randCharNum := rng.Intn(len(specialChar))
			password += string(specialChar[randCharNum])
		}
	}

	return password
}
