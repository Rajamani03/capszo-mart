package api

import (
	"capszo-mart/database"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func (server *Server) getAllItems(ctx *gin.Context) {
	var items []database.GroceryItem
	db := server.mongoDB.Database("capszo")
	groceriesColl := db.Collection("groceries")

	// get mart_id from params
	martID := ctx.Param("mart-id")

	// get all grocery items from groceries collections
	filter := bson.D{{Key: "mart_id", Value: martID}}
	cursor, err := groceriesColl.Find(context.TODO(), filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if err = cursor.All(context.TODO(), &items); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, items)
}
