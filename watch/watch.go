package watch

import (
	"fmt"
	"github.com/theckman/go-flock"
	"github.com/draxil/gomv"
	"gopkg.in/fsnotify.v1"
	"log"
	"os"
	"path/filepath"
	"time"
)

/*
   describes the interface required of a directory action, eg PostAction
*/
type Action interface {
	Process(*Watcher, string) bool
}

const (
	NoParanoia = 0 + iota
	BasicParanoia
	ExtraParanoia
)

type ParanoiaLevel int

/*
  A watcher config: a directory to watch,  it's associated actions and any global options
*/
type Config struct {
	Actions              []Action              /* List of actions to perform when new files arrive */
	AfterFileAction      func(filename string) /* Callback to call after a file action */
	ArchiveDir           string                /* If set, place to store files after they have been successfully processed */
	ErrorDir             string                /* If set, place to store files if an action fails */
	Dir                  string                /* Directory to watch */
	ProcessExistingFiles bool                  /* Process pre-existing files on startup */
	Paranoia             ParanoiaLevel         /* Wait and see if file is finished writing */
	Debug                bool                  /* Verbose output */
	ReportActions        bool                  /* Log actions */
	ReportErrors         bool                  /* Error output */
	TestingOptions       []string              /* Misc behaviour flags largely for testing */
	dont_block           bool
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

	/* before we start watching dispatch goroutines to process any pre-existing files:
	 */
	if w.Config.ProcessExistingFiles {
		w.process_existing()
	}

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

func (w *Watcher) process_existing() {
	w.debug("Processing existing files")
	f, err := os.Open(w.Config.Dir)
	if err != nil {
		panic(fmt.Sprintf("Error opening directory: %s", err))
	}

	fi, err := f.Readdirnames(0)
	if err != nil {
		panic(err)
	}

	for _, v := range fi {
		w.report_action("Processing existing file: " + v)

		/* Unlike the usual entrypoint these are "just filenames" so glue on the path first */
		path := w.Config.Dir + string(os.PathSeparator) + v
		path = filepath.Clean(path)
		go w.handleFile(path)
	}
}

func (w *Watcher) handle_event(e *fsnotify.Event) {
	/* We have had a signal from the fswatcher. Most things we don't care about, but Create events we are excited by: */
	if e.Op == fsnotify.Create {
		w.debug("Create event for ", e.Name)
		w.handleFile(e.Name)
	}
}

func (w *Watcher) handleFile(path string) {

	if !w.wantFile(path) {
		return
	}

	file_lock := flock.NewFlock(path)
	locked, err := file_lock.TryLock()
	
	if err != nil || !locked {
		w.error("Lock failed")
	} else {
		defer file_lock.Unlock()
	}

	if w.Config.Paranoia > NoParanoia {
		for w.paranoiaWait(path) {
			time.Sleep(250 * time.Millisecond)
		}
	}

	actions_ok := w.actions_for_file(path)

	_, filename := filepath.Split(path)

	already_archived := false
	archive := func( dir string ){
		if ! already_archived {
			w.report_action( "Archiving ", path, " to ",dir )
			e := gomv.MoveFile(path, dir+string(os.PathSeparator)+filename)
			if e != nil {
				w.error(e)
			}else{
				already_archived = true
			}
		}
	}
	
	if !actions_ok && w.Config.ErrorDir != "" {
		archive( w.Config.ErrorDir )
	}
	if actions_ok && w.Config.ArchiveDir != "" {
		archive( w.Config.ArchiveDir )
	}

	if w.Config.AfterFileAction != nil {
		w.Config.AfterFileAction(path)
	}
	if v := w.test_opts["exit_after_one"]; v {
		w.Close()
	}
}

func (w *Watcher) wantFile(filepath string) bool {
	fi, err := os.Stat(filepath)
	if err != nil {
		w.debug(fmt.Sprintf("Could not stat file (%s): %s", filepath, err))
		return false
	}

	if fi.IsDir() {
		w.debug("Rejecting dir")
		return false
	}

	// TODO: put this in as well when we have time to write a test
	/*if ! fi.Mode().IsRegular() {
		w.debug("Rejecting irregular file")
		return false
	}*/

	return true
}

func (w *Watcher) paranoiaWait(filepath string) bool {
	fi, err := os.Stat(filepath)
	if err != nil {
		w.error("Could not stat file to determine if it's ready. Going ahead!")
		return false
	}

	modtime := fi.ModTime()
	now := time.Now()

	if !now.After(modtime) {
		w.error("File modified in the future. Going ahead!")
		return true
	}

	dur := now.Sub(modtime)

	var waitfor time.Duration
	switch w.Config.Paranoia {
	case BasicParanoia:
		{
			waitfor = 2 * time.Second
		}
	case ExtraParanoia:
		{
			waitfor = 30 * time.Second
		}
	}

	if dur <= waitfor {
		w.debug("File modified recently, hang on")
		return true
	}

	// go ahead
	return false

}

func (w *Watcher) actions_for_file(file_path string) (bool) {
	for _, v := range w.Config.Actions {
		ok := v.Process(w, file_path)
		if( ! ok ){
			return false
		}
	}
	return true
}


func (w *Watcher) report_action(things ...interface{}) {
	if w.Config.ReportActions {
		w.report(things)
	}
}
func (w *Watcher) debug(things ...interface{}) {
	if w.Config.Debug {
		w.report(things)
	}
}

func (w *Watcher) error(things ...interface{}) {
	if w.Config.ReportErrors || w.Config.Debug {
		w.report(things)
	}
}

func (w *Watcher) report(things ...interface{}) {
	if w.Config.ReportErrors  || w.Config.Debug  {
		log.Println(things)
	}
}
