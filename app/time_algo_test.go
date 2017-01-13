// +build ignore

package main

import "time"
import "fmt"
import "gopkg.in/thehowl/go-osuapi.v1"

func main() {
	testTable := []struct {
		rankedStatus int
		lastUpdate   time.Time
	}{
		{1, time.Date(2007, time.December, 23, 22, 10, 01, 0, time.Local)},
		{4, time.Date(2007, time.December, 23, 22, 10, 01, 0, time.Local)},
		{3, time.Date(2007, time.December, 23, 22, 10, 01, 0, time.Local)},
		{3, time.Date(2016, time.December, 23, 22, 10, 01, 0, time.Local)},
		{1, time.Date(2016, time.December, 23, 22, 10, 01, 0, time.Local)},
	}
	for _, el := range testTable {
		d := calc(el.rankedStatus, el.lastUpdate)
		fmt.Printf("%-10s %-13v %-15f days\n", osuapi.ApprovedStatus(el.rankedStatus).String(), d, d.Hours()/24)
	}
}

func calc(rankedStatus int, lastUpdate time.Time) time.Duration {
	var multiplier float64

	switch rankedStatus {
	// ranked, approved
	case 1, 2:
		multiplier = float64(1) / 4 // rarely checked
	// loved
	case 4:
		multiplier = float64(1) // isn't likely to change much, but still more than a ranked or approved
	// qualified
	case 3:
		multiplier = float64(8) // qualified must be checked very often so that we know if it's been ranked
	// pending, wip
	case 0, -1:
		multiplier = float64(4) // must be checked often because they can change quickly
	// graveyard
	case -2:
		multiplier = float64(1) / 6 // really unlikely to change
	}
	return time.Duration(float64(time.Now().Unix()-lastUpdate.Unix())/(20*multiplier)) * time.Second
}
