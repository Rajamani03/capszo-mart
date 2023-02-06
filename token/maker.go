package token

import "time"

type Maker interface {
	// creates a new token for specific username and duration
	CreateToken(userID string, tokenFor TokenFor, tokenUse TokenUse, duration time.Duration) (string, error)

	// checks if token is valid or not
	VerifyToken(token string) (*Payload, error)
}
