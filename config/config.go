package config

import (
	"github.com/joho/godotenv"
	"os"
	"log"
)

var Config map[string]interface{}

func init() {
	err := godotenv.Load(".env")
	if err != nil {
	  log.Fatalf("Error loading .env file")
	}

	Base_URL := os.Getenv("BASE_URL")
	Config = map[string]interface{}{
		"baseUrl":          Base_URL,
		"drsTdUrl":         Base_URL + "DRSTD.aspx?",
		"drsUrl":           Base_URL + "DRS.aspx?",
		"ptpUrl":           Base_URL + "iom.aspx?",
		"bookingUpdateUrl": Base_URL + "refdata.aspx",
		"manifestUrl":      Base_URL + "mft.aspx?",
		"branchCode":       "TRZ",
		"backLogDays":      3,
		"loginId":          5024,
	}
}
