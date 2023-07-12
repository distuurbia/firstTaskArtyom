// A main package.
package main

import (
	"context"
	"fmt"
	"log"
	"net"

	_ "github.com/distuurbia/firstTaskArtyom/docs"
	"github.com/distuurbia/firstTaskArtyom/internal/config"
	"github.com/distuurbia/firstTaskArtyom/internal/interceptor"
	"github.com/distuurbia/firstTaskArtyom/proto_services"
	"google.golang.org/grpc"

	"github.com/caarlos0/env"
	"github.com/distuurbia/firstTaskArtyom/internal/handler"
	"github.com/distuurbia/firstTaskArtyom/internal/repository"
	"github.com/distuurbia/firstTaskArtyom/internal/service"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/go-playground/validator.v9"
)

const (
	// PostgresDatabase represents the identifier for the Postgres database.
	PostgresDatabase = 1
	// MongoDBDatabase represents the identifier for the MongoDB database.
	MongoDBDatabase = 2
)

func connectPostgres(cfg *config.Config) (*pgxpool.Pool, error) {
	conf, err := pgxpool.ParseConfig(cfg.PostgresPath)
	if err != nil {
		return nil, fmt.Errorf("error in method pgxpool.ParseConfig: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("error in method pgxpool.NewWithConfig: %v", err)
	}
	return pool, nil
}

func connectMongo(cfg *config.Config) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(cfg.MongoPath)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error in method mongo.Connect(): %v", err)
	}
	return client, nil
}

func connectRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	return client
}

//nolint:funlen //Disabled because project have too many connections.
func main() {
	var (
		database int
		handl    *handler.GRPCHandler
		cfg      config.Config
	)

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	redisClient := connectRedis(&cfg)
	defer func() {
		errClose := redisClient.Close()
		if errClose != nil {
			log.Fatalf("Failed to disconnect from Redis: %v", errClose)
		}
	}()
	repoRedis := repository.NewRedisRepository(redisClient)
	fmt.Print("Choose database:\n 1)Postgres\n 2)MongoDB\n")
	_, err := fmt.Scan(&database)
	if err != nil {
		fmt.Printf("Failed to read: %v", err)
	}
	v := validator.New()
	switch database {
	case PostgresDatabase:
		pool, errPGX := connectPostgres(&cfg)
		if errPGX != nil {
			fmt.Printf("Failed to connect to Postgres: %v", errPGX)
		}
		defer pool.Close()

		repoPostgres := repository.NewPgRepository(pool)
		carService := service.NewCarEntity(repoPostgres, repoRedis)
		userService := service.NewUserEntity(repoPostgres, &cfg)
		handl = handler.NewGRPCHandler(carService, userService, v)

	case MongoDBDatabase:
		mongoClient, errMongo := connectMongo(&cfg)
		if errMongo != nil {
			fmt.Printf("Failed to connect to MongoDB: %v", errMongo)
		}
		defer func() {
			errMongo := mongoClient.Disconnect(context.Background())
			if err != nil {
				fmt.Printf("Failed to disconnect from MongoDB: %v", errMongo)
			}
		}()

		repoMongo := repository.NewMongoRepository(mongoClient)
		carService := service.NewCarEntity(repoMongo, repoRedis)
		userService := service.NewUserEntity(repoMongo, &cfg)
		handl = handler.NewGRPCHandler(carService, userService, v)

	default:
		//nolint:gocritic
		log.Fatal("Not correct number!")
	}
	lis, err := net.Listen("tcp", "localhost:5433")
	if err != nil {
		log.Fatalf("cannot connect listener: %s", err)
	}
	customInterceptor := interceptor.NewCustomInterceptor(&cfg)
	serverRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(customInterceptor.UnaryInterceptor),
	)
	proto_services.RegisterCarServiceServer(serverRegistrar, handl)
	proto_services.RegisterUserServiceServer(serverRegistrar, handl)
	proto_services.RegisterImageServiceServer(serverRegistrar, handl)
	err = serverRegistrar.Serve(lis)
	if err != nil {
		log.Fatalf("cannot serve: %s", err)
	}
}
