package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type PerfStats struct {
	start   time.Time
	values  map[string]*stats
	opCnt   int
	ingress int64
	egress  int64
	ch      chan sample
	wg      *sync.WaitGroup
	grafana client.Client
	grafdb  string
}

type stats struct {
	opCnt       int
	totalElapse int64
	max         int64
	min         int64
	egress      int64
	ingress     int64
}

type sample struct {
	op      string
	elapse  int64
	egress  int64
	ingress int64
}

func NewPerfStats(svr, db, user, pass string) (*PerfStats, error) {
	ps := &PerfStats{
		start:  time.Now(),
		values: make(map[string]*stats, 20),
		ch:     make(chan sample, 100),
		wg:     &sync.WaitGroup{},
	}

	// Make client
	for i := 0; i < 5; i++ {
		client, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:     svr,
			Username: user,
			Password: pass,
		})
		if err != nil {
			log.Println("Initialize GrafanaError: ", err)
			time.Sleep(time.Second)
		} else {
			ps.grafana = client
			ps.grafdb = db
			break
		}
	}

	ps.wg.Add(1)

	go ps.run()

	return ps, nil
}

func (p *PerfStats) PostSample(op string, elapsed int64, ingress, egress int64) {
	p.ch <- sample{op, elapsed, ingress, egress}
}

func (p *PerfStats) run() {
	tick := time.Tick(2 * time.Second)
	lastOpCnt := 0

	for {
		select {
		case sa, ok := <-p.ch:
			if !ok {
				p.wg.Done()
				return
			}
			if _, prs := p.values[sa.op]; !prs {
				p.values[sa.op] = &stats{1, sa.elapse, sa.elapse, sa.elapse, sa.egress, sa.ingress}
			} else {
				p.values[sa.op].opCnt += 1
				p.values[sa.op].totalElapse += sa.elapse
				if p.values[sa.op].max < sa.elapse {
					p.values[sa.op].max = sa.elapse
				}
				if p.values[sa.op].min > sa.elapse {
					p.values[sa.op].min = sa.elapse
				}
				p.values[sa.op].egress += sa.egress
				p.values[sa.op].ingress += sa.ingress
			}
			p.opCnt += 1
			p.ingress += sa.ingress
			p.egress += sa.egress

			p.postGrafanaMetric(&sa)

		case <-tick:
			if p.opCnt != lastOpCnt {
				sp1 := int64(float64(p.egress*8) / float64(time.Since(p.start).Seconds()))
				sp2 := int64(float64(p.ingress*8) / float64(time.Since(p.start).Seconds()))
				log.Printf("%d operations, egress %d ( %s), ingress %d ( %s).",
					p.opCnt, p.egress, speedStr(sp1),
					p.ingress, speedStr(sp2))
			}
			lastOpCnt = p.opCnt
		}
	}
}

func (p *PerfStats) Shutdown() {
	close(p.ch)
	p.wg.Wait()

	for op, d := range p.values {
		if d.opCnt != 0 {
			//avgEplase, _ := time.ParseDuration(fmt.Sprintf("%dms", d.totalElapse/int64(d.opCnt)))
			espeed := int64(float64(d.egress*8) / float64(d.totalElapse/1000))
			ispeed := int64(float64(d.ingress*8) / float64(d.totalElapse/1000))
			log.Printf("%s: %d operations, egress %d (%s), ingress %d (%s) ",
				op, d.opCnt, d.egress, speedStr(espeed), d.ingress, speedStr(ispeed))
		}
	}
}

func speedStr(sp int64) string {
	var ut string
	div := int64(1)
	if sp & ^(1<<20-1) == 0 {
		div = 1 << 10
		ut = "Ki"
	} else if sp & ^(1<<30-1) == 0 {
		div = 1 << 20
		ut = "Mi"
	} else if sp & ^(1<<40-1) == 0 {
		div = 1 << 30
		ut = "Gi"
	}

	s := fmt.Sprintf("%.2f", float64(sp/div)) + ut + "b/s"

	return s
}

func (p *PerfStats) postGrafanaMetric(s *sample) {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  p.grafdb,
		Precision: "ns",
	})

	tags := map[string]string{
		"operation": s.op,
	}

	fields := map[string]interface{}{
		"elapse": s.elapse,
	}

	pt, _ := client.NewPoint(
		"Performance",
		tags,
		fields,
		time.Now(),
	)
	bp.AddPoint(pt)

	err := p.grafana.Write(bp)
	if err != nil {
		log.Println("Failed to write metric to grafana: ", err)
	}
}
