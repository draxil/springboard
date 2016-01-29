package watch

import (
	"gopkg.in/fsnotify.v1"
	"github.com/prometheus/prometheus/util/flock"
	"log"
)

type Action interface {
	Process(*Watcher, string)
}

type Config struct {
	Actions []Action
	Dir     string
	Debug   bool
	AfterFileAction func(filename string)
	dont_block bool
}

func Watch(c *Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	var w Watcher
	w.fswatch = watcher
	w.config = c

	return w.run()
}

type Watcher struct {
	config  *Config
	fswatch *fsnotify.Watcher
}

func (w *Watcher) Close(){
	w.fswatch.Close()
}

func (w *Watcher) run() error {
	

	done := make(chan bool)
	var werr error
	go func() {
		for {
			select {
			case event := <-w.fswatch.Events:
				go w.handle_event(&event)
			case err := <-w.fswatch.Errors:
				werr = err
				done <- true
			}
		}

	}()

	werr = w.fswatch.Add(w.config.Dir)

	if werr == nil && !w.config.dont_block {
		log.Println("waiting")
		<-done
		defer w.Close()
	}

	return werr
}

func (w *Watcher) handle_event(e * fsnotify.Event) {
	if e.Op == fsnotify.Create {
		w.debug("Create event for ", e.Name)
		release, existed, err := flock.New( e.Name )
		if ! existed {
			w.debug("File didn't exist flock will have created it. I am too chicken to delete things though.. ")
		}
		if err != nil {
			w.debug("Lock failed")
		}else{
			defer release.Release()
		}
		w.actions_for_file( e.Name )
		if w.config.AfterFileAction != nil {
			w.config.AfterFileAction( e.Name )
		}
	}
}

func (w *Watcher) actions_for_file( file_path string ){
	for _, v :=  range w.config.Actions {
		v.Process( w, file_path)
	}
}

func (w *Watcher) debug(things ...interface{}) {
	if w.config.Debug {
		log.Println(things)
	}
}
