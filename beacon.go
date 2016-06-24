package main

import (
	log "github.com/inconshreveable/log15"
	SYS "syscall"

	proc "github.com/vrecan/beacon/process"
	"github.com/vrecan/death"
)

func main() {
	mlog := log.New(log.Ctx{"module": "beacon/main"})
	mlog.Info("starting", log.Ctx{"event": "start"})
	p := proc.NewProcess()
	p.Start()
	death := death.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	death.WaitForDeath(p)
	mlog.Info("exiting", log.Ctx{"event": "close"})
}
