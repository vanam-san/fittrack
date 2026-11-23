package calendar

import "time"

type DateError string

func (e DateError) Error() string { return string(e) }

func Parse(date string) (time.Time, error) {
	day, err := time.Parse(time.DateOnly, date)
	if err != nil || day.Year() < 1 || day.Format(time.DateOnly) != date {
		return time.Time{}, DateError("Choose a valid calendar date.")
	}
	return day, nil
}

func Range(from, to string) (time.Time, time.Time, error) {
	start, startErr := Parse(from)
	end, endErr := Parse(to)
	if startErr != nil || endErr != nil {
		return time.Time{}, time.Time{}, DateError("Choose valid start and end dates.")
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, DateError("Choose an end date on or after the start date.")
	}
	return start, end, nil
}
