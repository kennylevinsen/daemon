// +build windows

package daemon

import (
	"fmt"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var elog debug.Log

type winSvc struct {
	d Daemon
}

func (ws *winSvc) Execute(args []string, cr <-chan svc.ChangeRequest, change chan<- svc.Status) (svcSpecific bool, errCode uint32) {
	change <- svc.Status{State: svc.StartPending}
	status := make(chan Status, 1)
	cb := func() {
		status <- ws.d.Status()
	}
	ws.d.SetCallback(cb)

	if err := ws.d.Start(args); err != nil {
		errCode = 1
		elog.Error(3, fmt.Sprintf("%s: start failed: %v", ws.d.Name(), err))
		goto exit
	}

	change <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for {
		select {
		case s := <-status:
			switch s {
			case Invalid:
				elog.Error(6, fmt.Sprintf("%s: invalid state", ws.d.Name()))
				errCode = 2
				goto exit
			case Stopped:
				elog.Error(7, fmt.Sprintf("%s: stopped by application", ws.d.Name()))
				goto exit
			}
		case c := <-cr:
			switch c.Cmd {
			case svc.Interrogate:
				change <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				goto exit
			default:
				elog.Error(4, fmt.Sprintf("%s: unexpected control request: #%d", ws.d.Name(), c))
			}
		}
	}

exit:
	elog.Info(5, fmt.Sprintf("%s: stopping", ws.d.Name()))
	change <- svc.Status{State: svc.StopPending}
	ws.d.Stop()
	change <- svc.Status{State: svc.Stopped}
	return false, errCode
}

// Run runs the daemon as either a Windows service or as a console application.
func Run(d Daemon) error {
	interactive, err := svc.IsAnInteractiveSession()
	if err != nil {
		interactive = true
	}

	if interactive {
		return Console(d)
	}

	elog, err = eventlog.Open(d.Name())
	if err != nil {
		return err
	}

	elog.Info(1, fmt.Sprintf("%s: starting", d.Name()))
	err = svc.Run(d.Name(), &winSvc{d: d})
	if err != nil {
		elog.Info(2, fmt.Sprintf("%s: service start failed: %v", d.Name(), err))
		return err
	}

	return nil
}
