package daemon

import (
	"errors"
	golog "log"
	"os"
	"os/signal"
)

// Status describes the status of the daemon.
type Status int

// Arguments, exposed to work like os.Args.
var Args []string

// Daemon statuses.
const (
	Invalid Status = iota
	Stopped
	Running
)

// Daemon is a common multiplatform background task.
type Daemon interface {
	// Name returns the name of the daemon.
	Name() string

	// Start is called when the daemon should start. Start must not block, and
	// should return an error if Start was unsuccessful. On error, the state of
	// the daemon must not have changed. Calling Start on a started daemon must
	// have no side-effects.
	Start() error

	// Stop is called when the daemon should stop. Calling Stop on a stopped
	// daemon must have no side-effects.
	Stop()

	// Status returns the current status of the daemon.
	Status() Status

	// SetCallback sets a callback that will be called when the status is
	// changed.
	SetCallback(func())
}

// Console is a basic console runner for daemons.
func Console(d Daemon) error {
	SetLogger(golog.New(os.Stderr, "", golog.LstdFlags))

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	status := make(chan Status, 2)

	cb := func() {
		status <- d.Status()
	}

	d.SetCallback(cb)

	Args = os.Args

	if err := d.Start(); err != nil {
		return err
	}

	for {
		select {
		case <-sigint:
			d.Stop()
			return nil
		case s := <-status:
			switch s {
			case Stopped:
				return nil
			case Invalid:
				d.Stop()
				return errors.New("daemon changed to invalid state")
			}
		}
	}
}
