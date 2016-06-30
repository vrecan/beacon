package summary

import (
	log "github.com/inconshreveable/log15"
	proc "github.com/vrecan/beacon/process"
	L "github.com/vrecan/life"
)

type Summary struct {
	*L.Life
	reporters chan interface{}
	l         log.Logger
}

func NewSummary(reporters chan interface{}) Summary {
	summary := Summary{Life: L.NewLife(), reporters: reporters, l: log.New(log.Ctx{"module": "beacon/summary"})}
	summary.SetRun(summary.run)
	return summary
}

func (s Summary) run() {
	defer s.WGDone()
Loop:
	for {
		select {
		case record := <-s.reporters:
			switch record.(type) {
			case proc.ProcessInfo:
				s.l.Info("process", "records", record.(proc.ProcessInfo))
			default:
				s.l.Error("process", "unknown type", record)
			}
		case <-s.Done:
			break Loop
		}
	}
}
