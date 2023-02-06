package middleware

import (
	"capszo-mart/token"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CustomerAuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

		payload, err := authMiddleware(authorizationHeader, tokenMaker)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		if payload.TokenFor != token.CustomerAccess {
			err = errors.New("TOKEN NOT FOR CUSTOMER ACCESS")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
