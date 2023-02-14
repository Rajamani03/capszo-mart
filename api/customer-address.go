package api

import (
	"capszo-mart/database"
	"capszo-mart/token"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (server *Server) updateCustomerAddress(ctx *gin.Context) {
	var request database.Address
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

	// convert userID string to objectID
	objectID, err := primitive.ObjectIDFromHex(tokenPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// update customer address
	// update := bson.D{{Key: "$set", Value: bson.D{{Key: "home_address", Value: request}}}}
	update := bson.M{"$set": bson.M{"home_address": request}}
	// update := gin.H{"$set": gin.H{"home_address": request}}
	_, err = customerColl.UpdateByID(context.TODO(), objectID, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusAccepted, gin.H{"message": "address updated successfully"})
}
