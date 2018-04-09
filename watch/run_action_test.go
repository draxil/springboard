package watch

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

// These tests assume a unix like system, but ATM that's all that's expected.

func Test_RunOK(t *testing.T) {

	temp_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(temp_dir) }()

	arch_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(arch_dir) }()

	wait := make(chan bool)
	cfg := Config{
		dont_block: true,
		Dir:        temp_dir,
		Debug:      false,
		// using an archive dir as this shows command OK nicely
		ArchiveDir : arch_dir,
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
	temp_file, err := ioutil.TempFile("", "springboard")
	if err != nil {
		panic(err)
	}
	log.Println(temp_file.Name())
	defer func() { os.Remove(temp_file.Name()) }()
	temp_file.Write([]byte("kruncha6"))
	temp_file.Close()
	os.Rename(temp_file.Name(), temp_dir+string(os.PathSeparator)+"foo")

	<-wait

	_, err = os.Open(temp_dir + string(os.PathSeparator) + "foo")
	if err == nil {
		t.Error("Able to open the file which should have gone")
	}
}

func Test_RunFail(t *testing.T) {

	temp_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(temp_dir) }()

	arch_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(arch_dir) }()

	wait := make(chan bool)
	cfg := Config{
		dont_block: true,
		Dir:        temp_dir,
		Debug:      true,
		// using an archive dir as this shows command OK nicely
		ArchiveDir : arch_dir,
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
	temp_file, err := ioutil.TempFile("", "springboard")
	if err != nil {
		panic(err)
	}
	log.Println(temp_file.Name())
	defer func() { os.Remove(temp_file.Name()) }()
	temp_file.Write([]byte("kruncha6"))
	temp_file.Close()
	os.Rename(temp_file.Name(), temp_dir+string(os.PathSeparator)+"foo")

	<-wait

	_, err = os.Open(temp_dir + string(os.PathSeparator) + "foo")
	if err != nil {
		t.Error("Not able to open the file which should remain")
	}
}

func Test_RunArgs(t *testing.T) {

	temp_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(temp_dir) }()

	arch_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(arch_dir) }()

	other_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(other_dir) }()


	wait := make(chan bool)
	cfg := Config{
		dont_block: true,
		Dir:        temp_dir,
		Debug:      true,
		// using an archive dir as this shows command OK nicely
		ArchiveDir : arch_dir,
		ReportActions : true,
		Actions: []Action{
			&RunAction{
				Cmd: "./test1.sh",
				Args : []string{
					other_dir,
				},
				
			},
		},
		AfterFileAction: func(file string) {
			wait <- true
		},
	}

	Watch(&cfg)
	temp_file, err := ioutil.TempFile("", "springboard")
	if err != nil {
		panic(err)
	}
	log.Println(temp_file.Name())
	defer func() { os.Remove(temp_file.Name()) }()
	temp_file.Write([]byte("kruncha6"))
	temp_file.Close()
	os.Rename(temp_file.Name(), temp_dir+string(os.PathSeparator)+"foo")

	<-wait

	_, err = os.Open(other_dir + string(os.PathSeparator) + "foo")
	if err != nil {
		t.Error("Not able to open the file which should now exist in " + other_dir)
	}
}

func Test_RunPostArgs(t *testing.T) {

	temp_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(temp_dir) }()

	arch_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(arch_dir) }()

	other_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	defer func() { os.Remove(other_dir) }()


	wait := make(chan bool)
	cfg := Config{
		dont_block: true,
		Dir:        temp_dir,
		Debug:      true,
		// using an archive dir as this shows command OK nicely
		ArchiveDir : arch_dir,
		ReportActions : true,
		Actions: []Action{
			&RunAction{
				Cmd: "/bin/cp",
				PostArgs : []string{
					other_dir,
				},
				
			},
		},
		AfterFileAction: func(file string) {
			wait <- true
		},
	}

	Watch(&cfg)
	temp_file, err := ioutil.TempFile("", "springboard")
	if err != nil {
		panic(err)
	}
	log.Println(temp_file.Name())
	defer func() { os.Remove(temp_file.Name()) }()
	temp_file.Write([]byte("kruncha6"))
	temp_file.Close()
	os.Rename(temp_file.Name(), temp_dir+string(os.PathSeparator)+"foo")

	<-wait

	_, err = os.Open(other_dir + string(os.PathSeparator) + "foo")
	if err != nil {
		t.Error("Not able to open the file which should now exist in " + other_dir)
	}
}
