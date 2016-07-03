package process

import (
	"github.com/asdine/storm"
	log "github.com/inconshreveable/log15"
	PROC "github.com/shirou/gopsutil/process"
	L "github.com/vrecan/life"
	"runtime"
	"time"
)

var empty = struct{}{}

type Process struct {
	*L.Life
	l      log.Logger
	Pids   map[int32]*Service
	Output chan interface{}
	DB     *storm.DB
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
		Pids:   make(map[int32]*Service, 0),
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

func (p Process) addNewPids(pids []int32) int {
	adds := 0
	for _, pid := range pids {
		_, exist := p.Pids[pid]
		if !exist {
			adds++
			ps, err := PROC.NewProcess(pid)
			if nil != err {
				p.l.Error("pids", "newProcess", err)
				continue
			}
			p.Pids[pid], err = NewService(ps)
			if err != nil {
				p.l.Error("pids", "newService", err)
			}
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
	info.adds = p.addNewPids(pids)
	info.deletes = p.removeDeadPids(pids)
	if firstRun {
		info.adds = 0
	}
	info.total = len(p.Pids)
	go p.gatherServiceStats()
	p.Output <- info
}

//Service is a process and it's collection of stats
type Service struct {
	Name           string
	User           string
	FD             int32
	ResidentMemory uint64
	CPUPercent     float64
	interval       int64

	memInfo *PROC.MemoryInfoStat
	process *PROC.Process
}

func (s *Service) UpdateStats() (err error) {
	s.interval++
	//sample memory because it's expensive to get
	if s.interval == 1 || s.interval%5 == 0 {
		s.memInfo, err = s.process.MemoryInfo()
		if err != nil {
			return err
		}
		s.ResidentMemory = s.memInfo.RSS
	}
	if runtime.GOOS != "darwin" {
		if s.interval == 1 || s.interval%5 == 0 {
			s.CPUPercent, err = s.process.Percent(0)
			if err != nil {
				return err
			}
		}
		s.FD, err = s.process.NumFDs()
		if err != nil {
			return err
		}
	}
	return nil
}

func NewService(p *PROC.Process) (s *Service, err error) {
	s = &Service{}
	s.process = p
	cmd, err := p.CmdlineSlice()
	if err != nil {
		return s, err
	}
	if len(cmd) > 0 {
		s.Name = string(cmd[0])
	}
	s.User, err = p.Username()
	if err != nil {
		return s, err
	}
	s.UpdateStats()
	return s, err

}

func (p Process) gatherServiceStats() {
	log.Info("Calling update stats on all process")
	for i, pid := range p.Pids {
		p.Pids[i].UpdateStats()
		//debug output of our process
		if pid.Name == "./beacon" {
			log.Info("service", "info", pid)
		}

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
