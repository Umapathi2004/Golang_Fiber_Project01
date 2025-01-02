package services

import (
	"GoFiber_Project01/DBConnection"
	"GoFiber_Project01/api_request"
	"GoFiber_Project01/config"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var (
	customerCodes = []string{"PAL101", "KRM209", "TAB742"}
	customer      = make(map[string]interface{})
	cutOffDate    = time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC)
)

func UpdateBookingConsignment() error {
	var (
		collectionName = "booking_consignment"
		docIndex       int
		totalDocs      int
		successCount   int
	)
	configration := config.Config
	db := DBConnection.DB
	MongoClient := DBConnection.MongoClient
	defer MongoClient.Disconnect(context.Background())

	for _, customerCode := range customerCodes {
		query := bson.M{"_id": customerCode}

		if err := db.Collection("customer").FindOne(context.Background(), query).Decode(&customer); err != nil {
			log.Printf("Error finding customer with code %s: %v", customerCode, err)
			continue
		}

		query = bson.M{
			"customer_code":           customerCode,
			"bdate":                   bson.M{"$gte": cutOffDate},
			"booking_info_updated_on": bson.M{"$exists": false},
		}

		cursor, err := db.Collection(collectionName).Find(context.Background(), query)
		if err != nil {
			log.Printf("Error finding documents for customer %s: %v", customerCode, err)
			continue
		}
		defer cursor.Close(context.Background())

		var data []bson.M
		if err := cursor.All(context.Background(), &data); err != nil {
			log.Printf("Error reading cursor for customer %s: %v", customerCode, err)
			continue
		}
		totalDocs = len(data)
		log.Printf("Customer %s: Found %d documents", customerCode, totalDocs)

		successCount = 0
		for idx, doc := range data {
			param := getBookingUpdateUrl(doc)
			if param != nil {
				docIndex = idx + 1
				log.Printf("%d/%d - Processing CNo: %v", docIndex, totalDocs, doc["cno"])
				bookingUpdateUrl, ok := configration["bookingUpdateUrl"].(string)
				if !ok {
					log.Printf("Invalid bookingUpdateUrl in configuration")
					continue
				}

				result := api_request.SendData("Booking Info", bookingUpdateUrl, param, doc)
				if result["success"] == true {
					updateResult, err := db.Collection(collectionName).UpdateOne(
						context.Background(),
						bson.M{"_id": doc["_id"]},
						bson.M{"$set": bson.M{"booking_info_updated_on": time.Now()}},
					)
					if err != nil || updateResult.ModifiedCount != 1 {
						log.Printf("Failed to update document CNo: %v", doc["cno"])
					} else {
						successCount++
						log.Printf("%d/%d - CNo: %v updated successfully", docIndex, totalDocs, doc["cno"])
					}
				}
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
		if _, err := db.Collection("main_server_update").InsertOne(context.Background(), bson.M(stat)); err != nil {
			log.Printf("Failed to insert stats for customer %s: %v", customerCode, err)
		}
	}
	return nil
}

func getBookingUpdateUrl(r bson.M) map[string]interface{} {
	configration := config.Config
	cno, ok := r["cno"].(string)
	if !ok {
		log.Printf("Invalid CNo in record: %v", r)
		return nil
	}

	dtStr, ok := r["bdate"].(string)
	if !ok {
		log.Printf("Invalid created_on field in record: %v", r)
		return nil
	}

	dt, err := time.Parse(time.RFC3339, dtStr)
	if err != nil {
		log.Printf("Error parsing created_on for CNo %s: %v", cno, err)
		return nil
	}

	sample := map[string]interface{}{
		"POD_NO":    cno,
		"REF_NO":    "",
		"CLIENT":    "TRZPAL101",
		"BDATE":     dt.Format("2006-01-02"),
		"CONSIGNEE": r["consignee_name"],
		"CC":        r["pc_code"],
		"WEIGHT": func() float64 {
			if r["weight"] != nil {
				return r["weight"].(float64)
			}
			return 0.100
		}(),
		"PIECES": func() int {
			if r["pieces"] != nil {
				return r["pieces"].(int)
			}
			return 1
		}(),
		"BILL_REF":    "",
		"AMOUNT":      0,
		"DESTINATION": r["bdes_name"],
		"DESTN":       r["bdes"],
		"ORIGIN":      configration["branchCode"],
		"SENDER_MOB": func() string {
			if r["primary_contact_number"] != nil {
				return r["primary_contact_number"].(string)
			}
			return ""
		}(),
		"SENDER_EMAIL": func() string {
			if r["primary_contact_email"] != nil {
				return r["primary_contact_email"].(string)
			}
			return ""
		}(),
		"CUST_INVOICE":      "",
		"CUST_INVOICEAMT":   "",
		"FLYER_NO":          "",
		"xmlorigin":         configration["branchCode"],
		"id":                configration["loginId"],
		"RECIPIENT_ADDRESS": r["to_address"],
		"RECIPIENT_MOB":     "",
		"PINCODE":           r["pincode"],
	}

	return sample
}
