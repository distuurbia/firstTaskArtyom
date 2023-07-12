package repository

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"testing"

	"github.com/distuurbia/firstTaskArtyom/internal/model"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mrpc *MongoRepository

var rpc *PgRepository

var rdsRps *RedisRepository

var testModel = model.Car{
	ID:             uuid.New(),
	Brand:          RandBrand(),
	ProductionYear: RandProductionYear(),
	IsRunning:      RandBool(),
}

/*Generating random fields*/

func RandBrand() string {
	currencies := []string{"Toyota", "Honda", "Mazda", "Skoda", "Lada"}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

func RandInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandProductionYear() int64 {
	return RandInt(1980, 2023)
}

func RandBool() bool {
	return rand.Intn(2) == 1
}

const (
	pgUsername = "artempg"
	pgPassword = "artempg"
	pgDB       = "db"
)

func SetupMongo() (*mongo.Client, func(), error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct pool: %w", err)
	}
	resource, err := pool.Run("mongo", "latest", []string{
		"MONGO_INITDB_ROOT_USERNAME=artemmdb",
		"MONGO_INITDB_ROOT_PASSWORD=artemmdb",
		"MONGO_INITDB_DATABASE=mdb"})
	if err != nil {
		return nil, nil, fmt.Errorf("could not start resource: %w", err)
	}

	port := resource.GetPort("27017/tcp")
	mongoURL := fmt.Sprintf("mongodb://artemmdb:artemmdb@localhost:%s", port)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect mongoDB: %w", err)
	}
	cleanup := func() {
		client.Disconnect(context.Background())
		pool.Purge(resource)
	}
	return client, cleanup, nil
}

func SetupPostgres() (*pgxpool.Pool, func(), error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct pool: %w", err)
	}
	resource, err := pool.Run("postgres", "latest", []string{
		"POSTGRES_USER=artempg",
		"POSTGRES_PASSWORD=artempg",
		"POSTGRES_DB=db"})
	if err != nil {
		logrus.Fatalf("can't start postgres container: %s", err)
	}
	cmd := exec.Command(
		"flyway",
		fmt.Sprintf("-user=%s", pgUsername),
		fmt.Sprintf("-password=%s", pgPassword),
		"-locations=filesystem:../../migrations",
		fmt.Sprintf("-url=jdbc:postgresql://%s:%s/%s", "localhost", resource.GetPort("5432/tcp"), pgDB), "-connectRetries=10",
		"migrate",
	)
	err = cmd.Run()
	if err != nil {
		logrus.Fatalf("can't run migration: %s", err)
	}
	dbURL := fmt.Sprintf("postgresql://artempg:artempg@localhost:%s/db", resource.GetPort("5432/tcp"))
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		logrus.Fatalf("can't parse config: %s", err)
	}
	dbpool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		logrus.Fatalf("can't connect to postgtres: %s", err)
	}
	cleanup := func() {
		dbpool.Close()
		pool.Purge(resource)
	}
	return dbpool, cleanup, nil
}

func SetupRedis() (*redis.Client, func(), error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct pool: %w", err)
	}
	resource, err := pool.Run("redis", "latest", []string{
		"REDIS_PASSWOES=artemrdb"})
	if err != nil {
		return nil, nil, fmt.Errorf("could not run the pool: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%s", resource.GetPort("6379/tcp")),
		DB:   0,
	})
	cleanup := func() {
		client.Close()
		pool.Purge(resource)
	}
	return client, cleanup, nil
}

func TestMain(m *testing.M) {
	dbpool, cleanupPgx, err := SetupPostgres()
	if err != nil {
		fmt.Println("can't construct the pool: ", err)
		cleanupPgx()
		os.Exit(1)
	}
	rpc = NewPgRepository(dbpool)

	client, cleanupMongo, err := SetupMongo()
	if err != nil {
		fmt.Println(err)
		cleanupMongo()
		os.Exit(1)
	}
	mrpc = NewMongoRepository(client)

	rdsClient, cleanupRds, err := SetupRedis()
	if err != nil {
		fmt.Println(err)
		cleanupRds()
		os.Exit(1)
	}
	rdsRps = NewRedisRepository(rdsClient)

	exitCode := m.Run()

	cleanupPgx()
	cleanupMongo()
	cleanupRds()
	os.Exit(exitCode)
}
