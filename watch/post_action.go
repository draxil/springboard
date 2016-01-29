package watch

import (
	"net/http"
	"os"
)

type PostAction struct {
	To string
	Mime string
}

func (a *PostAction) Process( w *Watcher, file string) {
	w.debug("Attempting to post ", file , " to ", a.To )
	mime_type := a.Mime
	reader, err := os.Open(file)
	
	if err != nil{
		w.debug("error opeing file ", file, " ", err)
		return
	}
	
	if mime_type == "" {
		// TODO: better
		mime_type = "text/plain"
	}

	rsp, err := http.Post( a.To, mime_type, reader)

	w.debug("Got response ", rsp.Status)
	if err != nil {
		w.debug("Posting ", file, " to ", a.To, " failed ", err )
	}
}
