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
	{
		app := cli.NewApp()
		var our_wc watch.Config
		app.Flags = global_flags( &our_wc )
		is := make_is(t)
		app.Run([]string{""})
		is(our_wc.Debug, false, "debug off")
		is(our_wc.ProcessExistingFiles, false, "process existing off")
		is(our_wc.ReportErrors, true, "Error reportin on by default")
		is(our_wc.ReportActions, false, "Action reporting on by default");
		if our_wc.Paranoia != watch.NoParanoia {
			t.Fatal("unexpected paranoia")
		}
	}
	{
		app := cli.NewApp()
		var our_wc watch.Config
		app.Flags = global_flags( &our_wc )
		is := make_is(t)
		app.Run([]string{"", "--archive=FISHBOWL", "--error-dir=CATBASKET", "--debug", "--log-actions", "--log-errors=false", "--process-existing"})
		is(our_wc.ArchiveDir, "FISHBOWL", "archive dir")
		is(our_wc.ErrorDir, "CATBASKET", "error dir")
		is(our_wc.Debug, true, "debug on")
		is(our_wc.ReportActions, true, "action reporting on")
		is(our_wc.ReportErrors, false, "error reporting off")
		is(our_wc.ProcessExistingFiles, true, "process existing on")
	}
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
	defer func(){
		os.Remove( arch_dir )
		os.Remove( temp_dir )
	}()

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
		"--paranoia=off",
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

func TestSimpleRun(t *testing.T){
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
	other_dir := mk_temp_dir();

	defer func(){
		os.Remove( temp_dir )
		os.Remove( arch_dir )
		os.Remove( other_dir )
	}()
	
	app.Run([]string{"", "--archive=" + string(os.PathSeparator) + arch_dir, 
		"--testing=noblock",
		"--testing=exit_after_one",
		"--paranoia=off",
		"--log-actions",
		"--debug", "run", "watch/test1.sh", other_dir, temp_dir, })

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
	
	
	is := make_is(t)
	
	_, fe := os.Stat( arch_dir + string(os.PathSeparator) + "foo")
	is( fe, nil, "File stat on archive version of the file doesnt error")

	_, fe = os.Stat( other_dir + string(os.PathSeparator) + "foo")
	is( fe, nil, "File stat on run created version of the file doesnt error")


}


func TestPostargRun(t *testing.T){
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
	other_dir := mk_temp_dir();

	defer func(){
		os.Remove( temp_dir )
		os.Remove( arch_dir )
		os.Remove( other_dir )
	}()
	
	app.Run([]string{"", "--archive=" + string(os.PathSeparator) + arch_dir, 
		"--testing=noblock",
		"--testing=exit_after_one",
		"--paranoia=off",
		"--log-actions",
		"--debug", "run", "--postarg=" + other_dir, "/bin/cp", temp_dir, })

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
	
	
	is := make_is(t)
	
	_, fe := os.Stat( arch_dir + string(os.PathSeparator) + "foo")
	is( fe, nil, "File stat on archive version of the file doesnt error")

	_, fe = os.Stat( other_dir + string(os.PathSeparator) + "foo")
	is( fe, nil, "File stat on run created version of the file doesnt error")


}


func TestParanoia(t *testing.T){
	skip_long(t)
	app := app()
	mk_temp_dir := func()(string){
		s, e :=  ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}
	
	l,_ := net.Listen("tcp","127.0.0.1:0")
	addr := l.Addr().String()
	read_ok := false
	stuff_happened := false
	body := ""
	requests := 0
	s := &http.Server{
		Addr : addr,
		Handler : http.HandlerFunc( func( w http.ResponseWriter, r *http.Request){
			stuff_happened = true
			requests++
			body_bytes, err := ioutil.ReadAll( r.Body )
			if err != nil {
				log.Println( err )
			}else{
				body = string(body_bytes)
				read_ok = true
			}
		}),
	}


	l.Close()
	go s.ListenAndServe()
	
	temp_dir := mk_temp_dir()
	arch_dir := mk_temp_dir()

	app.Run([]string{"", "--archive=" + string(os.PathSeparator) + arch_dir, 
		"--testing=noblock",
		"--testing=exit_after_one",
		"--debug", "post", "http://" + addr, temp_dir, })

	original_filename := temp_dir + string(os.PathSeparator) + "foo"
	f, oferr := os.Create(original_filename)
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
			_, fe := os.Stat( original_filename )
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
	
	is := make_is(t)
	
	_, fe := os.Stat( arch_dir + string(os.PathSeparator) + "foo")
	is( fe, nil, "File stat on archive version of the file doesnt error")

	
	is( stuff_happened, true, "http server got a request")
	is( read_ok, true, "Read ok")
	is( body, "part onepart two", "body")
	is( requests, 1, "got one request")
}

func skip_long( t *testing.T ){
	if os.Getenv("LONGTESTS") != "1" {
		t.Skip("Not running extended tests set LONGTESTS environment var to include these")
	}
}


func make_is(t *testing.T) func(interface{}, interface{}, string) {
	return func(a interface{}, b interface{}, describe string) {
		if a != b {
			t.Fatal(describe)
		}
	}
}
