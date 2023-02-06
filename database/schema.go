package database

import (
	"time"
)

type Customer struct {
	Name           string       `json:"name" bson:"name" binding:"required,alpha"`
	MobileNumber   string       `json:"mobile_number" bson:"mobile_number" binding:"required,numeric,len=10"`
	EmailAddress   string       `json:"email_address" bson:"email_address"`
	Address        Address      `json:"home_address" bson:"home_address"`
	Gender         string       `json:"gender" bson:"gender"`
	BirthDate      time.Time    `json:"birth_date" bson:"birth_date"`
	GroceryBasket  []BasketItem `json:"grocery_basket" bson:"grocery_basket"`
	NearestMartID  string       `json:"nearest_mart_id" bson:"nearest_mart_id"`
	Wishlist       []string     `json:"wishlist" bson:"wishlist"`
	RefreshTokenID string       `json:"-" bson:"refresh_token_id"`
	CreatedAt      time.Time    `json:"-" bson:"created_at"`
	UpdatedAt      time.Time    `json:"-" bson:"updated_at"`
	DeletedAt      time.Time    `json:"-" bson:"deleted_at"`
}

type GroceryItem struct {
	ID              uint64    `json:"grocery_item_id" gorm:"column:id;type:BIGINT UNSIGNED NOT NULL;AUTO_INCREMENT;primaryKey"`
	MartID          uint64    `json:"-" gorm:"column:mart_id"`
	Name            string    `json:"name"`
	ImageURL        string    `json:"image_url" gorm:"column:image_url"`
	Mrp             float64   `json:"mrp"`
	SellingPrice    float64   `json:"price"`
	CostPrice       float64   `json:"-"`
	Quantity        float64   `json:"quantity"`
	Unit            string    `json:"unit"`
	StepQuantity    float32   `json:"step_quantity"`
	IndividualLimit float64   `json:"individual_limit"`
	StockQuantity   float64   `json:"-"`
	Category        string    `json:"category"`
	Brand           string    `json:"brand"`
	SubCategory     string    `json:"sub_category"`
	OtherNames      string    `json:"other_names"`
	CreatedAt       time.Time `json:"-" bson:"created_at"`
	UpdatedAt       time.Time `json:"-" bson:"updated_at"`
	DeletedAt       time.Time `json:"-" bson:"deleted_at"`
}

type GroceryOrder struct {
	ID                   uint64    `json:"grocery_order_id" gorm:"column:id;type:BIGINT UNSIGNED NOT NULL;AUTO_INCREMENT;primaryKey"`
	CustomerID           uint64    `json:"-"`
	MartID               uint64    `json:"-" gorm:"column:mart_id" binding:"required,numeric"`
	CustomerMobileNumber string    `json:"customer_mobile_number" binding:"required,numeric"`
	Items                string    `json:"grocery_items" binding:"required"`
	ItemsPrivateData     string    `json:"-"`
	PackagingCharge      float64   `json:"packaging_charge"`
	DeliveryCharge       float64   `json:"delivery_charge"`
	Tax                  float64   `json:"tax"`
	TruckTips            float64   `json:"truck_tips" binding:"required,numeric"`
	Donation             float64   `json:"donation" binding:"required,numeric"`
	Discount             float64   `json:"discount"`
	Total                float64   `json:"total"`
	OrderedDate          time.Time `json:"ordered_date"`
	DeliveryAddress      string    `json:"delivery_address" binding:"required"`
	DeliveryDate         time.Time `json:"delivery_date"`
	Status               string    `json:"order_status"`
	Coupon               string    `json:"coupon" binding:"alphanum"`
	OnlinePayment        string    `json:"online_payment"`
	TruckID              uint64    `json:"-" gorm:"column:truck_id"`
	Distance             float32   `json:"-"`
}

type Hauler struct {
	ID             uint64    `json:"hauler_id"`
	Name           string    `json:"name" binding:"required,alpha"`
	MobileNumber   string    `json:"mobile_number" binding:"required,numeric,len=10"`
	Gender         string    `json:"gender"`
	Rating         string    `json:"rating"`
	Location       string    `json:"hauler_location"`
	Status         string    `json:"hauler_status"`
	RefreshTokenID string    `json:"-" gorm:"column:refresh_token_id"`
	CreatedAt      time.Time `json:"-" bson:"created_at"`
	UpdatedAt      time.Time `json:"-" bson:"updated_at"`
	DeletedAt      time.Time `json:"-" bson:"deleted_at"`
}

type Truck struct {
	ID            uint64    `json:"truck_id"`
	Brand         string    `json:"brand"`
	Model         string    `json:"model"`
	VehicleNumber string    `json:"vehicle_number"`
	Name          string    `json:"name"`
	Status        string    `json:"truck_status"`
	CreatedAt     time.Time `json:"-" bson:"created_at"`
	UpdatedAt     time.Time `json:"-" bson:"updated_at"`
	DeletedAt     time.Time `json:"-" bson:"deleted_at"`
}

type Mart struct {
	ID              uint64  `json:"mart_id"`
	Name            string  `json:"name"`
	AddressID       uint64  `json:"-" gorm:"column:address_id"`
	Address         Address `json:"mart_address" gorm:"foreignKey:AddressID;references:ID"`
	Status          string  `json:"mart_status"`
	PackagingCharge float64 `json:"packaging_charge"`
	DeliveryCharge  float64 `json:"delivery_charge"`
	RefreshTokenID  string  `json:"-" gorm:"column:refresh_token_id"`
}

type GroceryMisc struct {
	ID uint64 `json:"-" gorm:"column:id"`
}