package self

type UpdateSelfInput struct {
	Name              *string `json:"name,omitempty"`
	PreferredUsername *string `json:"preferredUsername,omitempty"`
	Birthdate         *string `json:"birthdate,omitempty"` // in format "YYYY-MM-DD"
}

func (u *UpdateSelfInput) ToCognitoAttributeMap() map[string]string {
	attributes := make(map[string]string)
	if u.Name != nil {
		attributes["name"] = *u.Name
	}
	if u.PreferredUsername != nil {
		attributes["preferred_username"] = *u.PreferredUsername
	}
	if u.Birthdate != nil {
		attributes["birthdate"] = *u.Birthdate
	}

	return attributes
}
