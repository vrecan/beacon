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
	Pids []PROC.Process
}

func NewProcess() Process {
	p := Process{
		Life: L.NewLife(),
		l:    log.New(log.Ctx{"module": "beacon/process"}),
	}
	p.Life.SetRun(p.run)
	return p
}

func (p *Process) removeInvalidPids(pids []int32) {
	var idsToRemove []int
	for i, oldPid := range p.Pids {
		var exists bool
		for _, pid := range pids {
			if oldPid.Pid == pid {
				exists = true
			}
			if !exists {
				n, _ := oldPid.Name()
				p.l.Info("pids", log.Ctx{"removing": n})
				idsToRemove = append(idsToRemove, i)
				//remove items that don't exist from the old pids
				// p.Pids[i] = p.Pids[len(p.Pids)-1]
				// p.Pids = p.Pids[:len(p.Pids)-1]
			}
		}
	}
	for _, id := range p.Pids {

	}
}

func (p *Process) gatherAllCurrentPidStats() {
	pids, err := PROC.Pids()
	if nil != err {
		p.l.Error("pids", err)
	}
	p.l.Info("pids", log.Ctx{"len": len(p.Pids)})
	p.removeInvalidPids(pids)

	for _, pid := range pids {
		ps, err := PROC.NewProcess(pid)
		if nil != err {
			p.l.Error("pids", err)
			continue
		}
		p.Pids = append(p.Pids, *ps)
		// name, err := ps.Name()
		// if nil != err {
		// p.l.Error("pids", err)
		// }
		// fmt.Println("NAME: ", name)
		// p.l.Info("pids", log.Ctx{"name": name})
	}
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
