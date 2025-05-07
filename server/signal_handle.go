package server

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/diceclone/core"
)

const EngineStatus_WAITING int32 = 1 << 1
const EngineStatus_BUSY int32 = 1 << 2
const EngineStatus_SHUTTING_DOWN = 1 << 3

var eStatus int32 = EngineStatus_WAITING

func WaitForSignal(wg *sync.WaitGroup, sig chan os.Signal) {
	defer wg.Done()

	// signal is received only when the process is in waiting state
	// before proceeding to receiving signal, change the process state to shutdown, so that no client can send request

	s := <-sig
	fmt.Printf("%v received.graceful shutdown activated...\n", s.String())

	for atomic.LoadInt32(&eStatus) == EngineStatus_BUSY {
	}

	// server should not go back to BUSY state, hence update the status to shutdown
	atomic.StoreInt32(&eStatus, EngineStatus_SHUTTING_DOWN)

	core.Shutdown()
	os.Exit(0)
}
