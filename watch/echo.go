package watch

import "fmt"

type EchoAction struct {
}
func (a *EchoAction) Process( w *Watcher, file string) {
	fmt.Println(file)
}
