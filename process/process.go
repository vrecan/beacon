package process

import (
	// "fmt"
	log "github.com/inconshreveable/log15"
	PROC "github.com/shirou/gopsutil/process"
	L "github.com/vrecan/life"
	"time"
)

type Process struct {
	*L.Life
	l    log.Logger
	Pids map[int32]PROC.Process
}

func NewProcess() Process {
	p := Process{
		Life: L.NewLife(),
		l:    log.New(log.Ctx{"module": "beacon/process"}),
		Pids: make(map[int32]PROC.Process, 0),
	}
	p.Life.SetRun(p.run)
	return p
}

func (p *Process) gatherAllCurrentPidStats() {
	pids, err := PROC.Pids()
	if nil != err {
		p.l.Error("pids", err)
	}
	p.l.Info("pids", log.Ctx{"len": len(p.Pids)})
	adds := 0
	for _, pid := range pids {
		_, exist := p.Pids[pid]
		if !exist {
			adds++
			ps, err := PROC.NewProcess(pid)
			if nil != err {
				p.l.Error("pids", err)
				continue
			}
			p.Pids[pid] = *ps
		}
	}
	//remove dead pids and get a count
	deletes := 0
	for pid, _ := range p.Pids {
		exists := false
		for _, npid := range pids {
			if pid == npid {
				exists = true
			}
		}
		if !exists {
			delete(p.Pids, pid)
			deletes++
		}
	}
	p.l.Info("pids", log.Ctx{"deleted": deletes, "added": adds})
}

func (p Process) run() {
	defer p.Life.WGDone()
	tick := time.NewTicker(5 * time.Second)
Loop:
	for {
		select {
		case <-tick.C:
			p.gatherAllCurrentPidStats()
		case <-p.Life.Done:
			break Loop
		}

	}
}
