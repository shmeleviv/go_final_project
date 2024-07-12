package datetime

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	DateFormat = "20060102"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {

	repeatSlice := strings.Split(repeat, " ")
	if repeatSlice[0] == "d" && len(repeatSlice) < 2 {
		return "incorrect d repeat format", errors.New("incorrect d repeat format")
	}
	if repeatSlice[0] == "y" && len(repeatSlice) > 1 {
		return "incorrect y repeat format", errors.New("incorrect y repeat format")
	}
	switch repeatSlice[0] {
	case "d":
		switch repeatSlice[1] {
		default:
			dayCount, err := strconv.Atoi(repeatSlice[1])
			if err != nil {
				return "Can't pasre interger", err
			}
			if dayCount >= 1 && dayCount <= 400 {

				if date == now.Format(DateFormat) {
					dateTime, err := time.Parse(DateFormat, date)
					if err != nil {
						return "Incorrect date format", err
					}
					if dayCount == 1 {

						return fmt.Sprintf("%v", dateTime.Format(DateFormat)), nil
					}
					dateTime = dateTime.AddDate(0, 0, dayCount)
					return fmt.Sprintf("%v", dateTime.Format(DateFormat)), nil
				}

				if date > now.Format(DateFormat) {

					dateTime, err := time.Parse(DateFormat, date)
					if err != nil {
						return "Incorrect date format", err
					}
					dateTime = dateTime.AddDate(0, 0, dayCount)
					return fmt.Sprintf("%v", dateTime.Format(DateFormat)), nil
				}

				if date < now.Format(DateFormat) {
					dateTime, err := time.Parse(DateFormat, date)
					if err != nil {
						return "Incorrect date format", err
					}

					for {
						dateTime = dateTime.AddDate(0, 0, dayCount)
						if dateTime.Format(DateFormat) < now.Format(DateFormat) {
							continue
						} else {
							return fmt.Sprintf("%v", dateTime.Format(DateFormat)), nil
						}

					}
				}

			}
			return "Incorrect d format. d must be in rage 1 to 400", errors.New("Incorrect d format. d must be in rage 1 to 400")
		}

	case "y":

		dateTime, err := time.Parse(DateFormat, date)
		if err != nil {
			return "Incorrect date format", err
		}

		if date < now.Format(DateFormat) {
			dateTime, err := time.Parse(DateFormat, date)
			if err != nil {
				return "Incorrect date format", err
			}

			for {
				dateTime = dateTime.AddDate(1, 0, 0)
				if dateTime.Format(DateFormat) < now.Format(DateFormat) {
					continue
				} else {
					return fmt.Sprintf("%v", dateTime.Format(DateFormat)), nil
				}

			}
		}
		dateTime = dateTime.AddDate(1, 0, 0)
		return fmt.Sprintf("%v", dateTime.Format(DateFormat)), nil
	default:

		return "Incorrect Format. Only d or y supporter", errors.New("Incorrect Format. Only d or y supporter")

	}
}
