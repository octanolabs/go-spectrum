package util

import (
	"sort"
	"strconv"
	"testing"
)

func TestDateValuesSliceSorting(t *testing.T) {

	dvs := DateValuesSlice{
		Values: []uint{7, 2, 1, 4, 5, 3, 6},
		Dates:  []string{"01/07/20", "01/02/20", "01/01/20", "01/04/20", "01/05/20", "01/03/20", "01/06/20"},
	}

	t.Log("before sort")
	t.Log("dates: ", dvs.Dates)
	t.Log("values: ", dvs.Values)

	sort.Sort(dvs)

	for k, v := range []uint{1, 2, 3, 4, 5, 6, 7} {

		s := strconv.FormatInt(int64(v), 10)

		if (dvs.Values[k] != v) || (dvs.Dates[k] != "01/0"+s+"/20") {
			t.Error("error: didn't sort properly", k, v, dvs.Values[k], dvs.Dates[k])
		}
	}

	t.Log("sorted")
	t.Log("dates: ", dvs.Dates)
	t.Log("values: ", dvs.Values)

}
