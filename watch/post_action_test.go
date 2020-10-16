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
	is := makeIs(t)
	
	l,_ := net.Listen("tcp","127.0.0.1:0")
	mine := l.Addr().String()
	stuffHappened := false
	readOk := false
	body := ""
	s := &http.Server{
		Addr : mine,
		Handler : http.HandlerFunc( func( w http.ResponseWriter, r *http.Request){
			stuffHappened = true
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
	

	tempDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
	 	panic(err)
	}
	wait := make( chan bool )

	defer func() { os.Remove(tempDir) }()
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug : false,
		Actions: []Action{
			&PostAction{
				To : "http://" + mine,
				Mime : "text/ralf",
			},
		},
		AfterFileAction : func( file string ){
			wait <- true
		},
	}

	Watch(&cfg)
	tempFile, err := ioutil.TempFile("", "springboard")
	if err != nil{
		panic(err)
	}
	log.Println(tempFile.Name())
	defer func(){ os.Remove( tempFile.Name() )}()
	tempFile.Write( []byte("kruncha6"))
	tempFile.Close()
	os.Rename( tempFile.Name(), tempDir + string(os.PathSeparator) + "foo")

	<- wait

	is( stuffHappened, true, "Post recieved")
	is( readOk, true, "Able to read body")
	is( body, "kruncha6", "Body checks out")
}

func Test_PostFail(t *testing.T){
	//is := makeIs(t)
	
	l,_ := net.Listen("tcp","127.0.0.1:0")
	mine := l.Addr().String()
	l.Close()

	tempDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
	 	panic(err)
	}
	wait := make( chan bool )

	defer func() { os.Remove(tempDir) }()
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug : false,
		Actions: []Action{
			&PostAction{
				To : "http://" + mine,
				Mime : "text/ralf",
			},
		},
		AfterFileAction : func( file string ){
			wait <- true
		},
	}

	Watch(&cfg)

	_, err = os.Create(tempDir + string(os.PathSeparator) + "foo")
	if err != nil {
	 	panic(err)
	}
	<- wait
}

func TestBasicAuth(t *testing.T){
	is := makeIs(t)
	
	l,_ := net.Listen("tcp","127.0.0.1:0")
	mine := l.Addr().String()
	stuffHappened := false
	readOk := false
	body := ""
	un, pwd := "", ""
	ba := false
	s := &http.Server{
		Addr : mine,
		Handler : http.HandlerFunc( func( w http.ResponseWriter, r *http.Request){
			stuffHappened = true
			bodyBytes, err := ioutil.ReadAll( r.Body )
			if err != nil {
				log.Println( err )
			}else{
				body = string(bodyBytes)
				readOk = true
				un, pwd, ba = r.BasicAuth()
			}
		}),
	}


	l.Close()
	go s.ListenAndServe()
	

	tempDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
	 	panic(err)
	}
	wait := make( chan bool )

	defer func() { os.Remove(tempDir) }()
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
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
		},
	}

	Watch(&cfg)
	tempFile, err := ioutil.TempFile("", "springboard")
	if err != nil{
		panic(err)
	}
	log.Println(tempFile.Name())
	defer func(){ os.Remove( tempFile.Name() )}()
	tempFile.Write( []byte("kruncha"))
	tempFile.Close()
	os.Rename( tempFile.Name(), tempDir + string(os.PathSeparator) + "foo")

	<- wait

	is( stuffHappened, true, "Post recieved")
	is( readOk, true, "Able to read body")
	is( body, "kruncha", "Body checks out")
	is( un, "parrappa", "Username")
	is( pwd, "therappa", "Password")
	is( ba, true, "Some basic auth happened")
}
