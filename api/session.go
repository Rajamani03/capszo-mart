package api

import (
	"capszo-mart/database"
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
	SessionID    string `json:"session_id" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (server *Server) renewToken(ctx *gin.Context) {
	var request renewTokenRequest
	var err error

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// verify refresh token
	tokenPayload, err := server.token.VerifyToken(request.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// convert sessionID string to objectID
	objectID, err := primitive.ObjectIDFromHex(request.SessionID)
	if err != nil {
		err = errors.New("INVALID SESSION ID")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// update session
	accessToken, refreshToken, err := server.updateSession(objectID, tokenPayload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"access_token": accessToken, "refresh_token": refreshToken})
}

func (server *Server) getAuthTokens(userID string, tokenFor token.TokenFor) (accessToken string, refreshToken string, tokenPayload *token.Payload, err error) {
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
	tokenPayload, err = server.token.VerifyToken(refreshToken)
	if err != nil {
		return
	}

	return
}

func (server *Server) createSession(userID string, tokenFor token.TokenFor) (accessToken string, refreshToken string, sessionID string, err error) {
	var session database.Session
	db := server.mongoDB.Database("capszo")
	sessionColl := db.Collection(string(database.SessionColl))

	// get auth tokens
	accessToken, refreshToken, payload, err := server.getAuthTokens(userID, tokenFor)
	if err != nil {
		return
	}

	// store the session
	session.UserID = userID
	session.TokenID = payload.ID.String()
	session.TokenFor = payload.TokenFor
	session.LastRenewed = payload.IssuedAt
	result, err := sessionColl.InsertOne(context.TODO(), session)
	if err != nil {
		return
	}
	sessionID = getID(result.InsertedID)

	return
}

func (server *Server) updateSession(sessionID primitive.ObjectID, tokenPayload *token.Payload) (accessToken string, refreshToken string, err error) {
	var session database.Session
	db := server.mongoDB.Database("capszo")
	sessionColl := db.Collection(string(database.SessionColl))

	// get the refresh tokenID
	filter := bson.M{"_id": sessionID}
	if err = sessionColl.FindOne(context.TODO(), filter).Decode(&session); err != nil {
		return
	}

	// compare the session info with token payload data
	if session.UserID != tokenPayload.UserID || session.TokenID != tokenPayload.ID.String() || session.TokenFor != tokenPayload.TokenFor {
		err = errors.New("INVALID TOKEN FOR USER")
		return
	}

	// get auth tokens
	accessToken, refreshToken, newTokenPayload, err := server.getAuthTokens(tokenPayload.UserID, tokenPayload.TokenFor)
	if err != nil {
		return
	}

	// update the session info
	update := bson.M{"$set": bson.M{"token_id": newTokenPayload.ID.String(), "last_renewed": newTokenPayload.IssuedAt}}
	if _, err = sessionColl.UpdateByID(context.TODO(), sessionID, update); err != nil {
		return
	}

	return
}

func (server *Server) getTestToken(ctx *gin.Context) {
	var customer database.Customer
	var mart database.Mart
	var truck database.Truck
	var err error
	db := server.mongoDB.Database("capszo")
	customerColl := db.Collection(string(database.CustomerColl))
	martColl := db.Collection(string(database.MartColl))
	truckColl := db.Collection(string(database.TruckColl))

	if err = customerColl.FindOne(context.TODO(), bson.M{}).Decode(&customer); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if err = martColl.FindOne(context.TODO(), bson.M{}).Decode(&mart); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if err = truckColl.FindOne(context.TODO(), bson.M{}).Decode(&truck); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get access token
	cat, _, _, err := server.getAuthTokens(getID(customer.ID), token.CustomerAccess)
	if err != nil {
		err = errors.New("TOKEN QUERY ERROR")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	mat, _, _, err := server.getAuthTokens(getID(mart.ID), token.MartAccess)
	if err != nil {
		err = errors.New("TOKEN QUERY ERROR")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	tat, _, _, err := server.getAuthTokens(getID(truck.ID), token.TruckAccess)
	if err != nil {
		err = errors.New("TOKEN QUERY ERROR")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"access_tokens": bson.A{cat, mat, tat}})
}
