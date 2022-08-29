package storage

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/modular-project/orders-service/model"
	"github.com/stretchr/testify/assert"
)

var TestConfigDB DBConnection = DBConnection{
	TypeDB:   POSTGRESQL,
	User:     "admin_restaurant",
	Password: "RestAuraNt_pgsql.561965697",
	Host:     "localhost",
	Port:     "5433",
	NameDB:   "testing",
}

func TestCleanup(t *testing.T) {
	err := NewDB(TestConfigDB)
	if err != nil {
		t.Fatalf("NewGormDB: %s", err)
	}
	models := []interface{}{&model.Order{}, &model.OrderProduct{}}
	err = Drop(models...)
	if err != nil {
		t.Fatalf("Failed to Create tables: %s", err)
	}
}

func generateData(t *testing.T) {
	orders := []model.Order{
		{
			TypeID:          model.Local,
			EmployeeID:      1,
			EstablishmentID: 1,
			TableID:         1,
			StatusID:        model.Pending,
			Total:           155055.34,
			OrderProducts: []model.OrderProduct{
				{ProductID: 1, Quantity: 3},
				{ProductID: 1, Quantity: 2},
				{ProductID: 2, Quantity: 7},
			},
		}, {
			TypeID:          model.Local,
			EmployeeID:      1,
			EstablishmentID: 1,
			TableID:         1,
			StatusID:        model.Completed,
			Total:           100,
			OrderProducts: []model.OrderProduct{
				{ProductID: 3, Quantity: 3, IsReady: true},
				{ProductID: 4, Quantity: 2, IsReady: true},
				{ProductID: 2, Quantity: 7},
			},
		}, {
			TypeID:          model.Local,
			EmployeeID:      1,
			EstablishmentID: 1,
			TableID:         2,
			StatusID:        model.Pending,
			Total:           200,
			OrderProducts: []model.OrderProduct{
				{ProductID: 3, Quantity: 3},
				{ProductID: 6, Quantity: 2},
				{ProductID: 1, Quantity: 2},
			},
		}, {
			TypeID:          model.Local,
			EmployeeID:      2,
			EstablishmentID: 2,
			TableID:         3,
			StatusID:        model.Pending,
			Total:           200,
			OrderProducts: []model.OrderProduct{
				{ProductID: 3, Quantity: 3},
				{ProductID: 6, Quantity: 2},
				{ProductID: 1, Quantity: 2},
			},
		},
	}
	if err := _db.CreateInBatches(&orders, len(orders)).Error; err != nil {
		t.Fatalf("failed to generate data: %s", err)
	}
}

func TestOrderStorage_AddProducts(t *testing.T) {
	if err := NewDB(TestConfigDB); err != nil {
		t.Fatalf("failed to start connection with db: %s", err)
	}
	models := []interface{}{
		model.Order{},
		model.OrderProduct{},
	}
	_db.AutoMigrate(models...)
	t.Cleanup(func() {
		err := _db.Migrator().DropTable(models...)
		if err != nil {
			t.Fatalf("Failed to Create tables: %s", err)
		}
	})
	os := NewOrderStorage()

	generateData(t)
	type args struct {
		oID uint64
		ps  []model.OrderProduct
	}
	tests := []struct {
		name    string
		os      OrderStorage
		args    args
		wantErr bool
	}{
		{
			name: "add ok",
			os:   os,
			args: args{
				oID: 1,
				ps: []model.OrderProduct{
					{ProductID: 1, Quantity: 3},
					{ProductID: 3, Quantity: 1},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.os.AddProducts(tt.args.oID, 100, tt.args.ps); (err != nil) != tt.wantErr {
				t.Errorf("OrderStorage.AddProducts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrderStorage_Waiter(t *testing.T) {
	if err := NewDB(TestConfigDB); err != nil {
		t.Fatalf("failed to start connection with db: %s", err)
	}
	models := []interface{}{
		model.Order{},
		model.OrderProduct{},
	}
	_db.AutoMigrate(models...)
	t.Cleanup(func() {
		err := _db.Migrator().DropTable(models...)
		if err != nil {
			t.Fatalf("Failed to Create tables: %s", err)
		}
	})
	os := NewOrderStorage()

	generateData(t)
	tests := []struct {
		name    string
		os      OrderStorage
		giveID  uint64
		want    []model.Order
		wantErr bool
	}{
		{
			name:   "waiter 1 - ok",
			os:     os,
			giveID: 1,
			want: []model.Order{
				{
					Model:   model.Model{ID: 1},
					TableID: 1,
					OrderProducts: []model.OrderProduct{
						{ProductID: 1, Quantity: 3},
						{ProductID: 1, Quantity: 2},
						{ProductID: 2, Quantity: 7},
					},
				}, {
					Model:   model.Model{ID: 3},
					TableID: 2,
					OrderProducts: []model.OrderProduct{
						{ProductID: 3, Quantity: 3},
						{ProductID: 6, Quantity: 2},
						{ProductID: 1, Quantity: 2},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.os.Waiter(tt.giveID)
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderStorage.Waiter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert := assert.New(t)
			for i, o := range tt.want {
				t.Logf("I - %d", i)
				assert.Equal(o.ID, got[i].ID, "ID")
				assert.Equal(o.TableID, got[i].TableID, "Table")
				assert.Equal(o.Total, got[i].Total, "total")
				for y, p := range o.OrderProducts {
					assert.Equal(p.ProductID, got[i].OrderProducts[y].ProductID, fmt.Sprintf("product - %d", y))
					assert.Equal(p.Quantity, got[i].OrderProducts[y].Quantity, fmt.Sprintf("quantity - %d", y))
				}
			}
		})
	}
}

func TestOrderStorage_Kitchen(t *testing.T) {
	if err := NewDB(TestConfigDB); err != nil {
		t.Fatalf("failed to start connection with db: %s", err)
	}
	models := []interface{}{
		model.Order{},
		model.OrderProduct{},
	}
	_db.AutoMigrate(models...)
	t.Cleanup(func() {
		err := _db.Migrator().DropTable(models...)
		if err != nil {
			t.Fatalf("Failed to Create tables: %s", err)
		}
	})
	os := NewOrderStorage()

	generateData(t)
	tests := []struct {
		name    string
		os      OrderStorage
		giveID  uint
		want    []model.OrderProduct
		wantErr bool
	}{
		{
			name:   "ok",
			os:     os,
			giveID: 1,
			want: []model.OrderProduct{
				{ID: 1, OrderID: 1, ProductID: 1, Quantity: 3},
				{ID: 2, OrderID: 1, ProductID: 1, Quantity: 2},
				{ID: 3, OrderID: 1, ProductID: 2, Quantity: 7},
				{ID: 7, OrderID: 3, ProductID: 3, Quantity: 3},
				{ID: 8, OrderID: 3, ProductID: 6, Quantity: 2},
				{ID: 9, OrderID: 3, ProductID: 1, Quantity: 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.os.Kitchen(uint64(tt.giveID))
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderStorage.Kitchen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OrderStorage.Kitchen() = %v, want %v", got, tt.want)
			}
		})
	}
}
