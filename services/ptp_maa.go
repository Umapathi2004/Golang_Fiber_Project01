package services

import (
	"GoFiber_Project01/DBConnection"
	"GoFiber_Project01/config"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"go.mongodb.org/mongo-driver/bson"
)

func UpdatePTP_MAA() error {
	var (
		collectionName = "in_consignment"
		docIndex       int
		totalDocs      int
		successCount   int
	)
	configration := config.Config
	db := DBConnection.DB
	MongoClient := DBConnection.MongoClient
	sqldb := DBConnection.Sqldb
	defer sqldb.Close()
	defer MongoClient.Disconnect(context.Background())
	successCount = 0

	query := bson.M{
		"txn":               "ptp",
		"cno":               bson.M{"$not": bson.M{"$regex": configration["branchCode"], "$options": "i"}},
		"maa_server_update": bson.M{"$exists": false},
	}

	cursor, err := db.Collection(collectionName).Find(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to query documents: %v", err)
	}
	defer cursor.Close(context.Background())

	var data []bson.M
	if err := cursor.All(context.Background(), &data); err != nil {
		return fmt.Errorf("failed to parse documents: %v", err)
	}
	totalDocs = len(data)
	log.Printf("Found %d docs", totalDocs)

	for idx, doc := range data {
		docIndex = idx + 1
		log.Printf("%d/%d - CNo: %v try to update", docIndex, totalDocs, doc["cno"])

		result := insertPTPDetails(sqldb, doc)
		if result == nil {
			updateResult, err := db.Collection(collectionName).UpdateOne(context.Background(), bson.M{"_id": doc["_id"]}, bson.M{"$set": bson.M{"maa_server_update": 1}})
			if err != nil || updateResult.ModifiedCount != 1 {
				log.Printf("Failed to update document: %v", doc["cno"])
			} else {
				successCount++
				log.Printf("%d/%d - %v updated to the local DB", docIndex, totalDocs, doc["cno"])
			}
		}
	}

	stat := map[string]interface{}{
		"date":        time.Now(),
		"txn":         "PTP",
		"server":      "MAA",
		"totalDocs":   totalDocs,
		"successDocs": successCount,
	}
	if _, err := db.Collection("main_server_update").InsertOne(context.Background(), bson.M(stat)); err != nil {
		return fmt.Errorf("failed to insert stats: %v", err)
	}
	return nil
}

func insertPTPDetails(sqldb *sql.DB, r bson.M) error {
	configration := config.Config
	origin, ok := r["cno"].(string)
	if !ok || len(origin) < 3 {
		return fmt.Errorf("invalid CNo format: %v", r["cno"])
	}
	origin = origin[:3]

	dtStr, ok := r["created_on"].(string)
	if !ok {
		return fmt.Errorf("invalid created_on field in record: %v", r)
	}
	dt, err := time.Parse(time.RFC3339, dtStr)
	if err != nil {
		return fmt.Errorf("error parsing created_on for CNo %s: %v", r["cno"], err)
	}
	cmd := "EXEC PtoP_Insert @tDate, @SlNo, @Origin, @StnCode, @CNo, @bcode, @rmks, @pntNo, @tload, @PTime, @opcode, @weight, @pcs, @type"
	_, err = sqldb.Exec(cmd,
		sql.Named("tDate", dt.Format("2006-01-02")),
		sql.Named("SlNo", 1),
		sql.Named("Origin", origin),
		sql.Named("StnCode", configration["branchCode"]),
		sql.Named("CNo", r["cno"]),
		sql.Named("bcode", func() string {
			if r["pccode"] != nil {
				return r["pccode"].(string)
			}
			return "PC00"
		}()),
		sql.Named("rmks", "ND"),
		sql.Named("pntNo", r["txn_id"]),
		sql.Named("tload", 1),
		sql.Named("PTime", dt.Format("15:04")),
		sql.Named("opcode", configration["branchCode"]),
		sql.Named("weight", func() float64 {
			if r["weight"] != nil {
				return r["weight"].(float64)
			}
			return 0.100
		}()),
		sql.Named("pcs", func() int {
			if r["pieces"] != nil {
				return r["pieces"].(int)
			}
			return 1
		}()),
		sql.Named("type", "0"),
	)

	if err != nil {
		log.Printf("Failed to execute MSSQL command: %v\n", err)
		return fmt.Errorf("failed to execute MSSQL command: %w", err)
	}
	return nil
}
