package cron

import (
	"strings"
	"fmt"
	"strconv"
	"crypto/md5"
	"encoding/hex"
	"time"
)

const (
	second = iota
	minute
	hour
	day
	month
	week
)

var bounds = map[int]struct{ min, max int }{
	second: {0, 59},
	minute: {0, 59},
	hour:   {0, 23},
	day:    {1, 31},
	month:  {1, 12},
	week:   {0, 6},
}

func parse(spec string) (*schedule, error) {
	s := &schedule{}
	var err error

	fields := strings.Fields(spec)
	if len(fields) != 6 {
		return s, fmt.Errorf("Schedule fields must have six componets %d found\n", len(fields))
	}

	if s.second, err = parseField(fields[second], bounds[second].min, bounds[second].max); err != nil {
		return s, err
	}

	if s.minute, err = parseField(fields[minute], bounds[second].min, bounds[second].max); err != nil {
		return s, err
	}

	if s.hour, err = parseField(fields[hour], bounds[second].min, bounds[second].max); err != nil {
		return s, err
	}

	if s.day, err = parseField(fields[day], bounds[second].min, bounds[second].max); err != nil {
		return s, err
	}

	if s.month, err = parseField(fields[month], bounds[second].min, bounds[second].max); err != nil {
		return s, err
	}

	if s.week, err = parseField(fields[week], bounds[second].min, bounds[second].max); err != nil {
		return s, err
	}
	switch {
	case len(s.day) < 31 && len(s.week) == 7:
		s.week = make(map[int]struct{})
	case len(s.day) == 31 && len(s.week) < 7:
		s.day = make(map[int]struct{})
	default:

	}
	return s, nil
}

func parseField(field string, min, max int) (map[int]struct{}, error) {
	schedule := make(map[int]struct{})
	if field == "*" || field == "?" {
		for i := min; i <= max; i++ {
			schedule[i] = struct{}{}
		}
		return schedule, nil
	}
	rangeAndStep := strings.Split(field, "/")
	lowAndHigh := strings.Split(rangeAndStep[0], "-")
	//example 1-30/1 * * * * * | */1 * * * * *
	//1-30/1
	if len(rangeAndStep) > 1 {
		var err error
		low := min
		high := max

		if rangeAndStep[0] != "" && rangeAndStep[0] != "*" && rangeAndStep[0] != "?" {
			//1-30
			if len(lowAndHigh) > 1 {
				if low, err = mustParseInt(lowAndHigh[0]); err != nil {
					return schedule, err
				}

				if high, err = mustParseInt(lowAndHigh[1]); err != nil {
					return schedule, err
				}

				//1-999/1 * * * * *
				//999 is illegal
				if low < min || high > max {
					return nil, fmt.Errorf("Field %s need[%d-%d] but [%d-%d] give\n", field, min, max, low, high)
				}
			} else {
				//999/1 * * * * *
				//format error
				return nil, fmt.Errorf("Field %s part of %s need format * or %d-%d\n", field, lowAndHigh[0], min, max)
			}
		}

		step, err := mustParseInt(rangeAndStep[1])
		if err != nil {
			return schedule, err
		}

		if step < min || step > high {
			return nil, fmt.Errorf("Field %s part of %s need [%d-%d] but %d give\n", field, rangeAndStep[1], min, max, step)
		}

		for i := low; i <= high; i += step {
			schedule[i] = struct{}{}
		}
		return schedule, nil
	}

	//example */1 * 1,2,16-23 * * *
	//day:1,2,16-23
	parts := strings.Split(field, ",")
	for _, part := range parts {
		//part match 16-23
		rangePart := strings.Split(part, "-")
		if len(rangePart) > 1 {
			low, err := mustParseInt(rangePart[0])
			if err != nil {
				return schedule, err
			}

			high, err := mustParseInt(rangePart[1])
			if err != nil {
				return schedule, err
			}

			if low > high {
				return nil, fmt.Errorf("Field %s part of %s is illegal that %d is greater than %d\n", field, rangePart, low, high)
			}

			if low < min || high > max {
				return nil, fmt.Errorf("Field %s part of %s need[%d-%d] but [%d-%d] give\n", field, rangePart, min, max, low, high)
			}

			for i := low; i <= high; i++ {
				schedule[i] = struct{}{}
			}
		} else {
			//part match 1,2
			i, err := mustParseInt(part)
			if err != nil {
				return schedule, err
			}
			if i < min || i > max {
				return nil, fmt.Errorf("Field %s need[%d-%d] but %d give\n", field, min, max, i)
			}
			schedule[i] = struct{}{}
		}
	}

	if len(schedule) == 0 {
		return nil, fmt.Errorf("Failed to parse [%s]\n", field)
	}
	return schedule, nil
}

func mustParseInt(expr string) (int, error) {
	num, err := strconv.Atoi(expr)
	if err != nil {
		fmt.Println(expr)
		return 0, fmt.Errorf("Failed to parse int from %s: %s\n", expr, err)
	}
	if num < 0 {
		return 0, fmt.Errorf("Negative number (%d) not allowed: %s\n", num, expr)
	}

	return num, nil
}

func UUID() string {
	h := md5.New()
	h.Write([]byte(time.Now().String()))
	return hex.EncodeToString(h.Sum(nil))
}
