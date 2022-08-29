package controller

import (
	"fmt"

	"github.com/modular-project/orders-service/model"
)

type OrderStorager interface {
	Kitchen(uint64) ([]model.OrderProduct, error)
	Search(*model.SearchOrder) ([]model.Order, error)
	Waiter(uint64) ([]model.Order, error)
	Create(*model.Order) error
	Products(uint64) ([]model.OrderProduct, error)
	AddProducts(uint64, float64, []model.OrderProduct) error
	PayPaypal(string) error
	Pay(uint64, float64) error
	SetPayment(uint64, float64, string) error
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

func (os OrderService) Create(o *model.Order) (uint64, error) {
	if err := os.str.Create(o); err != nil {
		return 0, fmt.Errorf("create order: %w", err)
	}
	return o.ID, nil
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

func (os OrderService) Kitchen(kID uint64) ([]model.OrderProduct, error) {
	if kID == 0 {
		return nil, fmt.Errorf("kitchen not found")
	}
	ps, err := os.str.Kitchen(kID)
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
	s.Types = []model.Type{model.Delivery}
	s.Users = []uint64{uID}
	s.Ests = nil
	orders, err := os.str.Search(&s)
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

// func (os OrderService) Capture(pID string) error {
// 	status, err := os.ps.CaptureOrder(pID)
// 	if err != nil {
// 		return fmt.Errorf("ps.CaptureOrder: %w", err)
// 	}
// 	if status != "COMPLETED" {
// 		return fmt.Errorf("payment with status %s", status)
// 	}
// 	if err := os.str.PayPaypal(pID); err != nil {
// 		return fmt.Errorf("str.PayPaypal: %w", err)
// 	}
// 	return nil
// }

// func (os OrderService) getTotal(op []model.OrderProduct) (float64, error) {
// 	var total float64
// 	if op == nil {
// 		return 0, nil
// 	}
// 	for _, o := range op {
// 		p, err := os.pr.Price(o.ProductID)
// 		if err != nil {
// 			return 0, fmt.Errorf("pr.Price: %w", err)
// 		}
// 		total += float64(o.Quantity) * p
// 	}
// 	return total, nil
// }

// func (os OrderService) Pay(oID uint64, oType model.Type, payMeth model.PaymentMethod) (string, error) {
// 	if oType == model.Delivery && payMeth == model.CASH {
// 		return "", fmt.Errorf("delivery order must be pay with paypal")
// 	}
// 	if oType == model.Local && payMeth == model.PAYPAL {
// 		return "", fmt.Errorf("local order must be pay with cash")
// 	}
// 	switch payMeth {
// 	case model.PAYPAL:
// 		o, err := os.str.Products(oID)
// 		if err != nil {
// 			return "", fmt.Errorf("str.Products: %w", err)
// 		}
// 		id, total, err := os.ps.CreateOrder(o)
// 		if err != nil {
// 			return "", fmt.Errorf("ps.CreateOrder: %w", err)
// 		}
// 		if err := os.str.SetPayment(oID, total, id); err != nil {
// 			return "", fmt.Errorf("str.SetPayment: %w", err)
// 		}
// 		return id, nil
// 	case model.CASH:
// 		op, err := os.str.Products(oID)
// 		if err != nil {
// 			return "", fmt.Errorf("str.Products: %w", err)
// 		}
// 		t, err := os.getTotal(op)
// 		if err != nil {
// 			return "", fmt.Errorf("os.getTotal: %w", err)
// 		}
// 		if err := os.str.Pay(oID, t); err != nil {
// 			return "", fmt.Errorf("str.Pay: %w", err)
// 		}
// 		return "", nil
// 	}
// 	return "", fmt.Errorf("invalid payment method")
// }
