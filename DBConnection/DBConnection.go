package DBConnection

import (
	"Blr_server_update/logs"
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

func InitMongoDB() (*mongo.Database, *mongo.Client, error) {
	SuccuseLog, ErrorLog := logs.Logger()
	if err := godotenv.Load(); err != nil {
		ErrorLog.Printf("Error loading .env file: %v\n", err)
	}

	uri := os.Getenv("MONGO_DB_CONNECTION_URI")
	dbName := os.Getenv("MONGO_DB_NAME")

	if uri == "" || dbName == "" {
		ErrorLog.Printf("MONGO_DB_CONNECTION_URL and MONGO_DB_NAME must be set in the environment variables\n")
	}

	clientOptions := options.Client().ApplyURI(uri)
	var err error
	MongoClient, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		ErrorLog.Printf("Error connecting to MongoDB: %v\n", err)
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := MongoClient.Ping(ctx, nil); err != nil {
		ErrorLog.Printf("Error pinging MongoDB: %v\n", err)
		return nil, nil, err
	}

	DB = MongoClient.Database(dbName)
	log.Printf("Connected to MongoDB successfully")
	SuccuseLog.Println("Connected to MongoDB successfully")
	return DB, MongoClient, nil
}

func InitMSSQL() (*sql.DB, error) {
	SuccessLog, ErrorLog := logs.Logger()
	if err := godotenv.Load(); err != nil {
		ErrorLog.Printf("Error loading .env file: %v\n", err)
	}

	server := os.Getenv("MSSQL_SERVER")
	user := os.Getenv("MSSQL_USER")
	password := os.Getenv("MSSQL_PASSWORD")
	port := os.Getenv("MSSQL_PORT")
	database := os.Getenv("MSSQL_DATABASE")

	if server == "" || user == "" || password == "" || port == "" || database == "" {
		ErrorLog.Printf("MSSQL connection parameters must be set in the environment variables\n")
	}

	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
		user, password, server, port, database)

	var err error
	Sqldb, err = sql.Open("sqlserver", connString)
	if err != nil {
		ErrorLog.Printf("Error connecting to MSSQL: %v\n", err)
		return nil, err
	}

	err = Sqldb.Ping()
	if err != nil {
		ErrorLog.Printf("Failed to ping MSSQL: %v\n", err)
		return nil, err
	}
	log.Printf("Connected to MSSQL successfully")
	SuccessLog.Println("Connected to MSSQL successfully!")
	return Sqldb, nil
}