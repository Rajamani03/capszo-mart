package api

import (
	"capszo-mart/database"
	"capszo-mart/token"
	"context"

	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func (server *Server) getCustomerInfo(ctx *gin.Context) {
	var customer database.Customer
	var err error
	db := server.mongoDB.Database("capszo")
	customerColl := db.Collection(string(database.CustomerColl))

	// get token payload
	tokenPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// get customer info
	filter := bson.M{"_id": tokenPayload.UserID}
	if err = customerColl.FindOne(context.TODO(), filter).Decode(&customer); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, customer)
}
