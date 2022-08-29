package handler

import (
	"context"
	"fmt"

	"github.com/modular-project/orders-service/model"
	pf "github.com/modular-project/protobuffers/order/order"
)

type OrderServicer interface {
	Products(oID uint64) ([]model.OrderProduct, error)
	Create(o *model.Order) (uint64, error)
	AddProducts(oID uint64, total float64, ps []model.OrderProduct) ([]uint64, error)
	Kitchen(kID uint64) ([]model.OrderProduct, error)
	Waiter(wID uint64) ([]model.Order, error)
	Search(s *model.SearchOrder) ([]model.Order, error)
	User(uID uint64, s model.SearchOrder) ([]model.Order, error)
	Establishment(uID uint64, s model.SearchOrder) ([]model.Order, error)
	// Pay(uint64, model.Type, model.PaymentMethod) (string, error)
	// Capture(string) error
}

type OrderUC struct {
	os OrderServicer
	pf.UnimplementedOrderServiceServer
}

func NewOrderUC(os OrderServicer) OrderUC {
	return OrderUC{os: os}
}

func (ouc OrderUC) CreateLocalOrder(c context.Context, o *pf.Order) (*pf.ID, error) {
	lo := o.GetLocalOrder()
	if lo == nil {
		return nil, fmt.Errorf("local order is nil")
	}
	mo := model.Order{
		TypeID:          model.Local,
		EmployeeID:      lo.EmployeeId,
		EstablishmentID: o.EstablishmentId,
		TableID:         lo.TableId,
		StatusID:        model.Pending,
		Total:           float64(o.Total),
	}
	if o.OrderProducts != nil {
		mo.OrderProducts = make([]model.OrderProduct, len(o.OrderProducts))
	}
	for i := range o.OrderProducts {
		mo.OrderProducts[i] = model.OrderProduct{
			ProductID: o.OrderProducts[i].ProductId,
			Quantity:  o.OrderProducts[i].Quantity,
		}
	}
	oID, err := ouc.os.Create(&mo)
	if err != nil {
		return nil, fmt.Errorf("os.create: %w", err)
	}
	return &pf.ID{Id: oID}, nil
}

func (ouc OrderUC) CreateDeliveryOrder(c context.Context, o *pf.Order) (*pf.ID, error) {
	do := o.GetRemoteOrder()
	if do == nil {
		return nil, fmt.Errorf("delivery order is nil")
	}
	mo := model.Order{
		UserID:        do.UserId,
		TypeID:        model.Delivery,
		StatusID:      model.WithoutPay,
		OrderProducts: make([]model.OrderProduct, len(o.OrderProducts)),
		Total:         float64(o.Total),
	}
	if o.OrderProducts == nil {
		return nil, fmt.Errorf("without products")
	}
	for i := range o.OrderProducts {
		mo.OrderProducts[i] = model.OrderProduct{
			ProductID: o.OrderProducts[i].ProductId,
			Quantity:  o.OrderProducts[i].Quantity,
		}
	}
	oID, err := ouc.os.Create(&mo)
	if err != nil {
		return nil, fmt.Errorf("os.create: %w", err)
	}
	return &pf.ID{Id: oID}, nil
}

func (ouc OrderUC) GetOrdersByUser(c context.Context, r *pf.OrdersByUserRequest) (*pf.OrdersResponse, error) {
	if r == nil {
		return nil, fmt.Errorf("nil request")
	}
	if r.Search.Users == nil {
		return nil, fmt.Errorf("nil user")
	}
	o, err := ouc.os.User(r.Search.Users[0], newSearch(r.Search))
	if err != nil {
		return nil, fmt.Errorf("os.User: %w", err)
	}
	if o == nil {
		return nil, nil
	}
	return &pf.OrdersResponse{Orders: protoOrder(o)}, nil
}

func (ouc OrderUC) GetOrdersByKitchen(c context.Context, id *pf.ID) (*pf.OrderProductsResponse, error) {
	ops, err := ouc.os.Kitchen(id.Id)
	if err != nil {
		return nil, fmt.Errorf("os.Kitchen: %w", err)
	}
	if ops == nil {
		return nil, nil
	}
	po := make([]*pf.OrderProduct, len(ops))
	for i := range ops {
		po[i].Id = ops[i].ID
		po[i].IsReady = ops[i].IsReady
		po[i].ProductId = ops[i].ProductID
		po[i].Quantity = ops[i].Quantity
	}
	return &pf.OrderProductsResponse{OrderProducts: po}, nil
}

func (ouc OrderUC) GetOrders(c context.Context, r *pf.OrdersRequest) (*pf.OrdersResponse, error) {
	if r == nil {
		return nil, fmt.Errorf("nil request")
	}
	s := newSearch(r.Search)
	os, err := ouc.os.Search(&s)
	if err != nil {
		return nil, fmt.Errorf("os.Search(): %w", err)
	}
	if os == nil {
		return nil, nil
	}
	return &pf.OrdersResponse{Orders: protoOrder(os)}, nil
}

func (ouc OrderUC) GetOrdersByEstablishment(c context.Context, r *pf.OrdersRequest) (*pf.OrdersResponse, error) {
	if r == nil {
		return nil, fmt.Errorf("nil request")
	}
	if r.Search.Establishments == nil {
		return nil, fmt.Errorf("nil establishment")
	}
	os, err := ouc.os.Establishment(r.Search.Establishments[0], newSearch(r.Search))
	if err != nil {
		return nil, fmt.Errorf("os.Establishment: %w", err)
	}
	if os == nil {
		return nil, nil
	}
	return &pf.OrdersResponse{Orders: protoOrder(os)}, nil
}

func (ouc OrderUC) GetOrderByWaiter(c context.Context, id *pf.ID) (*pf.OrdersResponse, error) {
	os, err := ouc.os.Waiter(id.Id)
	if err != nil {
		return nil, fmt.Errorf("os.Waiter: %w", err)
	}
	if os == nil {
		return nil, nil
	}
	return &pf.OrdersResponse{Orders: protoOrder(os)}, nil
}

func (ouc OrderUC) AddProductsToOrder(c context.Context, r *pf.AddProductsToOrderRequest) (*pf.AddProductsToOrderResponse, error) {
	if r == nil {
		return nil, fmt.Errorf("nil request")
	}
	ids, err := ouc.os.AddProducts(r.Id, float64(r.Total), orderProducts(r.Products))
	if err != nil {
		return nil, fmt.Errorf("os.AddProducts: %w", err)
	}
	return &pf.AddProductsToOrderResponse{Ids: ids}, nil
}

// func (ouc OrderUC) Capture(c context.Context, r *pf.CapturePaymentRequest) (*pf.CapturePaymentResponse, error) {
// 	if r == nil {
// 		return nil, fmt.Errorf("nil request")
// 	}
// 	if err := ouc.os.Capture(r.Id); err != nil {
// 		return nil, fmt.Errorf("os.Capture: %w", err)
// 	}
// 	return &pf.CapturePaymentResponse{}, nil
// }

func newOrderBy(s []*pf.SearchBy) []model.OrderBy {
	if s == nil {
		return nil
	}
	o := make([]model.OrderBy, len(s))
	for i := range s {
		o[i].By = model.By(s[i].By)
		o[i].Sort = model.Sort(s[i].Sort)
	}
	return o
}

func newSearch(ps *pf.SearchOrders) model.SearchOrder {
	if ps == nil {
		return model.SearchOrder{}
	}
	s := model.SearchOrder{
		Ests:  ps.Establishments,
		Users: ps.Users,
	}
	if ps.Default != nil {
		s.Search = model.Search{
			Limit:    int(ps.Default.Limit),
			Offset:   int(ps.Default.Offset),
			OrderBys: newOrderBy(ps.Default.SearchBy),
		}
	}
	if ps.Types != nil {
		t := make([]model.Type, len(ps.Types))
		for i := range ps.Types {
			t[i] = model.Type(ps.Types[i])
		}
		s.Types = t
	}
	if ps.Status != nil {
		st := make([]model.Status, len(ps.Status))
		for i := range ps.Status {
			st[i] = model.Status(ps.Status[i])
		}
		s.Status = st
	}
	if ps.Range != nil {
		s.Lower = float64(ps.Range[0])
		s.Higher = float64(ps.Range[1])
	}
	return s
}

func orderProducts(pop []*pf.OrderProduct) []model.OrderProduct {
	if pop == nil {
		return nil
	}
	op := make([]model.OrderProduct, len(pop))
	for i := range pop {
		op[i] = model.OrderProduct{
			ProductID: pop[i].ProductId,
			Quantity:  pop[i].Quantity,
		}
	}
	return op
}

func protoOrder(os []model.Order) []*pf.Order {
	if os == nil {
		return nil
	}
	po := make([]*pf.Order, len(os))
	for i := range os {
		t := os[i]
		po[i] = &pf.Order{
			Id:              t.ID,
			EstablishmentId: t.EstablishmentID,
			Total:           float32(t.Total),
			Status:          pf.Status(t.StatusID),
			OrderProducts:   make([]*pf.OrderProduct, len(t.OrderProducts)),
		}
		for y := range t.OrderProducts {
			po[i].OrderProducts[y] = &pf.OrderProduct{
				Id:        t.OrderProducts[y].ID,
				ProductId: t.OrderProducts[y].ProductID,
				Quantity:  t.OrderProducts[y].Quantity,
				IsReady:   t.OrderProducts[y].IsReady,
			}
		}
	}
	return po
}
