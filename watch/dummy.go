package watch


type DummyAction struct {
	LastFile string
}

func (a *DummyAction) Process( w *Watcher, file string) {
	a.LastFile = file
}
