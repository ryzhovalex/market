package time

import "time"

type Time = int64

func Utc() Time {
	return time.Now().Unix()
}
