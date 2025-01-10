package services

import (
	"GoFiber_Project01/DBConnection"
	"GoFiber_Project01/api_request"
	"GoFiber_Project01/config"
	"GoFiber_Project01/helpers"
	"GoFiber_Project01/logs"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
)

var (
	// customerCodes = []string{"PAL101", "KRM209", "TAB742"}
	customer   = make(map[string]interface{})
	cutOffDate = time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC)
)

func UpdateBookingConsignment() error {
	var (
		collectionName = "booking_consignment"
		docIndex       int
		totalDocs      int
		successCount   int
	)
	_, ErrorLog := logs.Logger()
	db, MongoClient, err := DBConnection.InitMongoDB()
	configration := config.Init()
	if err != nil {
		log.Printf("Error Connected to MongoDB: %v\n", err)
		return nil
	}
	// db := DBConnection.DB
	// MongoClient := DBConnection.MongoClient
	defer MongoClient.Disconnect(context.Background())
	errenv := godotenv.Load()
	if errenv != nil {
		ErrorLog.Println("Error loading .env file")
		return nil
	}

	customerCodes := strings.Split(os.Getenv("customerCodes"), ",")

	for _, customerCode := range customerCodes {
		query := bson.M{"_id": customerCode}

		if err := db.Collection("customer").FindOne(context.Background(), query).Decode(&customer); err != nil {
			ErrorLog.Printf("Error finding customer with code %s: %v\n", customerCode, err)
			continue
		}

		query = bson.M{
			"customer_code":           customerCode,
			"bdate":                   bson.M{"$gte": cutOffDate},
			"booking_info_updated_on": bson.M{"$exists": false},
		}

		cursor, err := db.Collection(collectionName).Find(context.Background(), query)
		if err != nil {
			ErrorLog.Printf("Error finding documents for customer %s: %v\n", customerCode, err)
			continue
		}
		defer cursor.Close(context.Background())

		var data []bson.M
		if err := cursor.All(context.Background(), &data); err != nil {
			ErrorLog.Printf("Error reading cursor for customer %s: %v\n", customerCode, err)
			continue
		}
		totalDocs = len(data)
		log.Printf("Customer %s: Found %d documents", customerCode, totalDocs)
		if totalDocs == 0 {
			ErrorLog.Printf("Customer %s: Found %d documents\n", customerCode, totalDocs)
			continue
		}
		for idx, doc := range data {
			param := getBookingUpdateUrl(doc)
			if param != nil {
				docIndex = idx + 1
				log.Printf("%d/%d - Processing CNo: %v", docIndex, totalDocs, doc["cno"])
				bookingUpdateUrl, _ := configration["bookingUpdateUrl"].(string)

				result := api_request.SendData("Booking Info", bookingUpdateUrl, param, doc)
				fmt.Println(result)
				// 	if result["success"] == true {
				// 		updateResult, err := db.Collection(collectionName).UpdateOne(
				// 			context.Background(),
				// 			bson.M{"_id": doc["_id"]},
				// 			bson.M{"$set": bson.M{"booking_info_updated_on": time.Now()}},
				// 		)
				// 		if err != nil {
				// 			ErrorLog.Printf("Error Failed to update document CNo: %v\n", doc["cno"])
				// 			continue

				// 		}
				// 		if updateResult.ModifiedCount == 1 {
				// 			successCount++
				// 			log.Printf("%d/%d - CNo: %v updated successfully\n", docIndex, totalDocs, doc["cno"])
				// 			SuccessLog.Printf("%d/%d - CNo: %v updated successfully\n", docIndex, totalDocs, doc["cno"])
				// 		}
				// 	}
			}
		}

		stat := map[string]interface{}{
			"date":          time.Now(),
			"txn":           "Booking Update",
			"server":        "BLR",
			"customer_code": customerCode,
			"totalDocs":     totalDocs,
			"successDocs":   successCount,
		}
		// if _, err := db.Collection("main_server_update").InsertOne(context.Background(), bson.M(stat)); err != nil {
		// 	ErrorLog.Printf("Failed to insert stats for customer %s: %v\n", customerCode, err)
		// 	continue
		// }
		fmt.Println(stat)
	}
	return nil
}

func getBookingUpdateUrl(r bson.M) map[string]interface{} {
	configration := config.Init()
	cno := r["cno"]
	err, dt := helpers.StringToDateConverter(r["bdate"])
	if err {
		fmt.Println(err)
		return nil
	}
	// dtStr, ok := r["bdate"].(string)
	// if !ok {
	// 	log.Printf("Invalid created_on field in record: %v", r)
	// 	return nil
	// }

	// dt, err := time.Parse(time.RFC3339, dtStr)
	// if err != nil {
	// 	log.Printf("Error parsing created_on for CNo %s: %v", cno, err)
	// 	return nil
	// }

	sample := map[string]interface{}{
		"POD_NO":    cno,
		"REF_NO":    "",
		"CLIENT":    "TRZPAL101",
		"BDATE":     dt.Format("2006-01-02"),
		"CONSIGNEE": r["customer_name"],
		"CC":        r["pc_code"],
		"WEIGHT": func() any {
			if r["bweight"] != nil {
				return r["bweight"]
			}
			return 0.100
		}(),
		"PIECES": func() any {
			if r["bpieces"] != nil {
				return r["bpieces"]
			}
			return 1
		}(),
		"BILL_REF":    "",
		"AMOUNT":      0,
		"DESTINATION": r["bdes_name"],
		"DESTN":       r["bdes"],
		"ORIGIN":      configration["branchCode"].(string),
		"SENDER_MOB": func() any {
			if r["primary_contact_number"] != nil {
				return r["primary_contact_number"]
			}
			return ""
		}(),
		"SENDER_EMAIL": func() any {
			if r["primary_contact_email"] != nil {
				return r["primary_contact_email"]
			}
			return ""
		}(),
		"CUST_INVOICE":      "",
		"CUST_INVOICEAMT":   "",
		"FLYER_NO":          "",
		"xmlorigin":         configration["branchCode"].(string),
		"id":                configration["loginId"],
		"RECIPIENT_ADDRESS": r["to_address"],
		"RECIPIENT_MOB":     "",
		"PINCODE":           r["pincode"],
	}

	return sample
}
