package services

import (
	"GoFiber_Project01/DBConnection"
	"GoFiber_Project01/config"

	// "GoFiber_Project01/helpers"
	"GoFiber_Project01/logs"
	"context"

	// "database/sql"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func UpdatePTP_MAA() error {
	var (
		collectionName = "in_consignment"
		docIndex       int
		totalDocs      int
		successCount   int
	)
	_, ErrorLog := logs.Logger()
	configration := config.Init()
	db, MongoClient, err := DBConnection.InitMongoDB()
	if err != nil {
		log.Printf("Error Connected to MongoDB: %v\n", err)
		return nil
	}
	// sqldb, err := DBConnection.InitMSSQL()
	// if err != nil {
	// 	log.Printf("Error Connected to MSSQL %v\n", err)
	// 	return nil
	// }
	// configration := config.Config
	// db := DBConnection.DB
	// MongoClient := DBConnection.MongoClient
	// sqldb := DBConnection.Sqldb
	// defer sqldb.Close()
	defer MongoClient.Disconnect(context.Background())
	successCount = 0

	query := bson.M{
		"txn":               "ptp",
		"cno":               bson.M{"$not": bson.M{"$regex": configration["branchCode"], "$options": "i"}},
		"maa_server_update": bson.M{"$exists": false},
	}

	cursor, err := db.Collection(collectionName).Find(context.Background(), query)
	if err != nil {
		ErrorLog.Printf("failed to query documents: %v\n", err)
		return nil
	}
	defer cursor.Close(context.Background())

	var data []bson.M
	if err := cursor.All(context.Background(), &data); err != nil {
		ErrorLog.Printf("failed to parse documents: %v", err)
		return nil
	}
	totalDocs = len(data)
	log.Printf("Found %d docs", totalDocs)
	if totalDocs == 0 {
		ErrorLog.Printf("Found %d docs\n", totalDocs)
		return nil
	}
	for idx, doc := range data {
		docIndex = idx + 1
		log.Printf("%d/%d - CNo: %v try to update", docIndex, totalDocs, doc["cno"])

		// result := insertPTPDetails(sqldb, doc)
		// if result == nil {
		// 	updateResult, err := db.Collection(collectionName).UpdateOne(context.Background(), bson.M{"_id": doc["_id"]}, bson.M{"$set": bson.M{"maa_server_update": 1}})
		// 	if err != nil {
		// 		ErrorLog.Printf("Error Failed to update document CNo: %v\n", doc["cno"])
		// 		continue

		// 	}
		// 	if updateResult.ModifiedCount == 1 {
		// 		successCount++
		// 		log.Printf("%d/%d - CNo: %v updated successfully\n", docIndex, totalDocs, doc["cno"])
		// 		SuccessLog.Printf("%d/%d - CNo: %v updated successfully\n", docIndex, totalDocs, doc["cno"])
		// 	}
		// }
	}

	stat := map[string]interface{}{
		"date":        time.Now(),
		"txn":         "PTP",
		"server":      "MAA",
		"totalDocs":   totalDocs,
		"successDocs": successCount,
	}
	// if _, err := db.Collection("main_server_update").InsertOne(context.Background(), bson.M(stat)); err != nil {
	// 	ErrorLog.Printf("failed to insert stats: %v\n", err)
	// }
	fmt.Println(stat)
	return nil
}

// func insertPTPDetails(sqldb *sql.DB, r bson.M) error {
// 	_, ErrorLog := logs.Logger()
// 	configration := config.Init()
// 	origin, ok := r["cno"].(string)
// 	if !ok || len(origin) < 3 {
// 		return fmt.Errorf("invalid CNo format: %v", r["cno"])
// 	}
// 	origin = origin[:3]
// 	err, dt := helpers.StringToDateConverter(r["created_on"])
// 	if err {
// 		fmt.Println(err)
// 		return nil
// 	}

// 	cmd := "EXEC PtoP_Insert @tDate, @SlNo, @Origin, @StnCode, @CNo, @bcode, @rmks, @pntNo, @tload, @PTime, @opcode, @weight, @pcs, @type"
// 	_, err = sqldb.Exec(cmd,
// 		sql.Named("tDate", dt.Format("2006-01-02")),
// 		sql.Named("SlNo", 1),
// 		sql.Named("Origin", origin),
// 		sql.Named("StnCode", configration["branchCode"].(string)),
// 		sql.Named("CNo", r["cno"]),
// 		sql.Named("bcode", func() any {
// 			if r["pccode"] != nil {
// 				return r["pccode"]
// 			}
// 			return "PC00"
// 		}()),
// 		sql.Named("rmks", "ND"),
// 		sql.Named("pntNo", r["txn_id"]),
// 		sql.Named("tload", 1),
// 		sql.Named("PTime", dt.Format("15:04")),
// 		sql.Named("opcode", configration["branchCode"].(string)),
// 		sql.Named("weight", func() float32 {
// 			if r["weight"] != nil {
// 				return r["weight"].(float32)
// 			}
// 			return 0.100
// 		}()),
// 		sql.Named("pcs", func() any {
// 			if r["pieces"] != nil {
// 				return r["pieces"]
// 			}
// 			return 1
// 		}()),
// 		sql.Named("type", "0"),
// 	)

// 	if err != nil {
// 		ErrorLog.Printf("Failed to execute MSSQL command: %v\n", err)
// 		return fmt.Errorf("failed to execute MSSQL command: %w", err)
// 	}
// 	return nil
// }
