package watch

import (
	"io/ioutil"
	"os"
	"testing"
	"net/http"
	"net"
	"log"
)


func Test_PostOK(t *testing.T){
	is := make_is(t)
	
	l,_ := net.Listen("tcp","127.0.0.1:0")
	mine := l.Addr().String()
	stuff_happened := false
	read_ok := false
	body := ""
	s := &http.Server{
		Addr : mine,
		Handler : http.HandlerFunc( func( w http.ResponseWriter, r *http.Request){
			stuff_happened = true
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
	

	temp_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
	 	panic(err)
	}
	wait := make( chan bool )
	filename := ""
	defer func() { os.Remove(temp_dir) }()
	cfg := Config{
		dont_block: true,
		Dir:        temp_dir,
		Debug : false,
		Actions: []Action{
			&PostAction{
				To : "http://" + mine,
				Mime : "text/ralf",
			},
		},
		AfterFileAction : func( file string ){
			wait <- true
			filename = file
		},
	}

	Watch(&cfg)
	temp_file, err := ioutil.TempFile("", "springboard")
	if err != nil{
		panic(err)
	}
	log.Println(temp_file.Name())
	defer func(){ os.Remove( temp_file.Name() )}()
	temp_file.Write( []byte("kruncha6"))
	temp_file.Close()
	os.Rename( temp_file.Name(), temp_dir + string(os.PathSeparator) + "foo")

	<- wait

	is( stuff_happened, true, "Post recieved")
	is( read_ok, true, "Able to read body")
	is( body, "kruncha6", "Body checks out")
}

func Test_PostFail(t *testing.T){
	//is := make_is(t)
	
	l,_ := net.Listen("tcp","127.0.0.1:0")
	mine := l.Addr().String()
	l.Close()

	temp_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
	 	panic(err)
	}
	wait := make( chan bool )
	filename := ""
	defer func() { os.Remove(temp_dir) }()
	cfg := Config{
		dont_block: true,
		Dir:        temp_dir,
		Debug : false,
		Actions: []Action{
			&PostAction{
				To : "http://" + mine,
				Mime : "text/ralf",
			},
		},
		AfterFileAction : func( file string ){
			wait <- true
			filename = file
		},
	}

	Watch(&cfg)

	_, err = os.Create(temp_dir + string(os.PathSeparator) + "foo")
	if err != nil {
	 	panic(err)
	}
	<- wait
}

func TestBasicAuth(t *testing.T){
	is := make_is(t)
	
	l,_ := net.Listen("tcp","127.0.0.1:0")
	mine := l.Addr().String()
	stuff_happened := false
	read_ok := false
	body := ""
	un, pwd := "", ""
	ba := false
	s := &http.Server{
		Addr : mine,
		Handler : http.HandlerFunc( func( w http.ResponseWriter, r *http.Request){
			stuff_happened = true
			body_bytes, err := ioutil.ReadAll( r.Body )
			if err != nil {
				log.Println( err )
			}else{
				body = string(body_bytes)
				read_ok = true
				un, pwd, ba = r.BasicAuth()
			}
		}),
	}


	l.Close()
	go s.ListenAndServe()
	

	temp_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
	 	panic(err)
	}
	wait := make( chan bool )
	filename := ""
	defer func() { os.Remove(temp_dir) }()
	cfg := Config{
		dont_block: true,
		Dir:        temp_dir,
		Debug : true,
		Actions: []Action{
			&PostAction{
				To : "http://" + mine,
				Mime : "text/ralf",
				BasicAuthUsername : "parrappa",
				BasicAuthPwd : "therappa",
			},
		},
		AfterFileAction : func( file string ){
			wait <- true
			filename = file
		},
	}

	Watch(&cfg)
	temp_file, err := ioutil.TempFile("", "springboard")
	if err != nil{
		panic(err)
	}
	log.Println(temp_file.Name())
	defer func(){ os.Remove( temp_file.Name() )}()
	temp_file.Write( []byte("kruncha"))
	temp_file.Close()
	os.Rename( temp_file.Name(), temp_dir + string(os.PathSeparator) + "foo")

	<- wait

	is( stuff_happened, true, "Post recieved")
	is( read_ok, true, "Able to read body")
	is( body, "kruncha", "Body checks out")
	is( un, "parrappa", "Username")
	is( pwd, "therappa", "Password")
	is( ba, true, "Some basic auth happened")}
