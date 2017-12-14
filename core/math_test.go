package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBusinessHoursBetweenTimes(t *testing.T) {
	tt := func(m string) time.Time {
		// businessHoursBetweenTimes() currently uses local time internally,
		// so let's ensure our unit test is in local time.
		// Mon Jan 2 15:04:05 -0700 MST 2006
		r, err := time.ParseInLocation("Mon Jan 2 15:00 2006", m, time.Now().Location())
		if err != nil {
			panic(err)
		}
		return r
	}
	tcs := []struct {
		name             string
		start            time.Time
		end              time.Time
		expectedDuration time.Duration
	}{
		{
			"start and end both on same non-workday",
			tt("Sat May 4 14:00 2006"),
			tt("Sat May 4 15:00 2006"),
			0,
		},
		{
			"end < start, they are swapped",
			tt("Fri May 3 15:00 2006"),
			tt("Fri May 3 14:00 2006"),
			time.Hour,
		},
		{
			"start before business hours, end same day",
			tt("Fri May 3 7:30 2006"),
			tt("Fri May 3 10:00 2006"),
			time.Minute * 30, // 9:30 - 10am are business hours
		},
		/*
			end < start, they are swapped
			start and end
		*/
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedDuration, businessHoursBetweenTimes(tc.start, tc.end))
		})
	}
}
