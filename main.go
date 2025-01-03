package main

import (
	"GoFiber_Project01/logs"
	"GoFiber_Project01/services"
	"fmt"
	"os"
)

func main() {
	logs.Logger()
	Service := os.Args[1]

	switch Service {
	case "bookingUpdate":
		services.UpdateBookingConsignment()
	case "ptp":
		services.UpdatePTP()
	case "ptp_maa":
		services.UpdatePTP_MAA()
	case "drs_td":
		services.UpdateDrsTd()
	case "drs":
		services.UpdateDRS()
	case "drsbooking":
		services.UpdateBookingDRS()
	case "drs_temp":
		services.UpdateDRSTemp()
	case "in_mft":
		services.UpdateInCommingManifest()
	case "out_mft":
		services.UpdateOutCommingManifest()
	case "out_mft_fix":
		services.Fixes()
	default:
		fmt.Printf("Invalid Service %v\n", Service)
		logs.ErrorLog.Printf("Invalid Service %v\n", Service)
	}
}
