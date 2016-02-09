package watch

import (
	"net/http"
	"os"
	"log"
)

type PostAction struct {
	To                string
	Mime              string
	BasicAuthUsername string
	BasicAuthPwd      string
}

func (a *PostAction) Process(w *Watcher, file string) {
	w.debug("Attempting to post ", file, " to ", a.To)
	mime_type := a.Mime
	reader, err := os.Open(file)

	if err != nil {
		w.debug("error opeing file ", file, " ", err)
		return
	}

	if mime_type == "" {
		// TODO: better
		mime_type = "text/plain"
	}

	req, err := http.NewRequest("POST", a.To, reader)

	if err != nil {
		w.debug("Error building request: ", err)
		return
	}

	log.Println("pre-p")

	if len(a.BasicAuthUsername) > 0 {
		log.Println("prep")
		req.SetBasicAuth(a.BasicAuthUsername, a.BasicAuthPwd)
	}
	
	client := &http.Client{}
	rsp, err := client.Do(req)

	if err != nil {
		w.debug("Posting ", file, " to ", a.To, " failed ", err)
		return
	}

	w.debug("Got response ", rsp.Status)

}
