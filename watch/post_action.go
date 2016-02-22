package watch

import (
	"net/http"
	"os"
	"time"
)

type PostAction struct {
	To                string
	Mime              string
	BasicAuthUsername string
	BasicAuthPwd      string
}

func (a *PostAction) Process(w *Watcher, file string) bool {
	w.debug("Attempting to post ", file, " to ", a.To)
	mime_type := a.Mime
	reader, err := os.Open(file)

	if err != nil {
		w.debug("error opeing file ", file, " ", err)
		return false
	}

	if mime_type == "" {
		// TODO: better
		mime_type = "text/plain"
	}

	req, err := http.NewRequest("POST", a.To, reader)

	if err != nil {
		w.debug("Error building request: ", err)
		return false
	}

	req.Header.Set("Content-Type", mime_type)

	if len(a.BasicAuthUsername) > 0 {
		req.SetBasicAuth(a.BasicAuthUsername, a.BasicAuthPwd)
	}

	var cli = &http.Client{
		Timeout: time.Second * 120,
	}
	rsp, err := cli.Do(req)

	if err != nil {
		w.debug("Posting ", file, " to ", a.To, " failed ", err)
		return false
	}

	w.debug("Got response ", rsp.Status)
	if rsp.StatusCode != http.StatusOK {
		return false
	}

	return true
}
