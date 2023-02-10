package database

import (
	"time"
)

type Customer struct {
	ID             interface{}  `json:"-" bson:"_id,omitempty"`
	Name           string       `json:"name" bson:"name" binding:"required,alpha"`
	MobileNumber   string       `json:"mobile_number" bson:"mobile_number" binding:"required,numeric,len=10"`
	EmailAddress   string       `json:"email_address" bson:"email_address"`
	Address        Address      `json:"home_address" bson:"home_address"`
	Gender         Gender       `json:"gender" bson:"gender"`
	BirthDate      time.Time    `json:"birth_date" bson:"birth_date"`
	GroceryBasket  []BasketItem `json:"grocery_basket" bson:"grocery_basket"`
	NearestMartID  string       `json:"nearest_mart_id" bson:"nearest_mart_id"`
	Wishlist       []string     `json:"wishlist" bson:"wishlist"`
	RefreshTokenID string       `json:"-" bson:"refresh_token_id"`
	CreatedAt      time.Time    `json:"-" bson:"created_at"`
	UpdatedAt      time.Time    `json:"-" bson:"updated_at"`
}

type Item struct {
	ID              interface{} `json:"item_id" bson:"_id,omitempty"`
	MartID          string      `json:"-" bson:"mart_id"`
	Name            string      `json:"name" bson:"name"`
	ImageURL        string      `json:"image_url" bson:"image_url"`
	Mrp             float64     `json:"mrp" bson:"mrp"`
	SellingPrice    float64     `json:"price" bson:"selling_price"`
	CostPrice       float64     `json:"-" bson:"cost_price"`
	Quantity        float64     `json:"quantity" bson:"quantity"`
	Unit            ItemUnit    `json:"unit" bson:"unit"`
	StepQuantity    float32     `json:"step_quantity" bson:"step_quantity"`
	IndividualLimit float64     `json:"individual_limit" bson:"individual_limit"`
	StockQuantity   float64     `json:"-" bson:"stock_quantity"`
	Brand           string      `json:"brand" bson:"brand"`
	Category        string      `json:"category" bson:"category"`
	SubCategory     string      `json:"sub_category" bson:"sub_category"`
	OtherNames      []string    `json:"other_names" bson:"other_names"`
	CreatedAt       time.Time   `json:"-" bson:"created_at"`
	UpdatedAt       time.Time   `json:"-" bson:"updated_at"`
}

type Order struct {
	ID                   interface{} `json:"order_id" bson:"_id,omitempty"`
	CustomerID           string      `json:"-" bson:"customer_id"`
	MartID               string      `json:"-" bson:"mart_id" gorm:"column:mart_id" binding:"required,numeric"`
	CustomerMobileNumber string      `json:"customer_mobile_number" bson:"customer_mobile_number" binding:"required,numeric"`
	Items                []OrderItem `json:"grocery_items" bson:"grocery_items" binding:"required"`
	PackagingCharge      float64     `json:"packaging_charge" bson:"packaging_charge"`
	DeliveryCharge       float64     `json:"delivery_charge" bson:"delivery_charge"`
	Tax                  float64     `json:"tax" bson:"tax"`
	TruckTips            float64     `json:"truck_tips" bson:"truck_tips" binding:"required,numeric"`
	Donation             float64     `json:"donation" bson:"donation" binding:"required,numeric"`
	Discount             float64     `json:"discount" bson:"discount"`
	Total                float64     `json:"total" bson:"total"`
	OrderedDate          time.Time   `json:"ordered_date" bson:"ordered_date"`
	DeliveryAddress      Address     `json:"delivery_address" bson:"delivery_address" binding:"required"`
	DeliveryDate         time.Time   `json:"delivery_date" bson:"delivery_date"`
	Status               OrderStatus `json:"order_status" bson:"order_status"`
	Coupon               string      `json:"coupon" bson:"coupon" binding:"alphanum"`
	OnlinePayment        string      `json:"online_payment" bson:"online_payment"`
	TruckID              string      `json:"-" bson:"truck_id" gorm:"column:truck_id"`
	Distance             float32     `json:"-" bson:"distance"`
}

type Mart struct {
	ID              interface{} `json:"mart_id" bson:"_id,omitempty"`
	Name            string      `json:"name" bson:"name"`
	MobileNumber    string      `json:"mobile_number" bson:"mobile_number" binding:"required,numeric,len=10"`
	Address         Address     `json:"mart_address" bson:"address"`
	Status          MartStatus  `json:"mart_status" bson:"status"`
	PackagingCharge float64     `json:"packaging_charge" bson:"packaging_charge"`
	DeliveryCharge  float64     `json:"delivery_charge" bson:"delivery_charge"`
	RefreshTokenID  string      `json:"-" bson:"refresh_token_id"`
	CreatedAt       time.Time   `json:"-" bson:"created_at"`
	UpdatedAt       time.Time   `json:"-" bson:"updated_at"`
}

type Truck struct {
	ID            interface{} `json:"truck_id" bson:"_id,omitempty"`
	Name          string      `json:"name" bson:"name"`
	MobileNumber  string      `json:"mobile_number" bson:"mobile_number" binding:"required,numeric,len=10"`
	HaulerIDs     []string    `json:"hauler_ids" bson:"hauler_ids"`
	Brand         string      `json:"brand" bson:"brand"`
	Model         string      `json:"model" bson:"model"`
	VehicleNumber string      `json:"vehicle_number" bson:"vehicle_number"`
	Status        TruckStatus `json:"truck_status" bson:"truck_status"`
	CreatedAt     time.Time   `json:"-" bson:"created_at"`
	UpdatedAt     time.Time   `json:"-" bson:"updated_at"`
}

type Hauler struct {
	ID             interface{}  `json:"hauler_id" bson:"_id,omitempty"`
	Name           string       `json:"name" binding:"required,alpha"`
	MobileNumber   string       `json:"mobile_number" binding:"required,numeric,len=10"`
	Gender         string       `json:"gender"`
	Rating         string       `json:"rating"`
	Location       string       `json:"hauler_location"`
	Status         HaulerStatus `json:"hauler_status"`
	RefreshTokenID string       `json:"-" gorm:"column:refresh_token_id"`
	CreatedAt      time.Time    `json:"-" bson:"created_at"`
	UpdatedAt      time.Time    `json:"-" bson:"updated_at"`
}

type GroceryMisc struct {
	ID uint64 `json:"-" gorm:"column:id"`
}
