// (c) 2017 Rick Arnold. Licensed under the BSD license (see LICENSE).

package cal

import (
	"time"
)

// Common holidays
var (
	NewYear      = NewHoliday(time.January, 1)
	GoodFriday   = NewHolidayFunc(calculateGoodFriday)
	EasterMonday = NewHolidayFunc(calculateEasterMonday)
	Christmas    = NewHoliday(time.December, 25)
	Christmas2   = NewHoliday(time.December, 26)
)

// European Central Bank Target2 holidays
var (
	ECBGoodFriday       = GoodFriday
	ECBEasterMonday     = EasterMonday
	ECBNewYearsDay      = NewYear
	ECBLabourDay        = NewHoliday(time.May, 1)
	ECBChristmasDay     = Christmas
	ECBChristmasHoliday = Christmas2
)

// AddEcbHolidays adds all ECB Target2 holidays to the calendar
func AddEcbHolidays(c *Calendar) {
	c.AddHoliday(
		ECBGoodFriday,
		ECBEasterMonday,
		ECBNewYearsDay,
		ECBLabourDay,
		ECBChristmasDay,
		ECBChristmasHoliday,
	)
}

// British holidays
var (
	GBNewYear       = NewHolidayFunc(calculateNewYearsHoliday)
	GBGoodFriday    = GoodFriday
	GBEasterMonday  = EasterMonday
	GBEarlyMay      = NewHolidayFloat(time.May, time.Monday, 1)
	GBSpringHoliday = NewHolidayFloat(time.May, time.Monday, -1)
	GBSummerHoliday = NewHolidayFloat(time.August, time.Monday, -1)
	GBChristmasDay  = Christmas
	GBBoxingDay     = Christmas2
)

// AddBritishHolidays adds all British holidays to the Calender
func AddBritishHolidays(c *Calendar) {
	c.AddHoliday(
		GBNewYear,
		GBGoodFriday,
		GBEasterMonday,
		GBEarlyMay,
		GBSpringHoliday,
		GBSummerHoliday,
		GBChristmasDay,
		GBBoxingDay,
	)
}

// Dutch holidays
var (
	NLNieuwjaar       = NewYear
	NLGoedeVrijdag    = GoodFriday
	NLPaasMaandag     = EasterMonday
	NLKoningsDag      = NewHolidayFunc(calculateKoningsDag)
	NLBevrijdingsDag  = NewHoliday(time.May, 5)
	NLHemelvaart      = DEChristiHimmelfahrt
	NLPinksterMaandag = DEPfingstmontag
	NLEersteKerstdag  = Christmas
	NLTweedeKerstdag  = Christmas2
)

// AddDutchHolidays adds all Dutch holidays to the Calendar
func AddDutchHolidays(c *Calendar) {
	c.AddHoliday(
		NLNieuwjaar,
		NLGoedeVrijdag,
		NLPaasMaandag,
		NLKoningsDag,
		NLBevrijdingsDag,
		NLHemelvaart,
		NLPinksterMaandag,
		NLEersteKerstdag,
		NLTweedeKerstdag,
	)
}

// US holidays
var (
	USNewYear      = NewYear
	USMLK          = NewHolidayFloat(time.January, time.Monday, 3)
	USPresidents   = NewHolidayFloat(time.February, time.Monday, 3)
	USMemorial     = NewHolidayFloat(time.May, time.Monday, -1)
	USIndependence = NewHoliday(time.July, 4)
	USLabor        = NewHolidayFloat(time.September, time.Monday, 1)
	USColumbus     = NewHolidayFloat(time.October, time.Monday, 2)
	USVeterans     = NewHoliday(time.November, 11)
	USThanksgiving = NewHolidayFloat(time.November, time.Thursday, 4)
	USChristmas    = Christmas
)

// AddUsHolidays adds all US holidays to the Calendar
func AddUsHolidays(cal *Calendar) {
	cal.AddHoliday(
		USNewYear,
		USMLK,
		USPresidents,
		USMemorial,
		USIndependence,
		USLabor,
		USColumbus,
		USVeterans,
		USThanksgiving,
		USChristmas,
	)
}

func calculateEaster(year int, loc *time.Location) time.Time {
	// Meeus/Jones/Butcher algorithm
	y := year
	a := y % 19
	b := y / 100
	c := y % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451

	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
}

func calculateGoodFriday(year int, loc *time.Location) (time.Month, int) {
	easter := calculateEaster(year, loc)
	// two days before Easter Sunday
	gf := easter.AddDate(0, 0, -2)
	return gf.Month(), gf.Day()
}

func calculateEasterMonday(year int, loc *time.Location) (time.Month, int) {
	easter := calculateEaster(year, loc)
	// the day after Easter Sunday
	em := easter.AddDate(0, 0, +1)
	return em.Month(), em.Day()
}

//KoningsDag (kingsday) is April 27th, 26th if the 27th is a Sunday
func calculateKoningsDag(year int, loc *time.Location) (time.Month, int) {
	koningsDag := time.Date(year, time.April, 27, 0, 0, 0, 0, loc)
	if koningsDag.Weekday() == time.Sunday {
		koningsDag = koningsDag.AddDate(0, 0, -1)
	}
	return koningsDag.Month(), koningsDag.Day()
}

// NewYearsDay is the 1st of January unless the 1st is a Saturday or Sunday
// in which case it occurs on the following Monday.
func calculateNewYearsHoliday(year int, loc *time.Location) (time.Month, int) {
	day := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	switch day.Weekday() {
	case time.Saturday:
		day = day.AddDate(0, 0, 2)
	case time.Sunday:
		day = day.AddDate(0, 0, 1)
	}
	return time.January, day.Day()
}
