package database

type Address struct {
	FlatNumber  string  `json:"flat_no" bson:"flat_no" binding:"required,alphanum"`
	FloorNumber string  `json:"floor_no" bson:"floor_no" binding:"required,alphanum"`
	MainAddress string  `json:"main_address" bson:"main_address" binding:"required"`
	District    string  `json:"district" bson:"district" binding:"required,alpha"`
	State       string  `json:"state" bson:"state" binding:"required,alpha"`
	Pincode     string  `json:"pincode" bson:"pincode" binding:"required,numeric,len=6"`
	Lattitude   float64 `json:"lattitude" bson:"lattitude" binding:"required,numeric"`
	Longitude   float64 `json:"longitude" bson:"longitude" binding:"required,numeric"`
}

type BasketItem struct {
	ItemID   string  `json:"item_id" binding:"required,numeric"`
	Quantity float64 `json:"quantity" binding:"required,numeric"`
}

type OrderItem struct {
	ItemID       uint64  `json:"item_id" binding:"required,numeric"`
	Quantity     float64 `json:"quantity" binding:"required,numeric"`
	Mrp          float64 `json:"mrp"`
	SellingPrice float64 `json:"price"`
}

type OrderItemPrivateData struct {
	ItemID    uint64  `json:"item_id"`
	CostPrice float64 `json:"cost_price"`
}

type Rating struct {
	SumOfRatings    float64 `json:"sum_of_ratings"`
	NumberOfRatings uint64  `json:"number_of_ratings"`
}

type Location struct {
	Lattiude  float64 `json:"lattitude" binding:"required,numeric"`
	Longitude float64 `json:"longitude" binding:"required,numeric"`
}

// type Gender byte

// const (
// 	Male Gender = iota
// 	Female
// 	Others
// )

// type ItemUnit byte

// const (
// 	Milligram ItemUnit = iota
// 	Gram
// 	Kilogram
// 	Millilitre
// 	Litre
// )

// type OrderStatus byte

// const (
// 	Confirmed OrderStatus = iota
// 	Delivered
// 	Cancelled
// )

// type HaulerStatus byte

// const (
// 	Available HaulerStatus = iota
// 	Delivering
// 	Off
// )
