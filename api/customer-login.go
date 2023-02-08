package api

import (
	"capszo-mart/database"
	"capszo-mart/token"
	"capszo-mart/util"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type loginOTPRequest struct {
	UserID       string    `json:"user_id" bson:"user_id"`
	MobileNumber string    `json:"mobile_number" bson:"mobile_number" binding:"required,numeric,len=10"`
	OTP          string    `json:"otp" bson:"otp"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}

func (server *Server) getCustomerLoginOTP(ctx *gin.Context) {
	var request loginOTPRequest
	var err error
	db := server.mongoDB.Database("capszo")
	loginInfoColl := db.Collection("login_info")
	customerColl := db.Collection("customers")

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if mobile number exists
	var customer map[string]interface{}
	filter := bson.D{{Key: "mobile_number", Value: request.MobileNumber}}
	err = customerColl.FindOne(context.TODO(), filter).Decode(&customer)
	if err != nil {
		err = errors.New("USER NOT FOUND")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	request.UserID = getID(customer["_id"])

	// generate otp
	// otp := util.GetOTP(6)
	otp := "654321"
	request.OTP = util.Hash(otp + server.config.Salt)

	// TTL index
	model := mongo.IndexModel{
		Keys:    bson.M{"created_at": 1},
		Options: options.Index().SetExpireAfterSeconds(60),
	}
	_, err = loginInfoColl.Indexes().CreateOne(ctx, model)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// store the login data to login_info collection
	request.CreatedAt = time.Now()
	result, err := loginInfoColl.InsertOne(context.TODO(), request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"validate_key": getID(result.InsertedID)})
}

func (server *Server) customerLogin(ctx *gin.Context) {
	var request LoginRequest
	var loginInfo loginOTPRequest
	var customer database.Customer
	var err error
	db := server.mongoDB.Database("capszo")
	loginInfoColl := db.Collection("login_info")
	customerColl := db.Collection("customers")

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// convert validate key string to objectID
	objectID, err := primitive.ObjectIDFromHex(request.ValidateKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get login info
	filter := bson.D{{Key: "_id", Value: objectID}}
	err = loginInfoColl.FindOne(context.TODO(), filter).Decode(&loginInfo)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// validate input and saved otp
	hotp := util.Hash(request.OTP + server.config.Salt)
	if hotp != loginInfo.OTP {
		err = errors.New("INCORRECT OTP")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// convert validate key string to objectID
	objectID, err = primitive.ObjectIDFromHex(loginInfo.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get customer info
	filter = bson.D{{Key: "_id", Value: objectID}}
	err = customerColl.FindOne(context.TODO(), filter).Decode(&customer)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// get access and refresh token
	accessToken, refreshToken, err := server.getAuthTokens(loginInfo.UserID, token.CustomerAccess)
	if err != nil {
		err = errors.New("TOKEN QUERY ERROR")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"access_token": accessToken, "refresh_token": refreshToken, "user_data": customer})
}
