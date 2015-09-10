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
		goto exit
	}

	change <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for {
		select {
		case s := <-status:
			switch s {
			case Invalid:
				errCode = 2
				goto exit
			case Stopped:
				goto exit
			}
		case c := <-cr:
			switch c.Cmd {
			case svc.Interrogate:
				change <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				goto exit
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request: #%d", c))
			}
		}
	}

exit:
	change <- svc.Status{State: svc.StopPending}
	ws.d.Stop()
	change <- svc.Status{State: svc.Stopped}
	return false, errCode
}

// Run runs the daemon as either a Windows service or as a console application.
func Run(d Daemon) error {
	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		return fmt.Errorf("unable to determine if started as a service: %v", err)
	}

	if isIntSess {
		return Console(d)
	}

	ws := &winSvc{d: d}

	elog, err = eventlog.Open(d.Name())
	if err != nil {
		return err
	}
	err = svc.Run(d.Name(), ws)
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", d.Name(), err))
		return err
	}

	elog.Info(1, fmt.Sprintf("%s service stopped", d.Name()))
	return nil
}
