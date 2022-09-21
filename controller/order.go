package controller

import (
	"fmt"

	"github.com/modular-project/orders-service/model"
)

type OrderStorager interface {
	Kitchen(kID, last uint64) ([]model.OrderProduct, error)
	Search(*model.SearchOrder) ([]model.Order, error)
	Waiter(uint64) ([]model.Order, error)
	WaiterPending(uint64) ([]model.Order, error)
	Create(*model.Order) error
	Products(uint64) ([]model.OrderProduct, error)
	AddProducts(uint64, float64, []model.OrderProduct) error
	PayPaypal(string) error
	Pay(uint64, float64) error
	SetPayment(uint64, float64, string) error
	User(uID uint64, limit, offset int) ([]model.Order, error)
}

type OrderService struct {
	str OrderStorager
}

func NewOrderService(str OrderStorager) OrderService {
	return OrderService{str: str}
}

func (os OrderService) Products(oID uint64) ([]model.OrderProduct, error) {
	if oID == 0 {
		return nil, fmt.Errorf("order not found")
	}
	ps, err := os.str.Products(oID)
	if err != nil {
		return nil, fmt.Errorf("products by order: %w", err)
	}
	return ps, nil
}

func (os OrderService) Create(o *model.Order) ([]uint64, error) {
	if err := os.str.Create(o); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	ids := make([]uint64, len(o.OrderProducts))
	for i := range o.OrderProducts {
		ids[i] = o.OrderProducts[i].ID
	}
	return ids, nil
}

func (os OrderService) AddProducts(oID uint64, total float64, ps []model.OrderProduct) ([]uint64, error) {
	if oID == 0 {
		return nil, fmt.Errorf("order not found")
	}
	if ps == nil {
		return nil, fmt.Errorf("products are nil")
	}
	if err := os.str.AddProducts(oID, total, ps); err != nil {
		return nil, fmt.Errorf("create order products: %w", err)
	}
	ids := make([]uint64, len(ps))
	for i := range ps {
		ids[i] = ps[i].ID
	}
	return ids, nil
}

func (os OrderService) Kitchen(kID, last uint64) ([]model.OrderProduct, error) {
	if kID == 0 {
		return nil, fmt.Errorf("kitchen not found")
	}
	ps, err := os.str.Kitchen(kID, last)
	if err != nil {
		return nil, fmt.Errorf("get by kitchen: %w", err)
	}
	return ps, nil
}

func (os OrderService) Waiter(wID uint64) ([]model.Order, error) {
	if wID == 0 {
		return nil, fmt.Errorf("user not found")
	}
	orders, err := os.str.Waiter(wID)
	if err != nil {
		return nil, fmt.Errorf("get by waiter: %w", err)
	}
	return orders, nil
}

func (os OrderService) WaiterPending(wID uint64) ([]model.Order, error) {
	if wID == 0 {
		return nil, fmt.Errorf("user not found")
	}
	orders, err := os.str.WaiterPending(wID)
	if err != nil {
		return nil, fmt.Errorf("get by waiter: %w", err)
	}
	return orders, nil
}

func (os OrderService) Search(s *model.SearchOrder) ([]model.Order, error) {
	orders, err := os.str.Search(s)
	if err != nil {
		return nil, fmt.Errorf("get by user: %w", err)
	}
	return orders, nil
}

func (os OrderService) User(uID uint64, s model.SearchOrder) ([]model.Order, error) {
	if uID == 0 {
		return nil, fmt.Errorf("user not found")
	}
	// s.Types = []model.Type{model.Delivery}
	// s.Users = []uint64{uID}
	// s.Ests = nil
	orders, err := os.str.User(uID, s.Limit, s.Offset)
	if err != nil {
		return nil, fmt.Errorf("get by user: %w", err)
	}
	return orders, nil
}

func (os OrderService) Establishment(uID uint64, s model.SearchOrder) ([]model.Order, error) {
	if uID == 0 {
		return nil, fmt.Errorf("user not found")
	}
	s.Users = nil
	s.Ests = []uint64{uID}
	orders, err := os.str.Search(&s)
	if err != nil {
		return nil, fmt.Errorf("get by user: %w", err)
	}
	return orders, nil
}
