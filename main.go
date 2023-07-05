// A main package.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/distuurbia/firstTaskArtyom/docs"
	"github.com/distuurbia/firstTaskArtyom/internal/config"
	"github.com/distuurbia/firstTaskArtyom/internal/handler"
	custommiddleware "github.com/distuurbia/firstTaskArtyom/internal/middleware"
	"github.com/distuurbia/firstTaskArtyom/internal/repository"
	"github.com/distuurbia/firstTaskArtyom/internal/service"
	"github.com/caarlos0/env"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
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
	// err = client.Ping(context.Background(), nil)
	// if err != nil {
	// 	return nil, fmt.Errorf("error in method client.Ping(): %v", err)
	// }
	return client, nil
}

func connectRedis() (*redis.Client, error) {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	// _, err := client.Ping(client.Context()).Result()
	// if err != nil {
	// 	return nil, fmt.Errorf("error in method client.Ping(): %v", err)
	// }
	return client, nil
}

// consumer read messages from redis stream and deleted it.
func consumer(client *redis.Client) {
	ctx := context.Background()
	for {
		result, err := client.XRead(ctx, &redis.XReadArgs{
			Streams: []string{"messagestream", "0"},
			Count:   1,
			Block:   0,
		}).Result()

		if err != nil {
			log.Fatalf("Error when reading messages from Redis Stream:%v", err)
		}

		for _, message := range result[0].Messages {
			fmt.Println("Received message:", message.Values["message"])
			_, err := client.XDel(ctx, "messagestream", message.ID).Result()
			if err != nil {
				log.Fatalf("Error when deleting message from Redis Stream:%v", err)
			}
		}
	}
}

// producer write in redis stream message once every 5 seconds.
func producer(client *redis.Client, timeSecondSleep time.Duration) {
	ctx := context.Background()
	for {
		_, err := client.XAdd(ctx, &redis.XAddArgs{
			Stream: "messagestream",
			Values: map[string]interface{}{
				"message": fmt.Sprintf("It's working %s", time.Now()),
			},
		}).Result()
		if err != nil {
			log.Fatalf("Error when writing a message to Redis Stream:%v", err)
		}
		time.Sleep(timeSecondSleep * time.Second)
	}
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
		database        int
		handl           *handler.Handler
		
	)
	const timeSecondSleep time.Duration = 5
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

	// go consumer(redisClient)
	// go producer(redisClient, timeSecondSleep)

	fmt.Print("Choose database:\n 1)Postgres\n 2)MongoDB\n")
	_, err = fmt.Scan(&database)
	if err != nil {
		fmt.Printf("Failed to read: %v", err)
	}
	v := validator.New()
	switch database {
	case PostgresDatabase:
		pool, err := connectPostgres()
		if err != nil {
			fmt.Printf("Failed to connect to Postgres: %v", err)
		}
		defer pool.Close()

		repoPostgres := repository.NewPgRepository(pool)
		carService := service.NewCarEntity(repoPostgres, repoRedis)
		userService := service.NewUserEntity(repoPostgres)
		handl = handler.NewHandler(carService, userService, v)

	case MongoDBDatabase:
		mongoClient, err := connectMongo()
		if err != nil {
			fmt.Printf("Failed to connect to MongoDB: %v", err)
		}
		defer func() {
			err := mongoClient.Disconnect(context.Background())
			if err != nil {
				fmt.Printf("Failed to disconnect from MongoDB: %v", err)
			}
		}()

		repoMongo := repository.NewMongoRepository(mongoClient)
		carService := service.NewCarEntity(repoMongo, repoRedis)
		userService := service.NewUserEntity(repoMongo)
		handl = handler.NewHandler(carService, userService, v)

	default:
		//nolint:gocritic
		log.Fatal("Not correct number!")
	}
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.POST("/signup", handl.SignUpUser)
	e.POST("/login", handl.GetByLogin)
	e.POST("/refresh", handl.RefreshToken)
	e.POST("/upload", handl.UploadImage)
	e.GET("/download/:filename", handl.DownloadImage)
	e.POST("/car", handl.Create, custommiddleware.JWTMiddleware)
	e.GET("/car/:id", handl.Get, custommiddleware.JWTMiddleware)
	e.PUT("/car", handl.Update, custommiddleware.JWTMiddleware)
	e.DELETE("/car/:id", handl.Delete, custommiddleware.JWTMiddleware)
	e.GET("/car", handl.GetAll, custommiddleware.JWTMiddleware)

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	address := fmt.Sprintf(":%d", cfg.Port)
	e.Logger.Fatal(e.Start(address))
}
