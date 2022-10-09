package adapter

import (
	"context"
	"fmt"
	"log"
	"os"

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
	c.SetLog(os.Stdout) // Set log to terminal stdout
	c.GetAccessToken(context.Background())
	// log.Println(c.Token)
	if sURL == "localhost" {
		sURL = `https://example.com`
	}
	ps := paypalService{
		c: c,
		appCtx: paypal.ApplicationContext{
			BrandName: bName,
			ReturnURL: sURL,
			CancelURL: fmt.Sprintf("%s/cancel.html", sURL),
			// ReturnURL: fmt.Sprintf("https://example.com/return"),
			// CancelURL: fmt.Sprintf("https://example.com/cancel"),
		},
	}
	return ps, nil
}

func (ps paypalService) CreateOrder(ctx context.Context, t float64) (string, error) {
	pur := paypal.PurchaseUnitRequest{
		ReferenceID: "ref-id",
		Amount:      &paypal.PurchaseUnitAmount{Currency: "MXN", Value: fmt.Sprintf("%.2f", t)},
	}
	po, err := ps.c.CreateOrder(ctx, paypal.OrderIntentCapture, []paypal.PurchaseUnitRequest{pur}, nil, &ps.appCtx)
	if err != nil {
		return "", fmt.Errorf("c.CreateOrder: %w", err)
	}
	return po.ID, nil
}

func (ps paypalService) CaptureOrder(ctx context.Context, id string) (string, error) {
	log.Println(id)
	r, err := ps.c.CaptureOrder(ctx, id, paypal.CaptureOrderRequest{})
	if err != nil {
		return "", fmt.Errorf("c.CaptureOrder: %w", err)
	}
	return r.Status, nil
}
