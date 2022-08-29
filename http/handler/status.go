package handler

import (
	"context"
	"fmt"

	"github.com/modular-project/orders-service/model"
	pf "github.com/modular-project/protobuffers/order/order"
)

type OrderStatusServicer interface {
	PayDelivery(c context.Context, oID, uID, eID uint64, aID string, pm model.PaymentMethod) (string, error)
	PayLocal(oID, eID uint64, pm model.PaymentMethod) error
	CompleteProduct(uint64) error
	CapturePayment(context.Context, string) (string, error)
}

type OrderStatusUC struct {
	pf.UnimplementedOrderStatusServiceServer
	oss OrderStatusServicer
}

func NewOrderStatusUC(oss OrderStatusServicer) OrderStatusUC {
	return OrderStatusUC{oss: oss}
}

func (ouc OrderStatusUC) PayDelivery(c context.Context, r *pf.PayDeliveryRequest) (*pf.PayDeliveryResponse, error) {
	if r == nil {
		return nil, fmt.Errorf("nil request")
	}
	id, err := ouc.oss.PayDelivery(c, r.OrdeId, r.UserId, r.EstablishmentId, r.Address, model.PaymentMethod(r.Payment))
	if err != nil {
		return nil, fmt.Errorf("oss.PayDelivery: %w", err)
	}
	return &pf.PayDeliveryResponse{Id: id}, nil
}

func (ouc OrderStatusUC) PayLocal(c context.Context, r *pf.PayLocalRequest) (*pf.PayLocalResponse, error) {
	if r == nil {
		return nil, fmt.Errorf("nil request")
	}
	if err := ouc.oss.PayLocal(r.OrdeId, r.EmployeeId, model.PaymentMethod(r.Payment)); err != nil {
		return nil, fmt.Errorf("oss.PayDelivery: %w", err)
	}
	return &pf.PayLocalResponse{}, nil
}

func (ouc OrderStatusUC) CompleteProduct(c context.Context, r *pf.CompleteProductRequest) (*pf.CompleteProductResponse, error) {
	if r == nil {
		return nil, fmt.Errorf("nil request")
	}
	if err := ouc.oss.CompleteProduct(r.Id); err != nil {
		return nil, fmt.Errorf("ouc.CompleteProduct: %w", err)
	}
	return &pf.CompleteProductResponse{}, nil
}

func (ouc OrderStatusUC) CapturePayment(c context.Context, r *pf.CapturePaymentRequest) (*pf.CapturePaymentResponse, error) {
	if r == nil {
		return nil, fmt.Errorf("nil request")
	}
	st, err := ouc.oss.CapturePayment(c, r.Id)
	if err != nil {
		return nil, fmt.Errorf("oss.CapturePayment: %w", err)
	}
	return &pf.CapturePaymentResponse{Status: st}, nil
}
