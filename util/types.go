package util

import (
	"time"
)

type DateValuesSlice struct {
	Values []uint
	Dates  []string
}

func (sbo DateValuesSlice) Len() int {
	return len(sbo.Dates)
}

func (sbo DateValuesSlice) Swap(i, j int) {
	sbo.Dates[i], sbo.Dates[j] = sbo.Dates[j], sbo.Dates[i]
	sbo.Values[i], sbo.Values[j] = sbo.Values[j], sbo.Values[i]
}

func (sbo DateValuesSlice) Less(i, j int) bool {
	ti, _ := time.Parse("01/02/06", sbo.Dates[i])
	tj, _ := time.Parse("01/02/06", sbo.Dates[j])

	return ti.Before(tj)
}
