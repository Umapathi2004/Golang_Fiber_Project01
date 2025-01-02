package main

import (
	"GoFiber_Project01/services"
	"fmt"
	"os"
)

func main() {
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
		services.In_mft()
	default:
		fmt.Println("Invalid Service")
	}
}
