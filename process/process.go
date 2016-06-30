package process

import (
	log "github.com/inconshreveable/log15"
	PROC "github.com/shirou/gopsutil/process"
	L "github.com/vrecan/life"
	"time"
)

type Process struct {
	*L.Life
	l      log.Logger
	Pids   map[int32]PROC.Process
	Output chan interface{}
}

type ProcessInfo interface {
	Adds() int
	Deletes() int
	Total() int
}

type Info struct {
	adds    int
	deletes int
	total   int
}

func (i Info) Adds() int {
	return i.Adds()
}

func (i Info) Deletes() int {
	return i.Deletes()
}

func (i Info) Total() int {
	return i.Total()
}

func NewProcess(output chan interface{}) Process {
	p := Process{
		Life:   L.NewLife(),
		l:      log.New(log.Ctx{"module": "beacon/process"}),
		Pids:   make(map[int32]PROC.Process, 0),
		Output: output,
	}
	p.Life.SetRun(p.run)
	return p
}

func (p Process) removeDeadPids(pids []int32) int {
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
	return deletes
}

func (p Process) AddNewPids(pids []int32) int {
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
	return adds
}

func (p Process) gatherAllCurrentPidStats() {
	info := &Info{}
	pids, err := PROC.Pids()
	if nil != err {
		p.l.Error("pids", err)
	}
	firstRun := false
	if len(p.Pids) == 0 {
		firstRun = true
	}
	info.adds = p.AddNewPids(pids)
	info.deletes = p.removeDeadPids(pids)
	if firstRun {
		info.adds = 0
	}
	info.total = len(p.Pids)
	p.Output <- info
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
