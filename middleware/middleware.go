package middleware

import (
	"capszo-mart/token"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func authMiddleware(authorizationHeader string, tokenMaker token.Maker) (payload *token.Payload, err error) {
	// check if auth header is present
	if len(authorizationHeader) == 0 {
		err = errors.New("NO AUTHORIZATION HEADER")
		return
	}

	// check auth header format
	fields := strings.Fields(authorizationHeader)
	if len(fields) < 2 {
		err = errors.New("INVALID AUTHORIZATION HEADER FORMAT")
		return
	}

	// check the auth type
	authorizationType := strings.ToLower(fields[0])
	if authorizationType != authorizationTypeBearer {
		err = fmt.Errorf("UNSUPPORTED AUTHORIZATION TYPE %s", authorizationType)
		return
	}

	// verify the access token
	accessToken := fields[1]
	payload, err = tokenMaker.VerifyToken(accessToken)
	if err != nil {
		err = errors.New("INVALID ACCESS TOKEN")
		return
	}

	// validate token use as access
	if payload.TokenUse != token.AccessUse {
		err = errors.New("INVALID TOKEN FOR ACCESS")
		return
	}

	return payload, nil
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
