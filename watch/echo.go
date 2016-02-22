package watch

import "fmt"

type EchoAction struct {
}
func (a *EchoAction) Process( w *Watcher, file string) (bool){
	w.report_action("Echoing ", file)
	fmt.Println(file)
	return true
}
