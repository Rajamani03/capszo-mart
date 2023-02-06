package api

import (
	"capszo-mart/token"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type renewTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (server *Server) renewToken(ctx *gin.Context) {
	var request renewTokenRequest
	var user map[string]interface{}
	var collectionName string
	var err error
	db := server.mongoDB.Database("capszo")

	// validate request format
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// verify refresh token
	payload, err := server.token.VerifyToken(request.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// identify user to get collection name
	switch payload.TokenFor {
	case token.CustomerAccess:
		collectionName = "customers"
	case token.HaulerAccess:
		collectionName = "haulers"
	case token.MartAccess:
		collectionName = "marts"
	}
	coll := db.Collection(collectionName)

	// convert userID string to objectID
	objectID, err := primitive.ObjectIDFromHex(payload.UserID)
	if err != nil {
		return
	}

	// get the refresh tokenID
	filter := bson.D{{Key: "_id", Value: objectID}}
	err = coll.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("TOKEN QUERY ERROR")))
		return
	}

	// check if both request tokenID matches
	if user["refresh_token_id"] != payload.ID.String() {
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("TOKEN REUSED OR COMPROMISED")))
		return
	}

	// get access and refresh token
	accessToken, refreshToken, err := server.getAuthTokens(payload.UserID, payload.TokenFor)
	if err != nil {
		err = errors.New("TOKEN QUERY ERROR")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"access_token": accessToken, "refresh_token": refreshToken})
}

func (server *Server) getAuthTokens(userID string, tokenFor token.TokenFor) (accessToken string, refreshToken string, err error) {
	db := server.mongoDB.Database("capszo")
	var collectionName string

	// identify user to get table name
	switch tokenFor {
	case token.CustomerAccess:
		collectionName = "customers"
	case token.HaulerAccess:
		collectionName = "haulers"
	case token.MartAccess:
		collectionName = "marts"
	}
	coll := db.Collection(collectionName)

	// get access and refresh token duration
	accessTokenDuration, err := time.ParseDuration(server.config.AccessTokenDuration)
	if err != nil {
		return
	}
	refreshTokenDuration, err := time.ParseDuration(server.config.RefreshTokenDuration)
	if err != nil {
		return
	}

	// generate access token and refresh token
	accessToken, err = server.token.CreateToken(userID, tokenFor, token.AccessUse, accessTokenDuration)
	if err != nil {
		return
	}
	refreshToken, err = server.token.CreateToken(userID, tokenFor, token.RefreshUse, refreshTokenDuration)
	if err != nil {
		return
	}

	// get payload from new refresh token
	payload, err := server.token.VerifyToken(refreshToken)
	if err != nil {
		return
	}

	// convert userID string to objectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return
	}

	// store the new refresh tokenID
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "refresh_token_id", Value: payload.ID.String()}}}}
	_, err = coll.UpdateByID(context.TODO(), objectID, update)
	if err != nil {
		return
	}

	return
}

// func (server *Server) getTestToken(ctx *gin.Context) {
// 	// get access and refresh token
// 	accessToken, refreshToken, err := server.getAuthTokens(2, token.CustomerAccess)
// 	if err != nil {
// 		err = errors.New("TOKEN QUERY ERROR")
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}

// 	// response
// 	ctx.JSON(http.StatusOK, gin.H{"access_token": accessToken, "refresh_token": refreshToken})
// }
