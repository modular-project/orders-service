package storage

import (
	"fmt"
	"log"

	"github.com/modular-project/orders-service/model"
	"gorm.io/gorm"
)

type OrderStorage struct {
	db *gorm.DB
}

func NewOrderStorage() OrderStorage {
	return OrderStorage{db: _db}
}

func (os OrderStorage) Complete(oID uint64) error {
	err := os.db.Model(&model.Order{Model: model.Model{ID: oID}}).Update("status_id", model.Completed).Error
	if err != nil {
		return fmt.Errorf("update status_id: %w", err)
	}
	return nil
}

func (os OrderStorage) Kitchen(eID, last uint64) ([]model.OrderProduct, error) {
	var ps []model.OrderProduct
	log.Println(last)
	tx := os.db.Model(&model.OrderProduct{}).Joins("LEFT JOIN orders as o ON o.id = order_products.order_id").
		Where("o.establishment_id = ? AND order_products.is_ready = false AND o.status_id <> ?", eID, model.WithoutPay)
	if last > 0 {
		tx.Where("order_products.id > ?", last)
	}
	err := tx.Order("order_products.id").Find(&ps).Error
	if err != nil {
		return nil, fmt.Errorf("find order products: %w", err)
	}
	return ps, nil
}

func (os OrderStorage) Search(s *model.SearchOrder) ([]model.Order, error) {
	var o []model.Order
	tx := os.db.Model(&o).Select("id, type_id, establishment_id, address_id, status_id, total, created_at")
	if s.Users != nil {
		tx.Where("user_id IN ?", s.Users)
	}
	if s.Status != nil {
		tx.Where("status_id IN ?", s.Status)
	}
	if s.Ests != nil {
		tx.Where("establishment_id IN ?", s.Ests)
	}
	if s.Types != nil {
		tx.Where("type_id IN ?", s.Types)
	}
	if s.Lower > 0 {
		tx.Where("total >= ?", s.Lower)
	}
	if s.Higher > 0 {
		tx.Where("total <= ?", s.Higher)
	}
	q := s.Query()
	if q != "" {
		tx = tx.Order(s.Query())
	}
	if s.Limit != 0 {
		tx = tx.Limit(s.Limit)
	}
	if s.Offset != 0 {
		tx = tx.Offset(s.Offset)
	}
	err := tx.Find(&o).Error
	if err != nil {
		return nil, fmt.Errorf("find orders: %w", err)
	}
	return o, nil
}

func (os OrderStorage) User(uID uint64, limit, offset int) ([]model.Order, error) {
	tx := os.db.Preload("OrderProducts", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "product_id", "quantity", "order_id")
	}).Select("id", "address_id", "total", "status_id", "user_id", "pay_id", "created_at").Where("user_id = ?", uID)
	if limit != 0 {
		tx = tx.Limit(limit)
	}
	if offset != 0 {
		tx = tx.Offset(offset)
	}
	var orders []model.Order
	res := tx.Find(&orders)
	if res.Error != nil {
		return nil, fmt.Errorf("find: %w", res.Error)
	}
	return orders, nil
}

func (os OrderStorage) Waiter(wID uint64) ([]model.Order, error) {
	var o []model.Order
	err := os.db.Preload("OrderProducts", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "product_id", "quantity", "order_id", "is_ready", "is_delivered")
	}).Select("id", "table_id", "total").Where("employee_id = ? AND status_id = ?", wID, model.Pending).Find(&o).Error
	if err != nil {
		return nil, fmt.Errorf("find order products: %w", err)
	}
	return o, nil
}

func (os OrderStorage) WaiterPending(wID uint64) ([]model.Order, error) {
	var o []model.Order
	err := os.db.Preload("OrderProducts", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_ready = true AND is_delivered = false").Select("id", "product_id", "quantity", "order_id", "is_ready", "is_delivered")
	}).Select("id", "table_id").Where("employee_id = ? AND status_id = ?", wID, model.Pending).Find(&o).Error
	if err != nil {
		return nil, fmt.Errorf("find order products: %w", err)
	}
	return o, nil
}

func (os OrderStorage) Create(o *model.Order) error {
	if o == nil {
		return fmt.Errorf("nil order")
	}
	if err := os.db.Create(o).Error; err != nil {
		return fmt.Errorf("create order: %w", err)
	}
	return nil
}

func (os OrderStorage) Products(oID uint64) ([]model.OrderProduct, error) {
	var ps []model.OrderProduct
	if err := os.db.Where("order_id = ?", oID).Find(&ps).Error; err != nil {
		return nil, fmt.Errorf("find all products by order: %w", err)
	}
	return ps, nil
}

func (os OrderStorage) updateTotal(oID uint64, total float64) error {
	o := model.Order{}
	if err := os.db.Where("id = ?", oID).Select("total").First(&o).Error; err != nil {
		return fmt.Errorf("first order: %w", err)
	}
	if err := os.db.Model(&model.Order{Model: model.Model{ID: oID}}).Update("total", o.Total+total).Error; err != nil {
		return fmt.Errorf("update total: %w", err)
	}
	return nil
}

func (os OrderStorage) AddProducts(oID uint64, total float64, ps []model.OrderProduct) error {
	if ps == nil {
		return fmt.Errorf("nil products")
	}
	if err := os.db.Model(&model.Order{Model: model.Model{ID: oID}}).Association("OrderProducts").Append(&ps); err != nil {
		return fmt.Errorf("append products to order: %w", err)
	}
	if err := os.updateTotal(oID, total); err != nil {
		return fmt.Errorf("os.updateTotal: %w", err)
	}
	return nil
}

func (os OrderStorage) Pay(oID uint64, total float64) error {
	if err := os.db.Model(&model.Order{Model: model.Model{ID: oID}}).Updates(&model.Order{Total: total, StatusID: model.Completed}).Error; err != nil {
		return fmt.Errorf("pay order in db: %w", err)
	}
	return nil
}

func (os OrderStorage) SetPayment(oID uint64, total float64, pID string) error {
	if err := os.db.Model(&model.Order{Model: model.Model{ID: oID}}).Updates(&model.Order{Total: total, StatusID: model.Completed, PayID: &pID}).Error; err != nil {
		return fmt.Errorf("set payment in db: %w", err)
	}
	return nil
}

func (os OrderStorage) PayPaypal(pID string) error {
	if err := os.db.Model(&model.Order{}).Where("pay_id = ?", pID).Update("status_id", model.Completed).Error; err != nil {
		return fmt.Errorf("pay order with paypal in db: %w", err)
	}
	return nil
}
