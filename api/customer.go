package api

import (
	"capszo-mart/database"
	"capszo-mart/token"
	"context"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (server *Server) getCustomerInfo(ctx *gin.Context) {
	var customer database.Customer
	var err error
	db := server.mongoDB.Database("capszo")
	customerColl := db.Collection(string(database.CustomerColl))

	// get token payload
	tokenPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// convert user id string to object id
	objectID, err := primitive.ObjectIDFromHex(tokenPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get customer info
	filter := bson.M{"_id": objectID}
	if err = customerColl.FindOne(context.TODO(), filter).Decode(&customer); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, customer)
}

func (server *Server) updateCustomerInfo(ctx *gin.Context) {
	var request database.Customer
	var err error
	db := server.mongoDB.Database("capszo")
	customerColl := db.Collection(string(database.CustomerColl))

	// get token payload
	tokenPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// convert user id string to object id
	objectID, err := primitive.ObjectIDFromHex(tokenPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// update customer info
	request.ID = objectID
	request.UpdatedAt = time.Now()
	update := bson.M{"$set": request}
	filter := bson.M{"_id": objectID}
	_, err = customerColl.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"message": "customer updated successfully", "mobile_number": request.MobileNumber})

}
