package main

import (
	"github.com/codegangsta/cli"
	"github.com/draxil/springboard/watch"
	"io/ioutil"
	"os"
	"testing"
	"time"
//	"bytes"
)

func Test_http_post_command(t *testing.T) {
	app := cli.NewApp()
	var our_wc watch.Config
	posted := false
	app.Commands = []cli.Command{
		http_post_command(&our_wc, func(wc *watch.Config) {
			posted = true
		}),
	}
	is := make_is(t)
	app.Run([]string{"", "post", "http://goo.com", "x"})
	is(posted, true, "Post executed")
	is(len(our_wc.Actions), 1, "One action generated")
	a := our_wc.Actions[0]
	pa, ok := a.(*watch.PostAction)
	is(ok, true, "Not a post action")
	is(pa.To, "http://goo.com", "Post to stored")
	is(our_wc.Dir, "x", "Dir stored")
}

func Test_http_post_command_opts(t *testing.T) {
	app := cli.NewApp()
	var our_wc watch.Config
	posted := false
	app.Commands = []cli.Command{
		http_post_command(&our_wc, func(wc *watch.Config) {
			posted = true
		}),
	}
	is := make_is(t)
	app.Run([]string{"", "post", "--uname", "x", "--pass", "y", "--mime=x/y", "http://goo.com", "x"})
	is(posted, true, "Post executed")
	is(len(our_wc.Actions), 1, "One action generated")
	a := our_wc.Actions[0]
	pa, ok := a.(*watch.PostAction)
	is(ok, true, "Not a post action")
	is(pa.To, "http://goo.com", "Post to stored")
	is(our_wc.Dir, "x", "Dir stored")
	is(pa.BasicAuthUsername, "x", "BasicAuthUsername")
	is(pa.BasicAuthPwd, "y", "BasicAuthPwd")
	is(pa.Mime, "x/y", "Mime")
}

func Test_glob_opts(t *testing.T) {
	app := cli.NewApp()
	var our_wc watch.Config
	app.Flags = global_flags( &our_wc )
	is := make_is(t)
	app.Run([]string{"", "--archive=FISHBOWL", "--debug"})
	is(our_wc.ArchiveDir, "FISHBOWL", "archive dir")
	is(our_wc.Debug, true, "debug on")
}

func TestRunSimpleEcho(t *testing.T){
	app := app()
	mk_temp_dir := func()(string){
		s, e :=  ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}
	
	temp_dir := mk_temp_dir()
	arch_dir := mk_temp_dir()
	out, ferr := ioutil.TempFile("", "springboard")
	if ferr != nil {
		panic( ferr )
	}
	defer func(){ 
		out.Close()
	}()
	
	real_stout := os.Stdout
	os.Stdout = out

	app.Run([]string{"", "--archive=" + string(os.PathSeparator) + arch_dir, 
		"--testing=noblock",
		"--testing=exit_after_one",
		"--debug", "echo", temp_dir, })

	original_filename := temp_dir + string(os.PathSeparator) + "foo"
	_, oferr := os.Create(original_filename)
	if oferr != nil {
		panic(oferr)
	}
	
	done := make(chan bool)
	timeout := make(chan bool)

	go func(){
		for {
			_, fe := os.Stat( original_filename )
			if os.IsNotExist( fe ) {
				done <- true
			}
		}
	}()
	go func(){
		time.Sleep(2 * time.Second)
		timeout <- true
	}()

	select {
	case <- done:
	case <- timeout:
		t.Fatal("Timed out awaitng file archive")
	}
	
	os.Stdout = real_stout
	
	is := make_is(t)
	
	_, fe := os.Stat( arch_dir + string(os.PathSeparator) + "foo")
	is( fe, nil, "File stat on archive version of the file doesnt error")

	out.Seek(0,0)
	
	buf := make([]byte, len(original_filename) + 1)
	_, err := out.Read(buf)
	if err != nil {
		panic( err )
	}

	is(string(buf), original_filename + "\n", "echo process worked as expected")

}

func make_is(t *testing.T) func(interface{}, interface{}, string) {
	return func(a interface{}, b interface{}, describe string) {
		if a != b {
			t.Fatal(describe)
		}
	}
}
