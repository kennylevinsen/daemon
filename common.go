// +build !windows

package daemon

// Run runs the daemon as a console application.
func Run(d Daemon) error {
	return Console(d)
}
