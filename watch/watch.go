package watch

import (
	"github.com/prometheus/prometheus/util/flock"
	"gopkg.in/fsnotify.v1"
	"log"
	"os"
	"path/filepath"
)

/*
   describes the interface required of a directory action, eg PostAction
*/
type Action interface {
	Process(*Watcher, string)
}

/*
  A watcher config: a directory to watch,  it's associated actions and any global options
*/
type Config struct {
	Actions         []Action              /* List of actions to perform when new files arrive */
	AfterFileAction func(filename string) /* Callback to call after a file action */
	ArchiveDir      string                /* If set, place to store files after they have been successfully processed */
	Dir             string                /* Directory to watch */
	Debug           bool                  /* Verbose output */
	TestingOptions  []string              /* Misc behaviour flags largely for testing */
	dont_block      bool
}

/*
   An active watcher.
*/
type Watcher struct {
	Config    *Config
	fswatch   *fsnotify.Watcher
	test_opts map[string]bool
}

/*
  Start watching the directory in this config. This is a blocking activity so can be wrapped in a goroutine if you want to do other things!
*/
func Watch(c *Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	var w Watcher
	w.fswatch = watcher
	w.Config = c

	return w.run()
}

/*
   Close the watcher, stop watching!
*/
func (w *Watcher) Close() {
	w.fswatch.Close()
}

func (w *Watcher) run() error {

	/* Populate testing flags
	 */
	w.test_opts = make(map[string]bool)
	opts := &w.test_opts
	itm := &w.Config.TestingOptions
	for i, _ := range *itm {
		(*opts)[(*itm)[i]] = true
	}

	if w.test_opts["noblock"] {
		w.Config.dont_block = true
	}

	done := make(chan bool)

	/* Setup goroutine which just waits for events and errors from the filesystem watcher:
	 */
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

	/* Add the actual directory we're watching to the fswatcher
	 */
	werr = w.fswatch.Add(w.Config.Dir)

	/* Assuming all has gone well (and config isn't telling us not to block)
	   then just wait for a signal down our "done" channel
	*/
	if werr == nil && !w.Config.dont_block {
		<-done
		defer w.Close()
	}

	return werr
}

func (w *Watcher) handle_event(e *fsnotify.Event) {
	/* We have had a signal from the fswatcher. Most things we don't care about, but Create events we are excited by: */
	if e.Op == fsnotify.Create {
		w.debug("Create event for ", e.Name)
		release, existed, err := flock.New(e.Name)

		if !existed {
			w.debug("File didn't exist flock will have created it. I am too chicken to delete things though.. ")
		}
		if err != nil {
			w.debug("Lock failed")
		} else {
			defer release.Release()
		}

		w.actions_for_file(e.Name)

		_, filename := filepath.Split(e.Name)

		if w.Config.ArchiveDir != "" {
			e := os.Rename(e.Name, w.Config.ArchiveDir+string(os.PathSeparator)+filename)
			if e != nil {
				w.debug(e)
			}
		}

		if w.Config.AfterFileAction != nil {
			w.Config.AfterFileAction(e.Name)
		}
		if v := w.test_opts["exit_after_one"]; v {
			w.Close()
		}
	}
}

func (w *Watcher) actions_for_file(file_path string) {
	for _, v := range w.Config.Actions {
		v.Process(w, file_path)
	}
}

func (w *Watcher) debug(things ...interface{}) {
	if w.Config.Debug {
		log.Println(things)
	}
}
