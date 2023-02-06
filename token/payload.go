package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExpiredToken = errors.New("TOKEN EXPIRED")
	ErrInvalidToken = errors.New("INVALID TOKEN")
)

type TokenUse byte

const (
	AccessUse TokenUse = iota + 1
	RefreshUse
)

type TokenFor byte

const (
	AdminAccess TokenFor = iota + 1
	CustomerAccess
	HaulerAccess
	MartAccess
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	UserID    string    `json:"user_id"`
	TokenFor  TokenFor  `json:"token_for"`
	TokenUse  TokenUse  `json:"token_use"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewPayload(userID string, tokenFor TokenFor, tokenUse TokenUse, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		UserID:    userID,
		TokenFor:  tokenFor,
		TokenUse:  tokenUse,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
