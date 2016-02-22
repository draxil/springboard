package watch

import "fmt"

type EchoAction struct {
}
func (a *EchoAction) Process( w *Watcher, file string) (bool){
	fmt.Println(file)
	return true
}
