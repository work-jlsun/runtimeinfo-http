package rumtimeinfo

import (
	"sort"
	"testing"
)

func TestAverage(t *testing.T) {
	arr := []uint64{1, 1, 1}
	if average(arr) != uint64(1) {
		t.Fatal("average func error")
	}
	arr = []uint64{1, 1, 4, 10, 20}
	if average(arr) != uint64(7) {
		t.Fatal("average func error")
	}
}

func TestPercentile(t *testing.T) {
	arr := []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	ifOk := true

	if percentile(100.0, arr, len(arr)) != 9 {
		ifOk = false
	}

	if percentile(90.0, arr, len(arr)) != 9 {
		ifOk = false
	}

	if percentile(50.0, arr, len(arr)) != 5 {
		ifOk = false
	}

	if percentile(10.0, arr, len(arr)) != 1 {
		ifOk = false
	}
	if ifOk != true {

		t.Fatal("percentile test test")
	}
}

func TestSort(t *testing.T) {
	arr := []uint64{0, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	sortArr := make(Uint64Slice, len(arr))
	copy(sortArr, arr[:])
	sort.Sort(sortArr)
	for i := 0; i != len(sortArr); i++ {
		if sortArr[i] != uint64(i) {
			t.Fatal("sort error")
		}
	}
}

/*
//test it
func TestHttp(t *testing.T) {
	ListenAndServeRunTimeInfo("7070")
}
*/
