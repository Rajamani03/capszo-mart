package token

import (
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

type PasetoMaker struct {
	paseto       paseto.Token
	parser       paseto.Parser
	symmetricKey paseto.V4SymmetricKey
}

func NewPasetoMaker(symmetricKey string) (Maker, error) {
	var maker *PasetoMaker

	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	key, err := paseto.V4SymmetricKeyFromBytes([]byte(symmetricKey))
	if err != nil {
		return nil, err
	}

	maker = &PasetoMaker{
		paseto:       paseto.NewToken(),
		parser:       paseto.NewParser(),
		symmetricKey: key,
	}

	return maker, nil
}

func (maker *PasetoMaker) CreateToken(userID string, tokenFor TokenFor, tokenUse TokenUse, duration time.Duration) (string, error) {
	var payload *Payload
	var err error

	payload, err = NewPayload(userID, tokenFor, tokenUse, duration)
	if err != nil {
		return "", err
	}

	err = maker.paseto.Set("payload", payload)
	if err != nil {
		return "", err
	}
	maker.paseto.SetExpiration(time.Now().Add(duration))

	return maker.paseto.V4Encrypt(maker.symmetricKey, nil), nil
}

func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	var payload *Payload = &Payload{}
	var parsedToken *paseto.Token
	var err error

	parsedToken, err = maker.parser.ParseV4Local(maker.symmetricKey, token, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = parsedToken.Get("payload", payload)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
