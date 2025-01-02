package DBConnection

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient *mongo.Client
	DB          *mongo.Database
	Sqldb       *sql.DB
)

func DBConfig() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	initMongoDB()

	initMSSQL()
}

func initMongoDB() {
	uri := os.Getenv("MONGO_DB_CONNECTION_URL")
	dbName := os.Getenv("MONGO_DB_NAME")

	if uri == "" || dbName == "" {
		log.Fatal("MONGO_DB_CONNECTION_URL and MONGO_DB_NAME must be set in the environment variables")
	}

	clientOptions := options.Client().ApplyURI(uri)
	var err error
	MongoClient, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := MongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("Error pinging MongoDB: %v", err)
	}

	DB = MongoClient.Database(dbName)
	log.Println("Connected to MongoDB successfully")
}

func initMSSQL() {
	server := os.Getenv("MSSQL_SERVER")
	user := os.Getenv("MSSQL_USER")
	password := os.Getenv("MSSQL_PASSWORD")
	port := os.Getenv("MSSQL_PORT")
	database := os.Getenv("MSSQL_DATABASE")

	if server == "" || user == "" || password == "" || port == "" || database == "" {
		log.Fatal("MSSQL connection parameters must be set in the environment variables")
	}

	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
		user, password, server, port, database)

	var err error
	Sqldb, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatalf("Error connecting to MSSQL: %v", err)
	}

	err = Sqldb.Ping()
	if err != nil {
		log.Fatalf("Failed to ping MSSQL: %v", err)
	}

	log.Println("Connected to MSSQL successfully!")
}
