package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	WithoutPay Status = iota + 1
	Pending
	Completed
)

const (
	Local Type = iota + 1
	Delivery
)

const (
	CASH PaymentMethod = iota + 1
	PAYPAL
)

type PaymentMethod uint32

type Status uint32

type Type uint32

type Model struct {
	ID        uint64         `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Product struct {
	Price float64
	Name  string
}

type Order struct {
	Model
	TypeID          Type
	UserID          uint64
	EmployeeID      uint64
	EstablishmentID uint64
	TableID         uint64
	AddressID       *string
	StatusID        Status
	Total           float64
	PayID           *string
	OrderProducts   []OrderProduct
}

type OrderProduct struct {
	ID        uint64 `gorm:"primarykey" json:"id"`
	OrderID   uint64
	ProductID uint64
	Quantity  uint32
	IsReady   bool
}
