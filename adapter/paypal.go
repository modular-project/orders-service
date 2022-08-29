package adapter

import (
	"context"
	"fmt"

	"github.com/plutov/paypal/v4"
)

type paypalService struct {
	c      *paypal.Client
	appCtx paypal.ApplicationContext
}

func NewPaypalSerive(cltID, secret, api, sURL, bName string) (paypalService, error) {
	c, err := paypal.NewClient(cltID, secret, api)
	if err != nil {
		return paypalService{}, fmt.Errorf("paypal.NewClient: %w", err)
	}
	ps := paypalService{
		c: c,
		appCtx: paypal.ApplicationContext{
			BrandName: bName,
			ReturnURL: fmt.Sprintf("%s/order/return", sURL),
			CancelURL: fmt.Sprintf("%s/order/cancel", sURL),
		},
	}
	return ps, nil
}

func (ps paypalService) CreateOrder(ctx context.Context, t float64) (string, error) {
	pur := paypal.PurchaseUnitRequest{
		ReferenceID: "ref-id",
		Amount:      &paypal.PurchaseUnitAmount{Currency: "MXN", Value: fmt.Sprint(t)},
	}
	po, err := ps.c.CreateOrder(ctx, paypal.OrderIntentCapture, []paypal.PurchaseUnitRequest{pur}, nil, &ps.appCtx)
	if err != nil {
		return "", fmt.Errorf("c.CreateOrder: %w", err)
	}
	return po.ID, nil
}

// func (ps paypalService) CreateOrder(ctx context.Context, t float64) (string, error) {
// 	var total float64
// 	if o == nil {
// 		return "", fmt.Errorf("nil order")
// 	}
// 	items := make([]paypal.Item, len(o))
// 	pur := paypal.PurchaseUnitRequest{Amount: &paypal.PurchaseUnitAmount{Currency: "MXN"}}
// 	for i := range o {
// 		// Get product
// 		p, err := ps.pr.Product(o[i].ProductID)
// 		if err != nil {
// 			return "", fmt.Errorf("product: %w", err)
// 		}
// 		// Add price to total
// 		total += p.Price * float64(o[i].Quantity)
// 		// Set product to item list
// 		items[i].Name = p.Name
// 		items[i].Quantity = fmt.Sprint(o[i].Quantity)
// 		items[i].UnitAmount = &paypal.Money{Currency: "MXN", Value: fmt.Sprint(p.Price)}
// 	}
// 	pur.Items = items
// 	pur.Amount.Value = fmt.Sprint(total)
// 	po, err := ps.c.CreateOrder(ctx, paypal.OrderIntentCapture, []paypal.PurchaseUnitRequest{pur}, nil, &ps.appCtx)
// 	if err != nil {
// 		return "", fmt.Errorf("c.CreateOrder: %w", err)
// 	}
// 	return po.ID, nil
// }

func (ps paypalService) CaptureOrder(ctx context.Context, id string) (string, error) {
	r, err := ps.c.CaptureOrder(ctx, id, paypal.CaptureOrderRequest{})
	if err != nil {
		return "", fmt.Errorf("c.CaptureOrder: %w", err)
	}
	return r.Status, nil
}
