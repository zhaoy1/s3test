package main

import "fmt"
import _ "errors"

type PerfStats struct {
	data []durationStats
}

type durationStats struct {
	ObjSize  int64
	OpCnt    int
	TotalDur int64
	Max      int64
	Min      int64
}

func NewPerfStats(objSizeList []int64) (*PerfStats, error) {
	d := make([]durationStats, len(objSizeList))
	for i := 0; i < len(objSizeList); i++ {
		d[i].ObjSize = objSizeList[i]
	}
	ps := &PerfStats{
		data: d,
	}

	return ps, nil
}

func (p *PerfStats) Add1Duration(objSize int64, duration int64) error {
	var processed bool
	for i := 0; i < len(p.data); i++ {
		d := &(p.data[i])
		if d.ObjSize == objSize {
			d.OpCnt += 1
			d.TotalDur += duration
			if duration > d.Max {
				d.Max = duration
			}
			if duration < d.Min {
				d.Min = duration
			}
			processed = true
			break
		}
	}

	if !processed {
		fmt.Println("Don't find a slot for a performance metric:", objSize)
	}

	return nil
}

func (p *PerfStats) PrintStats() {
	for _, d := range p.data {
		if d.OpCnt != 0 {
			speed := float64(d.ObjSize*8) / (float64(d.TotalDur) / float64(d.OpCnt) / 1000)
			fmt.Printf("%d: average speed: %.2f, average duration:%dms\n", d.ObjSize, speed, d.TotalDur/int64(d.OpCnt))
			//fmt.Println(d.ObjSize, ":", "average=", speed, "bps ,max=", d.Max, ",min=", d.Min)
		}
	}
}
