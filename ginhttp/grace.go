package ginhttp

import (
	"net/http"
	"runtime/debug"

	"github.com/jpillora/overseer"
	"github.com/spf13/cast"

	logger "github.com/senpan/xlogger"
)

var server = &http.Server{}
var addresses = make([]string, 0)

func graceStart(addr string, s *http.Server) {
	addresses = append(addresses, addr)
	server = s
	oversee()
}

func oversee() {
	overseer.Run(overseer.Config{
		Program:   prog,
		Addresses: addresses,
		Debug:     true, // display log of overseer actions
	})

}

func prog(state overseer.State) {
	tag := "xserver.GinServer.Program"
	logger.I(tag, "app (%s) listening...", state.ID)
	err := server.Serve(state.Listener)
	if err != nil {
		logger.E(tag, "Unhandled error: %+v\n stack:%s", err.Error(), cast.ToString(debug.Stack()))
	}
	logger.I(tag, "app (%s) exiting...", state.ID)
}
