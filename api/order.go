package api

import (
	"capszo-mart/database"
	"capszo-mart/token"
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type orderRequest struct {
	MartID               string               `json:"mart_id" binding:"required,alphanum"`
	CustomerMobileNumber string               `json:"customer_mobile_number" binding:"required,numeric,len=10"`
	Items                []database.OrderItem `json:"grocery_items" binding:"required"`
	TruckTips            float64              `json:"truck_tips" binding:"required,numeric"`
	Donation             float64              `json:"donation" binding:"required,numeric"`
	DeliveryAddress      database.Address     `json:"delivery_address" binding:"required"`
	Coupon               string               `json:"coupon" binding:"alphanum"`
}

func (server *Server) order(ctx *gin.Context) {
	var request orderRequest
	var order database.Order
	var err error
	db := server.mongoDB.Database("capszo")
	groceryColl := db.Collection("groceries")
	customerColl := db.Collection("customers")
	martColl := db.Collection("marts")
	orderColl := db.Collection("mart_orders")

	// get token payload
	tokenPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// make a slice of itemids with the basket items
	var itemIDs []primitive.ObjectID
	for _, item := range request.Items {
		objectID, err := primitive.ObjectIDFromHex(item.ItemID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		itemIDs = append(itemIDs, objectID)
	}

	// get details from DB of all items in the basket
	var itemsData []database.Item
	filter := bson.M{"_id": bson.M{"$in": itemIDs}}
	cursor, err := groceryColl.Find(context.TODO(), filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if err = cursor.All(context.TODO(), &itemsData); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get mart data
	var mart database.Mart
	objectID, err := primitive.ObjectIDFromHex(request.MartID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	filter = bson.M{"_id": objectID}
	if err = martColl.FindOne(context.TODO(), filter).Decode(&mart); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// create an item map for fast access
	itemMap := make(map[string]database.Item)
	for _, item := range itemsData {
		itemMap[getID(item.ID)] = item
	}

	// check if the request item quantity exceeds stock quantity
	for _, requestItem := range request.Items {
		itemData := itemMap[requestItem.ItemID]
		if requestItem.Quantity > itemData.StockQuantity {
			err = fmt.Errorf("%v-%v QUANTITY IS EXCEEDING THE STOCK", itemData.Name, getID(itemData.ID))
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	}

	// iterate over order items
	order.Total = 0
	var updateModels []mongo.WriteModel
	for index, requestItem := range request.Items {
		itemData := itemMap[requestItem.ItemID]

		// fill the incomplete fields for order items and private data
		request.Items[index].Mrp = itemData.Mrp
		request.Items[index].SellingPrice = itemData.SellingPrice
		request.Items[index].CostPrice = itemData.CostPrice

		// calculate order total
		order.Total += itemData.SellingPrice * requestItem.Quantity

		// append all the operations to update models
		filter := bson.M{"_id": itemData.ID}
		update := bson.M{"$set": bson.M{"stock_quantity": itemData.StockQuantity - requestItem.Quantity}}
		updateModel := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
		updateModels = append(updateModels, updateModel)
	}

	// update the stock quantity of each item in DB
	_, err = groceryColl.BulkWrite(context.TODO(), updateModels)
	if err != nil {
		err = errors.New("ITEM STOCK QUANTITY UPDATE ERROR")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// fill all the required fields in order data for storing in DB
	order.CustomerID = tokenPayload.UserID
	order.MartID = request.MartID
	order.CustomerMobileNumber = request.CustomerMobileNumber
	order.Items = request.Items
	order.PackagingCharge = mart.PackagingCharge
	order.DeliveryCharge = mart.DeliveryCharge
	order.Tax = 0
	order.TruckTips = math.Abs(request.TruckTips)
	order.Donation = math.Abs(request.Donation)
	order.Discount = 0
	order.Total += (order.DeliveryCharge + order.PackagingCharge + order.Tax + order.TruckTips + order.Donation) - order.Discount
	order.Total = math.Ceil(order.Total)
	// order.Total = math.Ceil(order.Total*100) / 100
	order.OrderedDate = time.Now()
	order.DeliveryAddress = request.DeliveryAddress
	order.DeliveryDate = time.Now().Add(time.Hour * 24)
	order.Status = database.OrderConfirmed
	order.Coupon = request.Coupon
	order.OnlinePayment = "{}"
	order.TruckID = ""
	order.Distance = 0

	// store order in DB
	result, err := orderColl.InsertOne(context.TODO(), order)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	order.ID = result.InsertedID

	// clear the grocery basket of customer in DB
	objectID, err = primitive.ObjectIDFromHex(tokenPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	update := bson.M{"$set": bson.M{"grocery_basket": bson.A{}}}
	_, err = customerColl.UpdateByID(context.TODO(), objectID, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusCreated, gin.H{"order_info": order})
}

// func getQuantityGrams(unit string, quantity float64) float64 {
// 	switch unit {
// 	case "mg":
// 		return quantity * 0.001
// 	case "kg":
// 		return quantity * 1000
// 	case "l":
// 		return quantity * 1000
// 	default:
// 		return quantity
// 	}
// }
