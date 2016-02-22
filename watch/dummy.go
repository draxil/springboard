package watch

type DummyAction struct {
	LastFile   string
	FailPlease bool
}

func (a *DummyAction) Process(w *Watcher, file string) bool {
	a.LastFile = file
	if a.FailPlease {
		return false
	}
	return true
}
