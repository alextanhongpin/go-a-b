package main

import (
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/alextanhongpin/go-a-b/proto"
	bolt "github.com/coreos/bbolt"
)

const database = "bandit.db"

var bucket = []byte("bandit")

const port = ":50051"

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()

	db, err := bolt.Open(database, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("error connecting to database: %s", err.Error())
	}
	defer db.Close()

	// Create bucket if it doesn't exist
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Printf("error creating bucket: %s\n", err.Error())
	}

	pb.RegisterBanditServiceServer(grpcServer, &banditServer{
		db: db,
	})

	log.Printf("listening to port *%s. press ctrl + c to cancel.\n", port)
	log.Fatal(grpcServer.Serve(lis))
}

// Run a cron periodically that checks the keys and updates the rewards
