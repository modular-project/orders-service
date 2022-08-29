package storage

import (
	"fmt"
	"log"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DRIVER string

const (
	POSTGRESQL DRIVER = "POSTGRES"
	TESTING    DRIVER = "TESTING"
)

var (
	_db  *gorm.DB
	once sync.Once
)

type DBConnection struct {
	TypeDB   DRIVER
	User     string
	Password string
	Port     string
	NameDB   string
	Host     string
}

func NewDB(conn DBConnection) error {
	var err error
	once.Do(func() {
		switch conn.TypeDB {
		case POSTGRESQL:
			err = newPostgresDB(&conn)
		default:
			err = fmt.Errorf("invalid database type")
		}
	})
	return err
}
func Drop(tables ...interface{}) error {
	return _db.Migrator().DropTable(tables...)
}

func Migrate(tables ...interface{}) error {
	// err := _db.SetupJoinTable(&model.User{}, "Roles", &model.UserRole{})
	// if err != nil {
	// 	return fmt.Errorf("fail at setup join table :%w", err)
	// }
	err := _db.AutoMigrate(tables...)
	if err != nil {
		return err
	}
	return nil
}
func newPostgresDB(u *DBConnection) error {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		u.Host, u.User, u.Password, u.NameDB, u.Port)
	_db, err = gorm.Open(postgres.Open(dsn))
	if _db == nil {
		log.Fatalf("nil db at open db")
	}
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	log.Println("connected to postgres")
	return nil
}

// func New(driver DRIVER) {
// 	once.Do(func() {
// 		u := loadData()
// 		switch driver {
// 		case POSTGRESQL:
// 			newPostgresDB(&u)
// 		case TESTING:
// 			newTestingDB(&u)
// 		}
// 	})
// }

// func newTestingDB(u *dbUser) {
// 	var err error
// 	fmt.Print(u)
// 	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", u.User, u.Password, u.Host, u.Port, "testing")
// 	_db, err = gorm.Open(postgres.Open(dsn))
// 	if err != nil {
// 		log.Fatalf("no se pudo abrir la base de datos: %v", err)
// 	}

// 	fmt.Println("conectado a Testing")
// }

// func newPostgresDB(u *dbUser) {
// 	var err error
// 	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", u.User, u.Password, u.Host, u.Port, u.NameDB)
// 	_db, err = gorm.Open(postgres.Open(dsn))
// 	if err != nil {
// 		log.Fatalf("no se pudo abrir la base de datos: %v", err)
// 	}

// 	fmt.Println("conectado a postgres")
// }

// // DB return a unique instance of db
// func DB() *gorm.DB {
// 	return _db
// }

// func getEnv(env string) (string, error) {
// 	s, f := os.LookupEnv(env)
// 	if !f {
// 		return "", fmt.Errorf("environment variable (%s) not found", env)
// 	}
// 	return s, nil
// }

// func loadData() dbUser {
// 	typeDb, err := getEnv("RGE_TYPE")
// 	if err != nil {
// 		log.Fatalf(err.Error())
// 	}
// 	user, err := getEnv("RGE_USER")
// 	if err != nil {
// 		log.Fatalf(err.Error())
// 	}
// 	password, err := getEnv("RGE_PASSWORD")
// 	if err != nil {
// 		log.Fatalf(err.Error())
// 	}
// 	port, err := getEnv("RGE_PORT")
// 	if err != nil {
// 		log.Fatalf(err.Error())
// 	}
// 	name, err := getEnv("RGE_NAME_DB")
// 	if err != nil {
// 		log.Fatalf(err.Error())
// 	}
// 	host, err := getEnv("RGE_HOST")
// 	if err != nil {
// 		log.Fatalf(err.Error())
// 	}
// 	return dbUser{DRIVER(typeDb), user, password, port, name, host}
// }
