package watch

import (
	"os/exec"
)

type RunAction struct {
	Cmd      string
	Args     []string
	PostArgs []string
}

func (a *RunAction) Process(w *Watcher, file string) bool {
	w.report_action("Attempting to run ", a.Cmd, " on ", file )

	final_args := a.Args
	final_args = append(final_args, file)
	final_args = append(final_args, a.PostArgs...)
	cm := exec.Command(a.Cmd, final_args...)

	rerr := cm.Run()

	if rerr != nil {
		exerr, exerr_ok := rerr.(*exec.ExitError)
		if exerr_ok {
			w.error("Command failed with status ", exerr)
		} else {
			w.error(rerr)
		}
		return false
	}
	w.report_action("Command successful")
	return true
}
