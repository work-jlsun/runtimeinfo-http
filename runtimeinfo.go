package rumtimeinfo

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"runtime"
	"sort"
	"strconv"
)

type Uint64Slice []uint64

func (s Uint64Slice) Len() int {
	return len(s)
}

func (s Uint64Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Uint64Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func percentile(perc float64, arr []uint64, length int) uint64 {
	if length == 0 {
		return 0
	}
	indexOfPerc := int(math.Floor(((perc / 100.0) * float64(length)) + 0.5))
	if indexOfPerc >= length {
		indexOfPerc = length - 1
	}
	return arr[indexOfPerc]
}

func average(arr []uint64) uint64 {
	var sum uint64
	if len(arr) > 0 {
		for i := 0; i < len(arr); i++ {
			sum += arr[i]
		}
		return sum / uint64(len(arr))
	} else {
		return 0
	}
}

type GoRunTimeInfo struct {
	//gorotine info
	NumGoroutine int

	// above are the mem infos
	// General statistics.
	Alloc      uint64 // bytes allocated and not yet freed
	TotalAlloc uint64 // bytes allocated (even if freed)
	Sys        uint64 // bytes obtained from system (sum of XxxSys below)
	Lookups    uint64 // number of pointer lookups
	Mallocs    uint64 // number of mallocs
	Frees      uint64 // number of frees

	// Main allocation heap statistics.
	HeapAlloc    uint64 // bytes allocated and not yet freed (same as Alloc above)
	HeapSys      uint64 // bytes obtained from system
	HeapIdle     uint64 // bytes in idle spans
	HeapInuse    uint64 // bytes in non-idle span
	HeapReleased uint64 // bytes released to the OS
	HeapObjects  uint64 // total number of allocated object

	// Low-level fixed-size structure allocator statistics.
	// Inuse is bytes used now.
	// Sys is bytes obtained from system.
	StackInuse  uint64 // bytes used by stack allocator
	StackSys    uint64
	MSpanInuse  uint64 // mspan structures
	MSpanSys    uint64
	MCacheInuse uint64 // mcache structures
	MCacheSys   uint64
	BuckHashSys uint64 // profiling bucket hash table
	GCSys       uint64 // GC metadata
	OtherSys    uint64 // other system allocations

	// Garbage collector statistics.
	NextGC       uint64 // next collection will happen when HeapAlloc ≥ this amount
	LastGC       uint64 // end time of last collection (nanoseconds since 1970)
	PauseTotalNs uint64
	//PauseNs       [256]uint64 // circular buffer of recent GC pause durations, most recent at [(NumGC+255)%256]
	//PauseEnd      [256]uint64 // circular buffer of recent GC pause end times
	NumGC         uint32
	GCCPUFraction float64 // fraction of CPU time used by GC

	// self calculate gc info
	AvgGCTimeUsec  uint64
	GCPauseUsec100 uint64
	GCPauseUsec99  uint64
	GCPauseUsec95  uint64
	GCPauseUsec90  uint64
	GCPauseUsec80  uint64
	GCPauseUsec70  uint64
	GCPauseUsec60  uint64
	GCPauseUsec50  uint64
}

func (s *GoRunTimeInfo) SetRunTimeInfos(memstat runtime.MemStats) {
	s.NumGoroutine = runtime.NumGoroutine()

	// sort the GC pause array
	length := 256
	if len(memstat.PauseNs) <= 256 {
		length = int(memstat.NumGC)
	}

	gcPauses := make(Uint64Slice, length)
	copy(gcPauses, memstat.PauseNs[:])
	sort.Sort(gcPauses)

	s.Alloc = memstat.Alloc
	s.TotalAlloc = memstat.TotalAlloc
	s.Sys = memstat.Sys
	s.Lookups = memstat.Lookups
	s.Mallocs = memstat.Mallocs
	s.Frees = memstat.Frees

	s.HeapAlloc = memstat.HeapAlloc
	s.HeapSys = memstat.HeapSys
	s.HeapIdle = memstat.HeapIdle
	s.HeapInuse = memstat.HeapInuse
	s.HeapReleased = memstat.HeapReleased
	s.HeapObjects = memstat.HeapObjects

	s.StackInuse = memstat.StackInuse
	s.StackSys = memstat.StackSys
	s.MSpanInuse = memstat.MSpanInuse
	s.MSpanSys = memstat.MSpanSys
	s.MCacheInuse = memstat.MCacheInuse
	s.MCacheSys = memstat.MCacheSys
	s.BuckHashSys = memstat.BuckHashSys
	s.GCSys = memstat.GCSys
	s.OtherSys = memstat.OtherSys

	s.NextGC = memstat.NextGC
	s.LastGC = memstat.LastGC
	s.PauseTotalNs = memstat.PauseTotalNs
	s.NumGC = memstat.NumGC
	//s.GCCPUFraction = memstat.GCCPUFraction //need by go release  1.5+

	// slef calculate gc info
	s.GCPauseUsec100 = percentile(100.0, gcPauses, len(gcPauses)) / 1000
	s.GCPauseUsec99 = percentile(99.0, gcPauses, len(gcPauses)) / 1000
	s.GCPauseUsec95 = percentile(95.0, gcPauses, len(gcPauses)) / 1000
	s.GCPauseUsec90 = percentile(90.0, gcPauses, len(gcPauses)) / 1000
	s.GCPauseUsec80 = percentile(80.0, gcPauses, len(gcPauses)) / 1000
	s.GCPauseUsec70 = percentile(70.0, gcPauses, len(gcPauses)) / 1000
	s.GCPauseUsec60 = percentile(60.0, gcPauses, len(gcPauses)) / 1000
	s.GCPauseUsec50 = percentile(50.0, gcPauses, len(gcPauses)) / 1000

	s.AvgGCTimeUsec = average(gcPauses) / 1000
}

const (
	CONTENT_LENGTH    = "Content-Length"
	CONTENT_TYPE      = "Content-Type"
	JSON_CONTENT_TYPE = "application/json;charset=UTF-8"
)

func ServeRuntimeHTTPInfo(w http.ResponseWriter, request *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	//log.Printf("memstats = %+v\n", memStats)
	var runtimeinfo GoRunTimeInfo
	runtimeinfo.SetRunTimeInfos(memStats)

	body, err := json.Marshal(runtimeinfo)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Json Marshal error"))
	} else {
		w.Header().Add(CONTENT_TYPE, JSON_CONTENT_TYPE)
		w.Header().Add(CONTENT_LENGTH, strconv.Itoa(len(body)))
		w.WriteHeader(200)
		w.Write(body)
	}
}

func ListenAndServeRunTimeInfo(listenPort string) {
	//start the httpserver
	http.HandleFunc("/", ServeRuntimeHTTPInfo)
	log.Fatal(http.ListenAndServe(":"+listenPort, nil))
}
