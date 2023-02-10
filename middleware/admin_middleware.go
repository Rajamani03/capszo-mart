package middleware

import (
	"capszo-mart/token"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminAuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

		payload, err := authMiddleware(authorizationHeader, tokenMaker)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		if payload.TokenFor != token.AdminAccess {
			err = errors.New("TOKEN NOT FOR ADMIN ACCESS")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
