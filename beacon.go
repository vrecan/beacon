package main

import (
	log "github.com/inconshreveable/log15"
	SYS "syscall"

	proc "github.com/vrecan/beacon/process"
	summary "github.com/vrecan/beacon/summary"
	"github.com/vrecan/death"
)

func main() {
	mlog := log.New(log.Ctx{"module": "beacon/main"})
	mlog.Info("starting", log.Ctx{"event": "start"})
	death := death.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	reportChannel := make(chan interface{}, 1000)
	p := proc.NewProcess(reportChannel)
	p.Start()
	s := summary.NewSummary(reportChannel)
	s.Start()
	death.WaitForDeath(p, s)
	mlog.Info("exiting", log.Ctx{"event": "close"})
}
