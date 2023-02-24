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

type signupOTPRequest struct {
	Name         string    `json:"name" bson:"name" binding:"required"`
	MobileNumber string    `json:"mobile_number" bson:"mobile_number" binding:"required,numeric,len=10"`
	OTP          string    `json:"otp" bson:"otp"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}

func (server *Server) getCustomerSignupOTP(ctx *gin.Context) {
	var request signupOTPRequest
	var err error
	db := server.mongoDB.Database("capszo")
	signupInfoColl := db.Collection("signup_info")
	customerColl := db.Collection(string(database.CustomerColl))

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if mobile number exists
	filter := bson.D{{Key: "mobile_number", Value: request.MobileNumber}}
	customerCount, err := customerColl.CountDocuments(context.TODO(), filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if customerCount > 0 {
		err = errors.New("USER ALREADY EXISTS")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// generate otp
	// otp := util.GetOTP(6)
	otp := "654321"
	request.OTP = util.Hash(otp + server.config.Salt)

	// TTL index
	model := mongo.IndexModel{
		Keys:    bson.M{"created_at": 1},
		Options: options.Index().SetExpireAfterSeconds(60),
	}
	_, err = signupInfoColl.Indexes().CreateOne(ctx, model)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// store the signup data to signup_info collection
	request.CreatedAt = time.Now()
	result, err := signupInfoColl.InsertOne(context.TODO(), request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"validate_key": getID(result.InsertedID)})
}

type loginRequest struct {
	ValidateKey string `json:"validate_key" binding:"required"`
	OTP         string `json:"otp" binding:"required,numeric,len=6"`
}

func (server *Server) customerSignup(ctx *gin.Context) {
	var request loginRequest
	var signupInfo signupOTPRequest
	var customer database.Customer
	var err error
	db := server.mongoDB.Database("capszo")
	signupInfoColl := db.Collection("signup_info")
	customerColl := db.Collection(string(database.CustomerColl))

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

	// get signup info
	filter := bson.D{{Key: "_id", Value: objectID}}
	err = signupInfoColl.FindOne(context.TODO(), filter).Decode(&signupInfo)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// validate input and saved otp
	hotp := util.Hash(request.OTP + server.config.Salt)
	if hotp != signupInfo.OTP {
		err = errors.New("INCORRECT OTP")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// store customer data in db
	customer.Name = signupInfo.Name
	customer.MobileNumber = signupInfo.MobileNumber
	customer.CreatedAt = time.Now()
	result, err := customerColl.InsertOne(context.TODO(), customer)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// create session and get access and refresh token
	sessionID, accessToken, refreshToken, err := server.createSession(getID(result.InsertedID), token.CustomerAccess)
	if err != nil {
		err = errors.New("SESSION ERROR")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusCreated, gin.H{"session_id": sessionID, "access_token": accessToken, "refresh_token": refreshToken, "user_info": customer})
}
