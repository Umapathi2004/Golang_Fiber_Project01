package helpers

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func StringToDateConverter(date interface{}) (bool, time.Time) {
	switch v := date.(type) {
	case time.Time:
		// Input is already a time.Time, return it as is
		return false, v
	case string:
		// Attempt to parse the string as an RFC3339 date
		if parsedDate, err := time.Parse(time.RFC3339, v); err == nil {
			return false, parsedDate
		}
	case primitive.DateTime:
		// Input is already a time.Time, return it as is
		return false, v.Time()
	}

	// Return the zero value of time.Time and true if the conversion failed
	return true, time.Time{}
}

// func ToInt64(s interface{}) int64 {
// 	switch v := s.(type) {
// 	case int:
// 		return int64(v)
// 	case int8:
// 		return int64(v)
// 	case int16:
// 		return int64(v)
// 	case int32:
// 		return int64(v)
// 	case int64:
// 		return int64(v)
// 	case float32:
// 		return int64(v)
// 	case float64:
// 		return int64(v)
// 	case string:
// 		if v == "" {
// 			return 0
// 		}
// 		val, err := strconv.ParseFloat(v, 64)
// 		if err != nil {
// 			return 0
// 		}
// 		return int64(val)
// 	default:
// 		return 0
// 	}
// }
