package main

import (
	"fmt"
	"log"
	"net"
	"os"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/modular-project/orders-service/adapter"
	"github.com/modular-project/orders-service/controller"
	"github.com/modular-project/orders-service/http/handler"
	"github.com/modular-project/orders-service/model"
	"github.com/modular-project/orders-service/storage"
	pf "github.com/modular-project/protobuffers/order/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newDBConn() storage.DBConnection {
	env := "ORDER_DB_HOST"
	host, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("environment variable (%s) not found", env)
	}
	env = "ORDER_DB_PORT"
	port, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("environment variable (%s) not found", env)
	}
	env = "ORDER_DB_USER"
	user, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("environment variable (%s) not found", env)
	}
	env = "ORDER_DB_PWD"
	pwd, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("environment variable (%s) not found", env)
	}
	env = "ORDER_DB_NAME"
	name, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("environment variable (%s) not found", env)
	}
	return storage.DBConnection{
		TypeDB:   storage.POSTGRESQL,
		User:     user,
		Password: pwd,
		Host:     host,
		Port:     port,
		NameDB:   name,
	}
}

func newPaypalService() controller.PaypalServicer {
	env := "PP_CLTID"
	ctlID, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("environment variable (%s) not found", env)
	}
	env = "PP_SECRET"
	secret, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("environment variable (%s) not found", env)
	}
	env = "PP_API"
	api, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("environment variable (%s) not found", env)
	}
	env = "FRONT_HOST"
	sUrl, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("environment variable (%s) not found", env)
	}
	env = "APP_NAME"
	bName, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("environment variable (%s) not found", env)
	}
	ps, err := adapter.NewPaypalSerive(ctlID, secret, api, sUrl, bName)
	if err != nil {
		log.Fatalf("fatal at started paypal service: %s", err)
	}
	return ps
}

func Recovery(i interface{}) error {
	return status.Errorf(codes.Unknown, "panic triggered: %v", i)
}

func startGRPC() *grpc.Server {
	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(Recovery),
	}
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(opts...),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_middleware.ChainStreamServer(),
		)),
	)
	return server
}

func main() {
	if err := storage.NewDB(newDBConn()); err != nil {
		log.Fatalf("fatal at start db: %s", err)
	}
	storage.Migrate(&model.Order{}, &model.OrderProduct{})
	ose := controller.NewOrderService(storage.NewOrderStorage())
	oss := controller.NewOrderStatusService(storage.NewOrderStatusStorage(), newPaypalService())
	env := "ORDER_HOST"
	host, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("environment variable (%s) not found", env)
	}
	env = "ORDER_PORT"
	port, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("environment variable (%s) not found", env)
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	ouc := handler.NewOrderUC(ose)
	osuc := handler.NewOrderStatusUC(oss)
	srv := startGRPC()
	pf.RegisterOrderServiceServer(srv, ouc)
	pf.RegisterOrderStatusServiceServer(srv, osuc)
	log.Printf("Order server started at %s:%s", host, port)
	err = srv.Serve(lis)
	if err != nil {
		log.Fatalf("failed to server at %s:%s, got error: %s", host, port, err)
	}
}
