package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/modular-project/orders-service/model"
)

type PaypalServicer interface {
	CreateOrder(context.Context, float64) (string, error)
	CaptureOrder(context.Context, string) (string, error)
}

type OrderStatusStorager interface {
	TotalPrice(oID, uID uint64) (float64, error)
	SetPaymentDelivery(oID, eID uint64, pID, aID string) error
	PayLocal(oID, eID uint64) error
	PayDelivey(string) error
	CompleteProduct(uint64) error
	DeliverProduct([]uint64) error
	CancelOrders([]uint64, uint64) error
}

type OrderStatusService struct {
	ost OrderStatusStorager
	ps  PaypalServicer
}

func NewOrderStatusService(ost OrderStatusStorager, ps PaypalServicer) OrderStatusService {
	return OrderStatusService{ost: ost, ps: ps}
}

func (oss OrderStatusService) CancelOrders(ids []uint64, uID uint64) error {
	if err := oss.ost.CancelOrders(ids, uID); err != nil {
		return fmt.Errorf("controller CancelOrders: %w", err)
	}
	return nil
}

func (oss OrderStatusService) DeliverProduct(ids []uint64) error {
	if err := oss.ost.DeliverProduct(ids); err != nil {
		return fmt.Errorf("ost.DeliverProduct: %w", err)
	}
	return nil
}

func (oss OrderStatusService) PayDelivery(c context.Context, oID uint64, uID uint64, eID uint64, aID string, pm model.PaymentMethod) (string, error) {
	if pm != model.PAYPAL {
		return "", fmt.Errorf("payment method must be paypal")
	}
	tp, err := oss.ost.TotalPrice(oID, uID)
	if err != nil {
		return "", fmt.Errorf("ost.TotalPrice: %w", err)
	}
	pID, err := oss.ps.CreateOrder(c, tp)
	if err != nil {
		return "", fmt.Errorf("ps.CreateOrder: %w", err)
	}
	if err := oss.ost.SetPaymentDelivery(oID, eID, pID, aID); err != nil {
		return "", fmt.Errorf("ost.PayDelivery: %w", err)
	}
	return pID, nil

}

func (oss OrderStatusService) PayLocal(oID uint64, eID uint64, pm model.PaymentMethod) error {
	if pm != model.CASH {
		return fmt.Errorf("payment method must be cash")
	}
	if err := oss.ost.PayLocal(oID, eID); err != nil {
		return fmt.Errorf("ost.PayLocal: %w", err)
	}
	return nil
}

func (oss OrderStatusService) CompleteProduct(opID uint64) error {
	if err := oss.ost.CompleteProduct(opID); err != nil {
		return fmt.Errorf("ost.CompleteProduct: %w", err)
	}
	return nil
}

func (oss OrderStatusService) CapturePayment(c context.Context, pID string) (string, error) {
	s, err := oss.ps.CaptureOrder(c, pID)
	if err != nil {
		return "", fmt.Errorf("ps.CaptureOrder: %w", err)
	}
	if !strings.EqualFold(s, "COMPLETED") {
		return s, fmt.Errorf("payment status is not completed")
	}
	if err := oss.ost.PayDelivey(pID); err != nil {
		return s, fmt.Errorf("ost.PayDelivery: %w", err)
	}
	return s, nil
}
