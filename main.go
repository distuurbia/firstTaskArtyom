// A main package.
package main

import (
	"context"
	"fmt"
	"log"
	"net"

	_ "github.com/distuurbia/firstTaskArtyom/docs"
	"github.com/distuurbia/firstTaskArtyom/internal/config"
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

func connectPostgres() (*pgxpool.Pool, error) {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
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
func connectMongo() (*mongo.Client, error) {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	clientOptions := options.Client().ApplyURI(cfg.MongoPath)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error in method mongo.Connect(): %v", err)
	}
	return client, nil
}

func connectRedis() (*redis.Client, error) {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
		return nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	return client, nil
}

// @title Car API
// @version 1.0

// @host localhost:5433
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

//nolint:funlen //Disabled because project have too many connections.
func main() {
	var (
		database int
		handl    *handler.GRPCHandler
	)
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	redisClient, err := connectRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer func() {
		errClose := redisClient.Close()
		if errClose != nil {
			log.Fatalf("Failed to disconnect from Redis: %v", errClose)
		}
	}()
	repoRedis := repository.NewRedisRepository(redisClient)
	fmt.Print("Choose database:\n 1)Postgres\n 2)MongoDB\n")
	_, err = fmt.Scan(&database)
	if err != nil {
		fmt.Printf("Failed to read: %v", err)
	}
	v := validator.New()
	switch database {
	case PostgresDatabase:
		pool, errPGX := connectPostgres()
		if errPGX != nil {
			fmt.Printf("Failed to connect to Postgres: %v", errPGX)
		}
		defer pool.Close()

		repoPostgres := repository.NewPgRepository(pool)
		carService := service.NewCarEntity(repoPostgres, repoRedis)
		userService := service.NewUserEntity(repoPostgres)
		handl = handler.NewGRPCHandler(carService, userService, v)

	case MongoDBDatabase:
		mongoClient, errMongo := connectMongo()
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
		userService := service.NewUserEntity(repoMongo)
		handl = handler.NewGRPCHandler(carService, userService, v)

	default:
		//nolint:gocritic
		log.Fatal("Not correct number!")
	}
	lis, err := net.Listen("tcp", "localhost:5433")
	if err != nil {
		log.Fatalf("cannot connect listener: %s", err)
	}
	serverRegistrar := grpc.NewServer()
	proto_services.RegisterCarServiceServer(serverRegistrar, handl)
	proto_services.RegisterUserServiceServer(serverRegistrar, handl)
	err = serverRegistrar.Serve(lis)
	if err != nil {
		log.Fatalf("cannot serve: %s", err)
	}
}
