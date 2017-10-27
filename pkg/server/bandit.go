package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	bolt "github.com/coreos/bbolt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/alextanhongpin/go-a-b/proto"
	"github.com/alextanhongpin/go-bandit"
)

func init() {
	// Ensure that a random value is returned by the bandit algorithm
	rand.Seed(time.Now().UnixNano())
}

type banditServer struct {
	db    *bolt.DB
	cache *redis.Client
}

func (s *banditServer) GetExperiments(ctx context.Context, msg *pb.GetExperimentsRequest) (*pb.GetExperimentsResponse, error) {
	var experiments []*pb.Experiment
	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return fmt.Errorf("bucket: %s does not exist", bucket)
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
		return nil, grpc.Errorf(codes.Internal, "GetExperiments: %v", err)
	}

	return &pb.GetExperimentsResponse{
		Data:  experiments,
		Count: int64(len(experiments)),
	}, nil
}

func (s *banditServer) GetExperiment(ctx context.Context, msg *pb.GetExperimentRequest) (*pb.GetExperimentResponse, error) {
	id, err := uuid.FromString(msg.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	var exp *pb.Experiment
	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return fmt.Errorf("bucket: %s does not exist", bucket)
		}

		v := b.Get([]byte(id.String()))
		if len(v) == 0 {
			return fmt.Errorf("key-value: value for %s does not exist", id.String())
		}
		if err := json.Unmarshal(v, &exp); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, grpc.Errorf(codes.Internal, "GetExperiment: %v", err)
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
		return nil, grpc.Errorf(codes.Internal, err)
	}

	if err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		// keyTimestamp := []byte(time.Now().UTC().Format(time.RFC3339))
		if err := b.Put([]byte(id), experimentBytes); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, grpc.Errorf(codes.Internal, "PostExperiment: %v", err)
	}

	return *pb.PostExperimentResponse{
		Id: id,
	}, nil
}

func (s *banditServer) DeleteExperiment(ctx context.Context, msg *pb.DeleteExperimentRequest) (*pb.DeleteExperimentResponse, error) {
	id, err := uuid.FromString(msg.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "The id provided is invalid: %v", err)
	}
	if err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if err := b.Delete([]byte(id.String())); err != nil {
			return fmt.Errorf("delete: %v", err)
		}
		return nil
	}); err != nil {
		return nil, grpc.Errorf(codes.Internal, "DeleteExperiment: %v", err)
	}

	return &pb.DeleteExperimentResponse{
		Ok: true,
	}, nil
}

func (s *banditServer) GetArm(ctx context.Context, msg *pb.GetArmRequest) (*pb.GetArmResponse, error) {
	id, err := uuid.FromString(msg.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	var exp *pb.Experiment
	if err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		v := b.Get([]byte(id.String()))
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
	arm := eps.SelectArm()

	// create unique id
	armID := "1"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        armID,
		ExpiresAt: 300, // 5 minutes (make this a config)
		Issuer:    "",
	})
	ss, err := token.SignedString(jwtKey)
	// Create a unique id that the user can associate with
	// redis set arm:arm_id:chosen_arm
	return &pb.GetArmResponse{
		Arm:   int64(arm),
		Token: ss,
	}, nil
}

func (s *banditServer) UpdateArm(ctx context.Context, msg *pb.UpdateArmRequest) (*pb.UpdateArmResponse, error) {
	// User needs to pass the unique id that is generated when selecting the arm
	// If both the id matches, then perform the update
	// This is useful for two reason:
	// 1. Unique tracking
	// 2. Setting the reward to 0 if user did not response
	id, err := uuid.FromString(msg.Id)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "The id provided is invalid: %v", err)
	}
	// redis get arm:arm_id:chosen_arm
	// If it doesn't match, then error message unable to update

	if err = s.db.Update(func(tx *bolt.Tx) error {
		var exp *pb.Experiment
		b := tx.Bucket(bucket)
		v := b.Get([]byte(id.String()))

		if len(v) == 0 {
			return fmt.Errorf("the item does not exist or has been deleted")
		}
		if err := json.Unmarshal(v, &exp); err != nil {
			return err
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
			return err
		}
		if err = b.Put([]byte(id.String()), newByte); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	return &pb.UpdateArmResponse{
		Ok: true,
	}, nil
}

// func ValidateArm() {
// 	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJuYmYiOjE0NDQ0Nzg0MDB9.u1riaD1rW97opCoAuRCTy4w58Br-Zk-bh7vLiRIsrpU"

// 	// Parse takes the token string and a function for looking up the key. The latter is especially
// 	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
// 	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
// 	// to the callback, providing flexibility.
// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		// Don't forget to validate the alg is what you expect:
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
// 		}

// 		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
// 		return hmacSampleSecret, nil
// 	})

// 	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 		fmt.Println(claims["foo"], claims["nbf"])
// 	} else {
// 		fmt.Println(err)
// 	}
// }
