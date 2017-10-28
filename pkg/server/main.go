package main

import (
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/alextanhongpin/go-a-b/proto"
	bolt "github.com/coreos/bbolt"
	"github.com/go-redis/redis"
	"github.com/robfig/cron"
)

var bucket = []byte("bandit")

const database = "bandit.db"
const port = ":50051"
const jwtKey = "secret"

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	// SETUP: Redis
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// SETUP: Bolt
	db, err := bolt.Open(database, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("error connecting to database: %s", err.Error())
	}
	defer db.Close()

	// SETUP: cron
	c := cron.New()
	c.AddFunc("0 */5 * * * *", func() {
		// get a list of keys from redis that matches the pattern
		// perform a write to the database
		log.Println("running every 5 minutes")
	})

	// Create bucket if it doesn't exist
	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Printf("error creating bucket: %s\n", err.Error())
	}

	// SETUP: gRPC
	grpcServer := grpc.NewServer()
	pb.RegisterBanditServiceServer(grpcServer, &banditServer{
		db:    db,
		cache: client,
	})

	log.Printf("listening to port *%s. press ctrl + c to cancel.\n", port)
	log.Fatal(grpcServer.Serve(lis))
}

// Run a cron periodically that checks the keys and updates the rewards
