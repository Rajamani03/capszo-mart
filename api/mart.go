package api

import (
	"capszo-mart/database"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (server *Server) getMartOrderPreference(ctx *gin.Context) {
	var mart database.Mart
	var err error
	db := server.mongoDB.Database("capszo")
	martColl := db.Collection(string(database.MartColl))

	// get query parameters
	martID := ctx.Query("mart-id")

	// convert validate key string to objectID
	objectID, err := primitive.ObjectIDFromHex(martID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get mart order preferences
	filter := bson.M{"_id": objectID}
	opts := options.FindOne().SetProjection(bson.M{"order_preferences": 1})
	if err = martColl.FindOne(context.TODO(), filter, opts).Decode(&mart); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, mart.OrderPreferences)
}
