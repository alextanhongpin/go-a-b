package main

import (
	"context"
	"log"
	"net/http"

	"github.com/golang/glog"

	pb "github.com/alextanhongpin/go-a-b/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

func run() error {
	ctx := context.Background()
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithInsecure()}
	if err := pb.RegisterBanditServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts); err != nil {
		return err
	}
	log.Println("listening to port *:3000. press ctrl + c to cancel.")
	return http.ListenAndServe(":3000", mux)
}

func main() {
	defer glog.Flush()
	glog.Fatal(run())
}
