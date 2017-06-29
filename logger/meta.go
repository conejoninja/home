package logger

import (
	"strconv"
	"time"

	"github.com/conejoninja/home/common"
)

func CalculateMeta(sensor string, start, end time.Time, prefix string) {

	values := db.GetValuesBetweenTime(sensor, start, end)

	var meta common.Meta

	if len(values) > 0 {
		defVal := values[0]
		if defVal.Type == "" || defVal.Type == "number" {
			val, _ := common.GetFloat(values[0].Value)
			meta.Max = val
			meta.Min = val
			var tmpVal float64
			for _, value := range values {
				val, _ := common.GetFloat(value.Value)
				if val > meta.Max {
					meta.Max = val
				}
				if val < meta.Min {
					meta.Min = val
				}
				tmpVal += val
				meta.N++
			}
			meta.Avg = tmpVal / float64(meta.N)
		}
		db.AddMeta([]byte(sensor+"-"+prefix+strconv.Itoa(int(start.Unix()))), meta)
	}
}

func CalculateMetaHour(sensor string, start time.Time) {
	start = time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), 0, 0, 0, time.UTC)
	end := start.Add(1 * time.Hour)
	CalculateMeta(sensor, start, end, "hour-")
}

func CalculateMetaDay(sensor string, start time.Time) {
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	CalculateMeta(sensor, start, end, "day-")
}

func CalculateMetaWeek(sensor string, start time.Time) {
	weekday := int(start.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekday = weekday - 1
	start = start.Add(-1 * time.Duration(weekday) * 24 * time.Hour)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(7 * 24 * time.Hour)
	CalculateMeta(sensor, start, end, "week-")
}

func CalculateMetaMonth(sensor string, start time.Time) {
	start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(start.Year(), start.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	end = start.Add(-1 * time.Second)
	CalculateMeta(sensor, start, end, "hourly-")
}

func CalculateMetaAll(sensor string, start time.Time) {
	CalculateMetaHour(sensor, start)
	CalculateMetaDay(sensor, start)
	CalculateMetaWeek(sensor, start)
	CalculateMetaMonth(sensor, start)
}
