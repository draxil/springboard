package watch

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

// These tests assume a unix like system, but ATM that's all that's expected.

func Test_RunOK(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(tempDir) }()

	archDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(archDir) }()

	wait := make(chan bool)
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug:      false,
		// using an archive dir as this shows command OK nicely
		ArchiveDir : archDir,
		ReportActions : true,
		Actions: []Action{
			&RunAction{
				Cmd: "/bin/echo",
			},
		},
		AfterFileAction: func(file string) {
			wait <- true
		},
	}

	Watch(&cfg)
	tempFile, err := ioutil.TempFile("", "springboard")
	if err != nil {
		panic(err)
	}
	//log.Println(tempFile.Name())
	defer func() { os.Remove(tempFile.Name()) }()
	tempFile.Write([]byte("kruncha6"))
	tempFile.Close()
	os.Rename(tempFile.Name(), tempDir+string(os.PathSeparator)+"foo")

	<-wait

	_, err = os.Open(tempDir + string(os.PathSeparator) + "foo")
	if err == nil {
		t.Error("Able to open the file which should have gone")
	}
}

func Test_RunFail(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(tempDir) }()

	archDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(archDir) }()

	wait := make(chan bool)
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug:      true,
		// using an archive dir as this shows command OK nicely
		ArchiveDir : archDir,
		ReportActions : true,
		Actions: []Action{
			&RunAction{
				Cmd: "/bin/false",
			},
		},
		AfterFileAction: func(file string) {
			wait <- true
		},
	}

	Watch(&cfg)
	tempFile, err := ioutil.TempFile("", "springboard")
	if err != nil {
		panic(err)
	}
	log.Println(tempFile.Name())
	defer func() { os.Remove(tempFile.Name()) }()
	tempFile.Write([]byte("kruncha6"))
	tempFile.Close()
	os.Rename(tempFile.Name(), tempDir+string(os.PathSeparator)+"foo")

	<-wait

	_, err = os.Open(tempDir + string(os.PathSeparator) + "foo")
	if err != nil {
		t.Error("Not able to open the file which should remain")
	}
}

func Test_RunArgs(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(tempDir) }()

	archDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(archDir) }()

	otherDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(otherDir) }()


	wait := make(chan bool)
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug:      true,
		// using an archive dir as this shows command OK nicely
		ArchiveDir : archDir,
		ReportActions : true,
		Actions: []Action{
			&RunAction{
				Cmd: "./test1.sh",
				Args : []string{
					otherDir,
				},
				
			},
		},
		AfterFileAction: func(file string) {
			wait <- true
		},
	}

	Watch(&cfg)
	tempFile, err := ioutil.TempFile("", "springboard")
	if err != nil {
		panic(err)
	}
	log.Println(tempFile.Name())
	defer func() { os.Remove(tempFile.Name()) }()
	tempFile.Write([]byte("kruncha6"))
	tempFile.Close()
	os.Rename(tempFile.Name(), tempDir+string(os.PathSeparator)+"foo")

	<-wait

	_, err = os.Open(otherDir + string(os.PathSeparator) + "foo")
	if err != nil {
		t.Error("Not able to open the file which should now exist in " + otherDir)
	}
}

func Test_RunPostArgs(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(tempDir) }()

	archDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(archDir) }()

	otherDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(otherDir) }()


	wait := make(chan bool)
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug:      true,
		// using an archive dir as this shows command OK nicely
		ArchiveDir : archDir,
		ReportActions : true,
		Actions: []Action{
			&RunAction{
				Cmd: "/bin/cp",
				PostArgs : []string{
					otherDir,
				},
				
			},
		},
		AfterFileAction: func(file string) {
			wait <- true
		},
	}

	Watch(&cfg)
	tempFile, err := ioutil.TempFile("", "springboard")
	if err != nil {
		panic(err)
	}
	log.Println(tempFile.Name())
	defer func() { os.Remove(tempFile.Name()) }()
	tempFile.Write([]byte("kruncha6"))
	tempFile.Close()
	os.Rename(tempFile.Name(), tempDir+string(os.PathSeparator)+"foo")

	<-wait

	_, err = os.Open(otherDir + string(os.PathSeparator) + "foo")
	if err != nil {
		t.Error("Not able to open the file which should now exist in " + otherDir)
	}
}
