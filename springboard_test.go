package main

import (
	"github.com/urfave/cli"
	"github.com/draxil/springboard/watch"
	"io/ioutil"
	"os"
	"testing"
	"time"
	"net/http"
	"net"
	"log"
//	"bytes"
)

func Test_http_post_command(t *testing.T) {
	app := cli.NewApp()
	var ourWc watch.Config
	posted := false
	app.Commands = []cli.Command{
		http_post_command(&ourWc, func(wc *watch.Config) {
			posted = true
		}),
	}
	is := makeIs(t)
	app.Run([]string{"", "post", "http://goo.com", "x"})
	is(posted, true, "Post executed")
	is(len(ourWc.Actions), 1, "One action generated")
	a := ourWc.Actions[0]
	pa, ok := a.(*watch.PostAction)
	is(ok, true, "Not a post action")
	is(pa.To, "http://goo.com", "Post to stored")
	is(ourWc.Dir, "x", "Dir stored")
}

func Test_http_post_command_opts(t *testing.T) {
	app := cli.NewApp()
	var ourWc watch.Config
	posted := false
	app.Commands = []cli.Command{
		http_post_command(&ourWc, func(wc *watch.Config) {
			posted = true
		}),
	}
	is := makeIs(t)
	app.Run([]string{"", "post", "--uname", "x", "--pass", "y", "--mime=x/y", "http://goo.com", "x"})
	is(posted, true, "Post executed")
	is(len(ourWc.Actions), 1, "One action generated")
	a := ourWc.Actions[0]
	pa, ok := a.(*watch.PostAction)
	is(ok, true, "Not a post action")
	is(pa.To, "http://goo.com", "Post to stored")
	is(ourWc.Dir, "x", "Dir stored")
	is(pa.BasicAuthUsername, "x", "BasicAuthUsername")
	is(pa.BasicAuthPwd, "y", "BasicAuthPwd")
	is(pa.Mime, "x/y", "Mime")
}

func Test_glob_opts(t *testing.T) {
	{
		app := cli.NewApp()
		var ourWc watch.Config
		app.Flags = globalFlags( &ourWc )
		is := makeIs(t)
		app.Run([]string{""})
		is(ourWc.Debug, false, "debug off")
		is(ourWc.ProcessExistingFiles, false, "process existing off")
		is(ourWc.ReportErrors, true, "Error reportin on by default")
		is(ourWc.ReportActions, false, "Action reporting on by default");
		if ourWc.Paranoia != watch.NoParanoia {
			t.Fatal("unexpected paranoia")
		}
	}
	{
		app := cli.NewApp()
		var ourWc watch.Config
		app.Flags = globalFlags( &ourWc )
		is := makeIs(t)
		app.Run([]string{"", "--archive=FISHBOWL", "--error-dir=CATBASKET", "--debug", "--log-actions", "--log-errors=false", "--process-existing"})
		is(ourWc.ArchiveDir, "FISHBOWL", "archive dir")
		is(ourWc.ErrorDir, "CATBASKET", "error dir")
		is(ourWc.Debug, true, "debug on")
		is(ourWc.ReportActions, true, "action reporting on")
		is(ourWc.ReportErrors, false, "error reporting off")
		is(ourWc.ProcessExistingFiles, true, "process existing on")
	}
}

func TestRunSimpleEcho(t *testing.T){
	app := app()
	mkTempDir := func()(string){
		s, e :=  ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}
	
	tempDir := mkTempDir()
	archDir := mkTempDir()
	defer func(){
		os.Remove( archDir )
		os.Remove( tempDir )
	}()

	out, ferr := ioutil.TempFile("", "springboard")
	if ferr != nil {
		panic( ferr )
	}
	defer func(){ 
		out.Close()
	}()
	
	realStout := os.Stdout
	os.Stdout = out

	app.Run([]string{"", "--archive=" + string(os.PathSeparator) + archDir, 
		"--testing=noblock",
		"--testing=exit_after_one",
		"--paranoia=off",
		"--debug", "echo", tempDir, })

	originalFilename := tempDir + string(os.PathSeparator) + "foo"
	_, oferr := os.Create(originalFilename)
	if oferr != nil {
		panic(oferr)
	}
	
	done := make(chan bool)
	timeout := make(chan bool)

	go func(){
		for {
			_, fe := os.Stat( originalFilename )
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
	
	os.Stdout = realStout
	
	is := makeIs(t)
	
	_, fe := os.Stat( archDir + string(os.PathSeparator) + "foo")
	is( fe, nil, "File stat on archive version of the file doesnt error")

	out.Seek(0,0)
	
	buf := make([]byte, len(originalFilename) + 1)
	_, err := out.Read(buf)
	if err != nil {
		panic( err )
	}

	is(string(buf), originalFilename + "\n", "echo process worked as expected")
}

func TestSimpleRun(t *testing.T){
	app := app()
	mkTempDir := func()(string){
		s, e :=  ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}
	
	tempDir := mkTempDir()
	archDir := mkTempDir()
	otherDir := mkTempDir();

	defer func(){
		os.Remove( tempDir )
		os.Remove( archDir )
		os.Remove( otherDir )
	}()
	
	app.Run([]string{"", "--archive=" + string(os.PathSeparator) + archDir, 
		"--testing=noblock",
		"--testing=exit_after_one",
		"--paranoia=off",
		"--log-actions",
		"--debug", "run", "watch/test1.sh", otherDir, tempDir, })

	originalFilename := tempDir + string(os.PathSeparator) + "foo"
	_, oferr := os.Create(originalFilename)
	if oferr != nil {
		panic(oferr)
	}
	
	done := make(chan bool)
	timeout := make(chan bool)

	go func(){
		for {
			_, fe := os.Stat( originalFilename )
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
	
	
	is := makeIs(t)
	
	_, fe := os.Stat( archDir + string(os.PathSeparator) + "foo")
	is( fe, nil, "File stat on archive version of the file doesnt error")

	_, fe = os.Stat( otherDir + string(os.PathSeparator) + "foo")
	is( fe, nil, "File stat on run created version of the file doesnt error")


}


func TestPostargRun(t *testing.T){
	app := app()
	mkTempDir := func()(string){
		s, e :=  ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}
	
	tempDir := mkTempDir()
	archDir := mkTempDir()
	otherDir := mkTempDir();

	defer func(){
		os.Remove( tempDir )
		os.Remove( archDir )
		os.Remove( otherDir )
	}()
	
	app.Run([]string{"", "--archive=" + string(os.PathSeparator) + archDir, 
		"--testing=noblock",
		"--testing=exit_after_one",
		"--paranoia=off",
		"--log-actions",
		"--debug", "run", "--postarg=" + otherDir, "/bin/cp", tempDir, })

	originalFilename := tempDir + string(os.PathSeparator) + "foo"
	_, oferr := os.Create(originalFilename)
	if oferr != nil {
		panic(oferr)
	}
	
	done := make(chan bool)
	timeout := make(chan bool)

	go func(){
		for {
			_, fe := os.Stat( originalFilename )
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
	
	
	is := makeIs(t)
	
	_, fe := os.Stat( archDir + string(os.PathSeparator) + "foo")
	is( fe, nil, "File stat on archive version of the file doesnt error")

	_, fe = os.Stat( otherDir + string(os.PathSeparator) + "foo")
	is( fe, nil, "File stat on run created version of the file doesnt error")


}


func TestParanoia(t *testing.T){
	skipLong(t)
	app := app()
	mkTempDir := func()(string){
		s, e :=  ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}
	
	l,_ := net.Listen("tcp","127.0.0.1:0")
	addr := l.Addr().String()
	readOk := false
	stuffHappened := false
	body := ""
	requests := 0
	s := &http.Server{
		Addr : addr,
		Handler : http.HandlerFunc( func( w http.ResponseWriter, r *http.Request){
			stuffHappened = true
			requests++
			bodyBytes, err := ioutil.ReadAll( r.Body )
			if err != nil {
				log.Println( err )
			}else{
				body = string(bodyBytes)
				readOk = true
			}
		}),
	}


	l.Close()
	go s.ListenAndServe()
	
	tempDir := mkTempDir()
	archDir := mkTempDir()

	app.Run([]string{"", "--archive=" + string(os.PathSeparator) + archDir, 
		"--testing=noblock",
		"--testing=exit_after_one",
		"--debug", "post", "http://" + addr, tempDir, })

	originalFilename := tempDir + string(os.PathSeparator) + "foo"
	f, oferr := os.Create(originalFilename)
	if oferr != nil {
		panic(oferr)
	}

	f.Write(([]byte)("part one"))
	time.Sleep( 250 * time.Millisecond )
	f.Write(([]byte)("part two"))

	done := make(chan bool)
	timeout := make(chan bool)

	go func(){
		for {
			_, fe := os.Stat( originalFilename )
			if os.IsNotExist( fe ) {
				done <- true
			}
		}
	}()
	go func(){
		time.Sleep(4 * time.Second)
		timeout <- true
	}()

	select {
	case <- done:
	case <- timeout:
		t.Fatal("Timed out awaitng file archive")
	}
	
	is := makeIs(t)
	
	_, fe := os.Stat( archDir + string(os.PathSeparator) + "foo")
	is( fe, nil, "File stat on archive version of the file doesnt error")

	
	is( stuffHappened, true, "http server got a request")
	is( readOk, true, "Read ok")
	is( body, "part onepart two", "body")
	is( requests, 1, "got one request")
}

func skipLong( t *testing.T ){
	if os.Getenv("LONGTESTS") != "1" {
		t.Skip("Not running extended tests set LONGTESTS environment var to include these")
	}
}


func makeIs(t *testing.T) func(interface{}, interface{}, string) {
	return func(a interface{}, b interface{}, describe string) {
		if a != b {
			t.Fatal(describe)
		}
	}
}
