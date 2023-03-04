package api

import (
	"capszo-mart/database"
	"capszo-mart/token"
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (server *Server) getCustomerLoginOTP(ctx *gin.Context) {
	var request loginOTPRequest
	var err error
	db := server.mongoDB.Database("capszo")
	customerColl := db.Collection(string(database.CustomerColl))

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if mobile number exists
	var customer database.Customer
	filter := bson.D{{Key: "mobile_number", Value: request.MobileNumber}}
	err = customerColl.FindOne(context.TODO(), filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = errors.New("USER NOT FOUND")
			ctx.JSON(http.StatusBadRequest, errorResponse(err))

		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		}
		return
	}
	request.UserID = getID(customer.ID)

	// store login OTP
	validateKey, err := server.storeLoginOTP(request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"validate_key": validateKey})
}

func (server *Server) customerLogin(ctx *gin.Context) {
	var request loginRequest
	var loginInfo loginOTPRequest
	var customer database.Customer
	var err error
	db := server.mongoDB.Database("capszo")
	customerColl := db.Collection(string(database.CustomerColl))

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// validate login otp
	loginInfo, err = server.validateLoginOTP(request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// convert userID string to objectID
	objectID, err := primitive.ObjectIDFromHex(loginInfo.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get customer info
	filter := bson.D{{Key: "_id", Value: objectID}}
	err = customerColl.FindOne(context.TODO(), filter).Decode(&customer)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// create session and get access and refresh token
	sessionID, accessToken, refreshToken, err := server.createSession(loginInfo.UserID, token.CustomerAccess)
	if err != nil {
		err = errors.New("SESSION ERROR")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"session_id": sessionID, "access_token": accessToken, "refresh_token": refreshToken, "user_info": customer})
}
