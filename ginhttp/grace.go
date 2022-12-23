package ginhttp

import (
	"net/http"
	"runtime/debug"

	"github.com/jpillora/overseer"
	"github.com/spf13/cast"

	"github.com/senpan/xserver/logger"
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
	logger.GetLogger().Infof(tag, "app (%s) listening...", tag, state.ID)
	err := server.Serve(state.Listener)
	if err != nil {
		logger.GetLogger().Errorf(tag, "Unhandled error: %+v\n stack:%s", err, cast.ToString(debug.Stack()))
	}
	logger.GetLogger().Infof(tag, "app (%s) exiting...", state.ID)
}
