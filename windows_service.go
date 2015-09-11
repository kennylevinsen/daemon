// +build windows

package daemon

import (
	"fmt"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

type winSvcer struct {
	elog debug.Log
}

func (wsl *winSvcer) Fatal(v ...interface{}) {
	wsl.elog.Error(0, fmt.Sprint(v...))
}

func (wsl *winSvcer) Fatalf(format string, v ...interface{}) {
	wsl.elog.Error(0, fmt.Sprintf(format, v...))
}

func (wsl *winSvcer) Print(v ...interface{}) {
	wsl.elog.Info(0, fmt.Sprint(v...))
}

func (wsl *winSvcer) Printf(format string, v ...interface{}) {
	wsl.elog.Info(0, fmt.Sprintf(format, v...))
}

type winSvc struct {
	d Daemon
}

func (ws *winSvc) Execute(args []string, cr <-chan svc.ChangeRequest, change chan<- svc.Status) (svcSpecific bool, errCode uint32) {
	change <- svc.Status{State: svc.StartPending}
	status := make(chan Status, 2)
	cb := func() {
		status <- ws.d.Status()
	}
	ws.d.SetCallback(cb)

	Args = args

	if err := ws.d.Start(); err != nil {
		errCode = 1
		Fatalf("%s: application start failed: %v", ws.d.Name(), err)
		goto exit
	}

	change <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for {
		select {
		case s := <-status:
			switch s {
			case Invalid:
				Fatalf("%s: invalid state", ws.d.Name())
				errCode = 2
				goto exit
			case Stopped:
				Fatalf("%s: stopped by application", ws.d.Name())
				goto exit
			}
		case c := <-cr:
			switch c.Cmd {
			case svc.Interrogate:
				change <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				goto exit
			default:
				Fatalf("%s: unexpected control request: #%d", ws.d.Name(), c)
			}
		}
	}

exit:
	Printf("%s: stopping", ws.d.Name())
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

	elog, err := eventlog.Open(d.Name())
	if err != nil {
		return err
	}

	wsl := &winSvcer{elog: elog}
	SetLogger(wsl)

	Printf("%s: starting", d.Name())
	err = svc.Run(d.Name(), &winSvc{d: d})
	if err != nil {
		Fatalf("%s: service start failed: %v", d.Name(), err)
		return err
	}

	return nil
}
