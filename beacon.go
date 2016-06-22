package main

import (
	log "github.com/inconshreveable/log15"
	SYS "syscall"

	"github.com/vrecan/death"
)

func main() {
	mlog := log.New(log.Ctx{"module": "beacon/main"})
	mlog.Info("starting", log.Ctx{"event": "start"})
	death := death.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	death.WaitForDeathWithFunc(func() {
		mlog.Info("exiting", log.Ctx{"event": "close"})
	})
}
