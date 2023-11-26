package api

import (
	"capszo-mart/blueprint"
	"capszo-mart/database"
	"capszo-mart/token"
	"capszo-mart/util"
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
	"go.mongodb.org/mongo-driver/mongo/options"
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
	groceryColl := db.Collection(string(database.GroceryColl))
	customerColl := db.Collection(string(database.CustomerColl))
	martColl := db.Collection(string(database.MartColl))
	orderColl := db.Collection(string(database.OrderColl))

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
	order.PackagingCharge = mart.OrderPreferences.PackagingCharge
	order.DeliveryCharge = mart.OrderPreferences.DeliveryCharge
	order.Tax = database.GST{SGST: 0, CGST: 0}
	order.TruckTips = math.Abs(request.TruckTips)
	order.Donation = math.Abs(request.Donation)
	order.Discount = 0
	order.Total += (order.DeliveryCharge + order.PackagingCharge + order.Tax.SGST + order.Tax.CGST + order.TruckTips + order.Donation) - order.Discount
	order.Total = math.Ceil(order.Total)
	// order.Total = math.Ceil(order.Total*100) / 100
	order.OrderedDate = time.Now()
	order.DeliveryAddress = request.DeliveryAddress
	order.DeliveryDate = time.Now().Add(time.Hour * 24)
	order.Status = database.OrderConfirmed
	order.Coupon = request.Coupon
	order.OTP = util.GetOTP(4)
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
	ctx.JSON(http.StatusCreated, order)
}

type getOrderRequest struct {
	MartID               string `json:"mart_id"`
	TruckID              string `json:"truck_id"`
	CustomerMobileNumber string `json:"customer_mobile_number"`
	OrderedDate          string `json:"ordered_date"`
	Status               string `json:"order_status"`
}

func (server *Server) getOrders(ctx *gin.Context) {
	var request getOrderRequest
	var orders []database.Order
	var err error
	db := server.mongoDB.Database("capszo")
	orderColl := db.Collection(string(database.OrderColl))

	// get token payload
	tokenPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// set the filter wrt request
	var filter = make(primitive.M)

	if request.MartID != "" {
		filter["mart_id"] = request.MartID
	}
	if request.TruckID != "" {
		filter["truck_id"] = request.TruckID
	}
	if request.CustomerMobileNumber != "" {
		filter["customer_mobile_number"] = request.CustomerMobileNumber
	}

	if request.OrderedDate != "" {
		orderedDate, err := time.Parse("2006-01-02", request.OrderedDate)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		dateFilter := bson.M{"$gt": primitive.NewDateTimeFromTime(orderedDate.AddDate(0, 0, -1)), "$lt": primitive.NewDateTimeFromTime(orderedDate.AddDate(0, 0, 1))}
		filter["ordered_date"] = dateFilter
	}
	if request.Status != "" {
		filter["status"] = request.Status
	}

	// filter by user id
	switch tokenPayload.TokenFor {
	case token.CustomerAccess:
		filter["customer_id"] = tokenPayload.UserID
	case token.MartAccess:
		filter["mart_id"] = tokenPayload.UserID
	case token.TruckAccess:
		filter["truck_id"] = tokenPayload.UserID
	}

	// find order by filters
	cursor, err := orderColl.Find(context.TODO(), filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if err = cursor.All(context.TODO(), &orders); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// transform order json based on views
	var transformedOrders []map[string]interface{}
	for _, order := range orders {
		switch tokenPayload.TokenFor {
		case token.AdminAccess:
			transformedOrder, err := blueprint.OrderTransform(order, blueprint.AdminOrder)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}
			transformedOrders = append(transformedOrders, transformedOrder)
		case token.CustomerAccess:
			transformedOrder, err := blueprint.OrderTransform(order, blueprint.CustomerOrder)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}
			transformedOrders = append(transformedOrders, transformedOrder)
		case token.MartAccess:
			transformedOrder, err := blueprint.OrderTransform(order, blueprint.MartOrder)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}
			transformedOrders = append(transformedOrders, transformedOrder)
		default:
			transformedOrder, err := blueprint.OrderTransform(order, blueprint.MartOrder)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}
			transformedOrders = append(transformedOrders, transformedOrder)
		}
	}

	// response
	ctx.JSON(http.StatusOK, transformedOrders)
}

func (server *Server) getOrder(ctx *gin.Context) {
	var order database.Order
	db := server.mongoDB.Database("capszo")
	orderColl := db.Collection(string(database.OrderColl))

	// get token payload
	tokenPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// get mart_id from params
	martID := ctx.Param("id")

	var filter = make(primitive.M)

	// filter by user id
	switch tokenPayload.TokenFor {
	case token.CustomerAccess:
		filter["customer_id"] = tokenPayload.UserID
	case token.MartAccess:
		filter["mart_id"] = tokenPayload.UserID
	case token.TruckAccess:
		filter["truck_id"] = tokenPayload.UserID
	}

	// get order using _id
	objectID, err := primitive.ObjectIDFromHex(martID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	filter["_id"] = objectID
	if err = orderColl.FindOne(context.TODO(), filter).Decode(&order); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// transform order json based on views
	var transformedOrder map[string]interface{}
	switch tokenPayload.TokenFor {
	case token.AdminAccess:
		transformedOrder, err = blueprint.OrderTransform(order, blueprint.AdminOrder)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	case token.CustomerAccess:
		transformedOrder, err = blueprint.OrderTransform(order, blueprint.CustomerOrder)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	case token.MartAccess:
		transformedOrder, err = blueprint.OrderTransform(order, blueprint.MartOrder)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	default:
		transformedOrder, err = blueprint.OrderTransform(order, blueprint.MartOrder)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	// response
	ctx.JSON(http.StatusOK, transformedOrder)
}

type deliverOrderRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	OTP     string `json:"otp" binding:"required,numeric,len=4"`
}

func (server *Server) deliverOrder(ctx *gin.Context) {
	var request deliverOrderRequest
	var order database.Order
	var err error
	db := server.mongoDB.Database("capszo")
	orderColl := db.Collection(string(database.OrderColl))

	// get token payload
	tokenPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// get order using _id
	objectID, err := primitive.ObjectIDFromHex(request.OrderID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get otp of the specific order
	filter := bson.M{"_id": objectID, "mart_id": tokenPayload.UserID}
	opts := options.FindOne().SetProjection(bson.M{"otp": 1})
	if err = orderColl.FindOne(context.TODO(), filter, opts).Decode(&order); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if order.OTP != request.OTP {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("INVALID OTP")))
		return
	}

	// update order status
	update := bson.M{"$set": bson.M{"status": database.OrderDelivered}}
	_, err = orderColl.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"message": "order delivered"})
}
