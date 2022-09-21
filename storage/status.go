package storage

import (
	"fmt"
	"log"

	"github.com/modular-project/orders-service/model"
	"gorm.io/gorm"
)

type orderStatusStorage struct {
	db *gorm.DB
}

func NewOrderStatusStorage() orderStatusStorage {
	return orderStatusStorage{db: _db}
}
func (os orderStatusStorage) TotalPrice(oID uint64, uID uint64) (float64, error) {
	o := model.Order{}
	log.Print("oRDER ID: ", oID, "User ID: ", uID)
	if err := os.db.Where("id = ? AND user_id = ?", oID, uID).Select("total").First(&o).Error; err != nil {
		return 0, fmt.Errorf("first order: %w", err)
	}
	return o.Total, nil
}

func (os orderStatusStorage) SetPaymentDelivery(oID uint64, eID uint64, pID string, aID string) error {
	o := model.Order{
		EstablishmentID: eID,
		PayID:           &pID,
		AddressID:       &aID,
	}
	if err := os.db.Model(&model.Order{Model: model.Model{ID: oID}}).Updates(&o).Error; err != nil {
		return fmt.Errorf("update order: %w", err)
	}
	return nil
}

func (os orderStatusStorage) PayLocal(oID uint64, eID uint64) error {
	err := os.db.Model(&model.Order{}).Where("id = ? AND employee_id = ?", oID, eID).Update("status_id", model.Completed).Error
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}

func (os orderStatusStorage) PayDelivey(pID string) error {
	err := os.db.Model(&model.Order{}).Where("pay_id = ?", pID).Update("status_id", model.Completed).Error
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}

func (os orderStatusStorage) CompleteProduct(pID uint64) error {
	err := os.db.Model(&model.OrderProduct{}).Where("id = ?", pID).Update("is_ready", true).Error
	if err != nil {
		return fmt.Errorf("update order product status: %w", err)
	}
	return nil
}

func (os orderStatusStorage) DeliverProduct(ids []uint64) error {
	err := os.db.Table("order_products").Where("id IN ?", ids).Updates(&model.OrderProduct{IsDelivered: true}).Error
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	return nil
}
