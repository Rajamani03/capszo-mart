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
	ItemID   string  `json:"item_id" bson:"item_id" binding:"required,alphanum"`
	Quantity float64 `json:"quantity" bson:"quantity" binding:"required,numeric"`
}

type OrderItem struct {
	ItemID       string  `json:"item_id" bson:"item_id" binding:"required,alphanum"`
	Quantity     float64 `json:"quantity" bson:"quantity" binding:"required,numeric"`
	Mrp          float64 `json:"mrp" bson:"mrp"`
	SellingPrice float64 `json:"selling_price" bson:"selling_price"`
	CostPrice    float64 `json:"cost_price" bson:"cost_price"`
}

type Rating struct {
	SumOfRatings    float64 `json:"sum_of_ratings"`
	NumberOfRatings uint64  `json:"number_of_ratings"`
}

type Location struct {
	Lattiude  float64 `json:"lattitude" binding:"required,numeric"`
	Longitude float64 `json:"longitude" binding:"required,numeric"`
}

type MartOrderPreference struct {
	PackagingCharge float64 `json:"packaging_charge" bson:"packaging_charge"`
	DeliveryCharge  float64 `json:"delivery_charge" bson:"delivery_charge"`
}

type GST struct {
	SGST float64 `json:"sgst" bson:"sgst"`
	CGST float64 `json:"cgst" bson:"cgst"`
}

type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
	Others Gender = "others"
)

type ItemUnit string

const (
	Milligram  ItemUnit = "mg"
	Gram       ItemUnit = "g"
	Kilogram   ItemUnit = "kg"
	Millilitre ItemUnit = "ml"
	Litre      ItemUnit = "l"
)

type OrderStatus string

const (
	OrderConfirmed OrderStatus = "confirmed"
	OrderDelivered OrderStatus = "delivered"
	OrderCancelled OrderStatus = "cancelled"
)

type MartStatus string

const (
	MartOpen   MartStatus = "open"
	MartClosed MartStatus = "closed"
)

type TruckStatus string

const (
	TruckAvailable   TruckStatus = "available"
	TruckDelivering  TruckStatus = "delivering"
	TruckUnavailable TruckStatus = "unavailable"
)

type HaulerStatus string

const (
	HaulerAvailable  HaulerStatus = "available"
	HaulerDelivering HaulerStatus = "delivering"
	HaulerOnTruck    HaulerStatus = "on_truck"
	HaulerOff        HaulerStatus = "off"
)
