package database

import (
	"sort"
	"time"
)

type elem interface {
	Add(interface{})
}

type chartData struct {
	data map[string]elem
}

func (c *chartData) init() {
	c.data = make(map[string]elem)

}

func (c *chartData) addElement(stamp string, element elem) {
	if c.data[stamp] == nil {
		c.data[stamp] = element
	}
	c.data[stamp].Add(element)
}

func (c *chartData) getElement(stamp string) elem {
	return c.data[stamp]
}

func (c *chartData) getDates() []string {
	dates := make([]string, 0)

	for k := range c.data {
		dates = append(dates, k)
	}

	sort.Slice(dates, func(i, j int) bool {
		ti, _ := time.Parse("01/02/06", dates[i])
		tj, _ := time.Parse("01/02/06", dates[j])
		return ti.Before(tj)
	})

	return dates
}
