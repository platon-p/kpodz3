package domain

type OrderStatus string

const (
	OrderStatusCreated OrderStatus = "created"
	OrderStatusSuccess OrderStatus = "success"
	OrderStatusFailed  OrderStatus = "failed"
)

type Order struct {
	UserId int
	Name   string
	Amount int
	Status OrderStatus
}
