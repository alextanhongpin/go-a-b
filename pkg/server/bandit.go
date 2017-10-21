package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"google.golang.org/grpc/codes"

	pb "github.com/alextanhongpin/go-a-b/proto"
	"github.com/alextanhongpin/go-bandit"
	bolt "github.com/coreos/bbolt"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type banditServer struct {
	db *bolt.DB
}

func (s *banditServer) GetExperiments(ctx context.Context, msg *pb.GetExperimentsRequest) (*pb.GetExperimentsResponse, error) {
	var experiments []*pb.Experiment
	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return fmt.Errorf("bucket not found: %s", bucket)
		}
		// There's a possibility to stream the data, but due to grpc-gateway limitation
		// this is not possible at the moment
		b.ForEach(func(k, v []byte) error {
			var exp *pb.Experiment
			if err := json.Unmarshal(v, &exp); err != nil {
				return err
			}
			experiments = append(experiments, exp)
			return nil
		})
		return nil
	}); err != nil {
		return nil, grpc.Errorf(codes.Internal, "Error getting experiments: %v", err)
	}

	return &pb.GetExperimentsResponse{
		Data:  experiments,
		Count: int64(len(experiments)),
	}, nil
}

func (s *banditServer) GetExperiment(ctx context.Context, msg *pb.GetExperimentRequest) (*pb.GetExperimentResponse, error) {
	id, err := uuid.FromString(msg.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "The id provided is invalid: %v", err)
	}
	var exp *pb.Experiment
	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return fmt.Errorf("bucket not found: %s", bucket)
		}
		v := b.Get([]byte(id.String()))
		log.Println("got v", v, len(v))
		if len(v) == 0 {
			return fmt.Errorf("the item does not exist or has been deleted")
		}
		if err := json.Unmarshal(v, &exp); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, grpc.Errorf(codes.Internal, "Error getting experiment: %v", err)
	}
	return &pb.GetExperimentResponse{
		Data: exp,
	}, nil
}

func (s *banditServer) PostExperiment(ctx context.Context, msg *pb.PostExperimentRequest) (*pb.PostExperimentResponse, error) {

	id := uuid.NewV4().String()
	msg.Data.Id = id
	msg.Data.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	msg.Data.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	// Do validation for nArms
	if msg.Data.N <= 0 {
		msg.Data.N = 2
	}

	if int64(len(msg.Data.Counts)) != msg.Data.N {
		msg.Data.Counts = make([]int64, msg.Data.N)
	}

	if int64(len(msg.Data.Rewards)) != msg.Data.N {
		msg.Data.Rewards = make([]float64, msg.Data.N)
	}

	if msg.Data.Epsilon == 0 {
		msg.Data.Epsilon = 0.1
	}

	experimentBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not marshal data: %v", err)
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		// keyTimestamp := []byte(time.Now().UTC().Format(time.RFC3339))
		if err := b.Put([]byte(id), experimentBytes); err != nil {
			return err
		}
		return nil
	})
	return nil, grpc.Errorf(codes.Unimplemented, "This method is not implement")
}

func (s *banditServer) DeleteExperiment(ctx context.Context, msg *pb.DeleteExperimentRequest) (*pb.DeleteExperimentResponse, error) {
	id, err := uuid.FromString(msg.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "The id provided is invalid: %v", err)
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if err := b.Delete([]byte(id.String())); err != nil {
			return fmt.Errorf("error deleting item from bucket: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "error deleting data from bucket: %v", err)
	}
	return &pb.DeleteExperimentResponse{
		Ok: true,
	}, nil
}

func (s *banditServer) GetArm(ctx context.Context, msg *pb.GetArmRequest) (*pb.GetArmResponse, error) {
	id, err := uuid.FromString(msg.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "The id provided is invalid: %v", err)
	}
	var exp *pb.Experiment
	if err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		v := b.Get([]byte(id.String()))
		log.Println("got v", v, len(v))
		if len(v) == 0 {
			return fmt.Errorf("the item does not exist or has been deleted")
		}
		if err := json.Unmarshal(v, &exp); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, grpc.Errorf(codes.Internal, "%v", err)
	}
	eps := bandit.NewEpsilonGreedy(int(exp.N), exp.Epsilon)
	eps.SetRewards(exp.Rewards)
	eps.SetCounts(exp.Counts)
	arm, _ := eps.SelectArm()

	return &pb.GetArmResponse{
		Arm: int64(arm),
	}, nil
}

func (s *banditServer) UpdateArm(ctx context.Context, msg *pb.UpdateArmRequest) (*pb.UpdateArmResponse, error) {
	id, err := uuid.FromString(msg.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "The id provided is invalid: %v", err)
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		var exp *pb.Experiment
		b := tx.Bucket(bucket)
		v := b.Get([]byte(id.String()))

		if len(v) == 0 {
			return fmt.Errorf("the item does not exist or has been deleted")
		}
		if err := json.Unmarshal(v, &exp); err != nil {
			return fmt.Errorf("error unmarshalling data: %v", err)
		}

		eps := bandit.NewEpsilonGreedy(int(exp.N), exp.Epsilon)
		eps.SetRewards(exp.Rewards)
		eps.SetCounts(exp.Counts)
		eps.Update(int(msg.Arm), float64(msg.Reward))

		exp.Rewards = eps.Rewards
		exp.Counts = eps.Counts
		exp.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		newByte, err := json.Marshal(exp)
		if err != nil {
			return fmt.Errorf("error marshalling data: %v", err)
		}
		if err = b.Put([]byte(id.String()), newByte); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error(), nil)
	}
	return &pb.UpdateArmResponse{
		Ok: true,
	}, nil
}
