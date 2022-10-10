package handler

import (
	"context"
	"fmt"

	"github.com/modular-project/orders-service/model"
	pf "github.com/modular-project/protobuffers/order/order"
)

type OrderServicer interface {
	Products(oID uint64) ([]model.OrderProduct, error)
	Create(*model.Order) ([]uint64, error)
	AddProducts(oID uint64, total float64, ps []model.OrderProduct) ([]uint64, error)
	Kitchen(kID, last uint64) ([]model.OrderProduct, error)
	Waiter(wID uint64) ([]model.Order, error)
	WaiterPending(wID uint64) ([]model.Order, error)
	Search(s *model.SearchOrder) ([]model.Order, error)
	User(uID uint64, s model.SearchOrder) ([]model.Order, error)
	Establishment(uID uint64, s model.SearchOrder) ([]model.Order, error)
}

type OrderUC struct {
	os OrderServicer
	pf.UnimplementedOrderServiceServer
}

func NewOrderUC(os OrderServicer) OrderUC {
	return OrderUC{os: os}
}

func (ouc OrderUC) CreateLocalOrder(c context.Context, o *pf.Order) (*pf.CreateResponse, error) {
	lo := o.GetLocalOrder()
	if lo == nil {
		return &pf.CreateResponse{}, fmt.Errorf("local order is nil")
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
	ids, err := ouc.os.Create(&mo)
	if err != nil {
		return &pf.CreateResponse{}, fmt.Errorf("os.create: %w", err)
	}
	return &pf.CreateResponse{OrderId: mo.ID, ProductIds: ids}, nil
}

func (ouc OrderUC) CreateDeliveryOrder(c context.Context, o *pf.Order) (*pf.CreateResponse, error) {
	do := o.GetRemoteOrder()
	if do == nil {
		return &pf.CreateResponse{}, fmt.Errorf("delivery order is nil")
	}
	mo := model.Order{
		UserID:        do.UserId,
		TypeID:        model.Delivery,
		StatusID:      model.WithoutPay,
		OrderProducts: make([]model.OrderProduct, len(o.OrderProducts)),
		AddressID:     &do.AddressId,
		Total:         float64(o.Total),
	}
	if o.OrderProducts == nil {
		return &pf.CreateResponse{}, fmt.Errorf("without products")
	}
	for i := range o.OrderProducts {
		mo.OrderProducts[i] = model.OrderProduct{
			ProductID: o.OrderProducts[i].ProductId,
			Quantity:  o.OrderProducts[i].Quantity,
		}
	}
	ids, err := ouc.os.Create(&mo)
	if err != nil {
		return &pf.CreateResponse{}, fmt.Errorf("os.create: %w", err)
	}
	return &pf.CreateResponse{OrderId: mo.ID, ProductIds: ids}, nil
}

func (ouc OrderUC) GetOrdersByUser(c context.Context, r *pf.OrdersByUserRequest) (*pf.OrdersResponse, error) {
	if r == nil {
		return &pf.OrdersResponse{}, fmt.Errorf("nil request")
	}
	if r.Search.Users == nil {
		return &pf.OrdersResponse{}, fmt.Errorf("nil user")
	}
	o, err := ouc.os.User(r.Search.Users[0], newSearch(r.Search))
	if err != nil {
		return &pf.OrdersResponse{}, fmt.Errorf("os.User: %w", err)
	}
	if o == nil {
		return &pf.OrdersResponse{}, nil
	}
	return &pf.OrdersResponse{Orders: protoOrder(o)}, nil
}

func (ouc OrderUC) GetOrderByID(c context.Context, r *pf.GetOrderByIDRequest) (*pf.OrderResponse, error) {
	if r == nil {
		return &pf.OrderResponse{}, fmt.Errorf("nil request")
	}
	ps, err := ouc.os.Products(r.OrderId)
	if err != nil {
		return &pf.OrderResponse{}, fmt.Errorf("os.User: %w", err)
	}
	if ps == nil {
		return &pf.OrderResponse{}, nil
	}
	o := []model.Order{
		{
			OrderProducts: ps,
		},
	}
	return &pf.OrderResponse{Order: protoOrder(o)[0]}, nil
}

func (ouc OrderUC) GetOrdersByKitchen(c context.Context, r *pf.RequestKitchen) (*pf.OrderProductsResponse, error) {
	ops, err := ouc.os.Kitchen(r.Id, r.Last)
	if err != nil {
		return &pf.OrderProductsResponse{}, fmt.Errorf("os.Kitchen: %w", err)
	}
	if ops == nil {
		return &pf.OrderProductsResponse{}, nil
	}
	po := make([]*pf.OrderProduct, len(ops))
	for i := range ops {
		po[i] = &pf.OrderProduct{
			Id:          ops[i].ID,
			IsReady:     ops[i].IsReady,
			ProductId:   ops[i].ProductID,
			Quantity:    ops[i].Quantity,
			IsDelivered: ops[i].IsDelivered,
		}
	}
	return &pf.OrderProductsResponse{OrderProducts: po}, nil
}

func (ouc OrderUC) GetOrders(c context.Context, r *pf.OrdersRequest) (*pf.OrdersResponse, error) {
	if r == nil {
		return &pf.OrdersResponse{}, fmt.Errorf("nil request")
	}
	s := newSearch(r.Search)
	os, err := ouc.os.Search(&s)
	if err != nil {
		return &pf.OrdersResponse{}, fmt.Errorf("os.Search(): %w", err)
	}
	if os == nil {
		return &pf.OrdersResponse{}, nil
	}
	return &pf.OrdersResponse{Orders: protoOrder(os)}, nil
}

func (ouc OrderUC) GetOrdersByEstablishment(c context.Context, r *pf.OrdersRequest) (*pf.OrdersResponse, error) {
	if r == nil {
		return &pf.OrdersResponse{}, fmt.Errorf("nil request")
	}
	if r.Search.Establishments == nil {
		return &pf.OrdersResponse{}, fmt.Errorf("nil establishment")
	}
	os, err := ouc.os.Establishment(r.Search.Establishments[0], newSearch(r.Search))
	if err != nil {
		return &pf.OrdersResponse{}, fmt.Errorf("os.Establishment: %w", err)
	}
	if os == nil {
		return &pf.OrdersResponse{}, nil
	}
	return &pf.OrdersResponse{Orders: protoOrder(os)}, nil
}

func (ouc OrderUC) GetOrderByWaiter(c context.Context, id *pf.ID) (*pf.OrdersResponse, error) {
	os, err := ouc.os.Waiter(id.Id)
	if err != nil {
		return &pf.OrdersResponse{}, fmt.Errorf("os.Waiter: %w", err)
	}
	if os == nil {
		return &pf.OrdersResponse{}, nil
	}
	return &pf.OrdersResponse{Orders: protoOrder(os)}, nil
}

func (ouc OrderUC) GetOrderPendingByWaiter(c context.Context, id *pf.ID) (*pf.OrdersResponse, error) {
	os, err := ouc.os.WaiterPending(id.Id)
	if err != nil {
		return &pf.OrdersResponse{}, fmt.Errorf("os.WaiterPending: %w", err)
	}
	if os == nil {
		return &pf.OrdersResponse{}, nil
	}
	return &pf.OrdersResponse{Orders: protoOrder(os)}, nil
}

func (ouc OrderUC) AddProductsToOrder(c context.Context, r *pf.AddProductsToOrderRequest) (*pf.AddProductsToOrderResponse, error) {
	if r == nil {
		return &pf.AddProductsToOrderResponse{}, fmt.Errorf("nil request")
	}
	ids, err := ouc.os.AddProducts(r.Id, float64(r.Total), orderProducts(r.Products))
	if err != nil {
		return &pf.AddProductsToOrderResponse{}, fmt.Errorf("os.AddProducts: %w", err)
	}
	return &pf.AddProductsToOrderResponse{Ids: ids}, nil
}

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
			CreateAt:        uint64(t.CreatedAt.Unix()),
		}
		if t.UserID != 0 {
			po[i].Type = &pf.Order_RemoteOrder{RemoteOrder: &pf.RemoteOrder{UserId: t.UserID, AddressId: *t.AddressID}}
		} else {
			po[i].Type = &pf.Order_LocalOrder{LocalOrder: &pf.LocalOrder{EmployeeId: t.EmployeeID, TableId: t.TableID}}
		}
		for y := range t.OrderProducts {
			po[i].OrderProducts[y] = &pf.OrderProduct{
				Id:          t.OrderProducts[y].ID,
				ProductId:   t.OrderProducts[y].ProductID,
				Quantity:    t.OrderProducts[y].Quantity,
				IsReady:     t.OrderProducts[y].IsReady,
				IsDelivered: t.OrderProducts[y].IsDelivered,
			}
		}
	}
	return po
}
