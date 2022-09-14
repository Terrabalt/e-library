package sessiontoken

import (
	"encoding/json"
	"errors"
)

type AccessClaimsSchema struct {
	Email string `json:"sub"`
}

func (token AccessClaimsSchema) ToInterface() (inter map[string]interface{}, err error) {
	js, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(js, &inter)
	if err != nil {
		return nil, err
	}
	return
}

func (token *AccessClaimsSchema) FromInterface(inter map[string]interface{}) error {
	js, err := json.Marshal(inter)
	if err != nil {
		return err
	}

	err = json.Unmarshal(js, token)
	if err != nil {
		return err
	}
	return nil
}

var ErrAccessTokenMalformed = errors.New("important data missing from session token schema")

func (token AccessClaimsSchema) CheckMalform() error {
	if token.Email == "" {
		return ErrAccessTokenMalformed
	}
	return nil
}
