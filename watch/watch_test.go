package watch

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

func Test_NoAction(t *testing.T) {
	is := makeIs(t)
	tempDir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	wait := make(chan bool)
	filename := ""
	defer func() { os.Remove(tempDir) }()
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug:      true,
		AfterFileAction: func(file string) {
			wait <- true
			filename = file
		},
	}

	Watch(&cfg)

	_, err = os.Create(tempDir + string(os.PathSeparator) + "foo")
	if err != nil {
		panic(err)
	}
	<-wait
	is(filename, tempDir+string(os.PathSeparator)+"foo",
		"Filename match")
	os.Remove(tempDir)
}

func TestArchive(t *testing.T) {

	mkTempDir := func() string {
		s, e := ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}

	tempDir := mkTempDir()
	archDir := mkTempDir()

	defer func() { os.Remove(tempDir) }()
	defer func() { os.Remove(archDir) }()

	wait := make(chan bool)
	//filename := ""

	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug:      true,
		AfterFileAction: func(file string) {
			wait <- true
			//filename = file
		},
		ArchiveDir: archDir,
	}

	Watch(&cfg)

	_, err := os.Create(tempDir + string(os.PathSeparator) + "foo")
	if err != nil {
		panic(err)
	}
	<-wait

	fileIn := makeFileIn(t, "foo")
	
	fileIn(archDir, true, "File IS in arch dir")
	fileIn(tempDir, false, "File is NOT in source dir")
}

func TestErrorDir(t *testing.T){
	mkTempDir := func() string {
		s, e := ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}

	tempDir := mkTempDir()
	archDir := mkTempDir()
	errDir := mkTempDir()

	defer func() { os.Remove(tempDir) }()
	defer func() { os.Remove(archDir) }()
	defer func() { os.Remove(errDir) }()

	wait := make(chan bool)
	//filename := ""

	action := DummyAction{}
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug:      true,
		Actions:    []Action{ &action },
		AfterFileAction: func(file string) {
			wait <- true
			//filename = file
		},
		ArchiveDir: archDir,
		ErrorDir:   errDir,
	}

	Watch(&cfg)

	_, err := os.Create(tempDir + string(os.PathSeparator) + "foo")
	if err != nil {
		panic(err)
	}
	<-wait
	action.FailPlease = true

	_, err = os.Create(tempDir + string(os.PathSeparator) + "bar")
	if err != nil {
		panic(err)
	}
	<-wait

	
	{
		fileIn := makeFileIn(t, "foo")
		fileIn(archDir, true, "File IS in arch dir")
		fileIn(errDir, false, "File is NOT in err dir")
		fileIn(tempDir, false, "File is NOT in source dir")
	}
	{
		fileIn := makeFileIn(t, "bar")
		fileIn(archDir, false, "File is NOT in arch dir")
		fileIn(errDir, true, "File IS in err dir")
		fileIn(tempDir, false, "File is NOT in source dir")
	}
}

func TestHandleExistingOff(t *testing.T) {
	mkTempDir := func() string {
		s, e := ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}

	tempDir := mkTempDir()
	archDir := mkTempDir()

	defer func() { os.Remove(tempDir) }()
	defer func() { os.Remove(archDir) }()

	for _, v := range []string{"zip", "zap", "zop"} {
		_, err := os.Create(tempDir + string(os.PathSeparator) + v)
		if err != nil {
			panic(err)
		}
	}

	wait := make(chan bool)
	//filename := ""
	
	expecting := 1
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug:      true,
		AfterFileAction: func(file string) {
			expecting--
			if expecting == 0 {
				wait <- true
				//filename = file
			}
		},
		ArchiveDir: archDir,
	}

	Watch(&cfg)

	_, err := os.Create(tempDir + string(os.PathSeparator) + "foo")
	if err != nil {
		panic(err)
	}
	<-wait

	for _, v := range []string{"zip", "zap", "zop"} {
		fileIn := makeFileIn(t, v)
		fileIn(archDir, false, "File is NOT in arch dir")
		fileIn(tempDir, true, "File IS in source dir")
	}
	for _, v := range []string{"foo"} {
		fileIn := makeFileIn(t, v)
		fileIn(archDir, true, "File IS in arch dir")
		fileIn(tempDir, false, "File is NOT in source dir")
	}
}
func TestHandleExistingOn(t *testing.T) {
	mkTempDir := func() string {
		s, e := ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}

	tempDir := mkTempDir()
	archDir := mkTempDir()

	defer func() { os.Remove(tempDir) }()
	defer func() { os.Remove(archDir) }()

	for _, v := range []string{"zip", "zap", "zop"} {
		_, err := os.Create(tempDir + string(os.PathSeparator) + v)
		if err != nil {
			panic(err)
		}
	}

	wait := make(chan bool)
	//filename := ""
	expecting := 4
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug:      true,
		AfterFileAction: func(file string) {
			expecting--
			if expecting == 0 {
				wait <- true
				//filename = file
			}
		},
		ArchiveDir:           archDir,
		ProcessExistingFiles: true,
	}

	Watch(&cfg)

	_, err := os.Create(tempDir + string(os.PathSeparator) + "foo")
	if err != nil {
		panic(err)
	}
	<-wait

	for _, v := range []string{"foo", "zip", "zap", "zop"} {
		fileIn := makeFileIn(t, v)
		fileIn(archDir, true, "File IS in arch dir")
		fileIn(tempDir, false, "File is NOT in source dir")
	}
}

func TestIgnoreDir(t *testing.T){
	mkTempDir := func() string {
		s, e := ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}

	tempDir := mkTempDir()
	archDir := tempDir + string(os.PathSeparator) + "arch"
	derr := os.Mkdir( archDir, 0777 )
	
	if derr != nil {
		panic(derr)
	}


	defer func() { os.Remove(tempDir) }()

	wait := make(chan bool)
	filename := ""
	expecting := 1
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug:      true,
		AfterFileAction: func(file string) {
			expecting--
			if expecting == 0 {
				wait <- true
				filename = file
			}
		},
		ArchiveDir:           archDir,
		ProcessExistingFiles: true,
	}

	Watch(&cfg)

	derr = os.Mkdir(  tempDir + string(os.PathSeparator) + "zing", 0700 )
	if derr != nil {
		panic(derr)
	}
	
	tfn := tempDir + string(os.PathSeparator) + "foo"
	_, err := os.Create(tfn)
	if err != nil {
		panic(err)
	}
	<-wait

	is := makeIs(t)
	is( filename, tfn, "got foo not any of the dirs" )
	fileIn := makeFileIn(t, "foo")
	fileIn( tempDir + string(os.PathSeparator) + "arch", true, "foo gets archived")
}

func TestParanoiaOff( t *testing.T ){
	skipLong(t)
	checkParanoid( t, NoParanoia )
}

func TestParanoiaOn( t *testing.T ){
	skipLong(t)
	checkParanoid( t, BasicParanoia )
}

func checkParanoid( t *testing.T, paranoid ParanoiaLevel ){
	mkTempDir := func() string {
		s, e := ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}

	tempDir := mkTempDir()
	archDir := tempDir + string(os.PathSeparator) + "arch"
	derr := os.Mkdir( archDir, 0777 )
	
	if derr != nil {
		panic(derr)
	}


	defer func() { os.Remove(tempDir) }()

	wait := make(chan bool)
	//filename := ""
	expecting := 1
	cfg := Config{
		dontBlock: true,
		Dir:        tempDir,
		Debug:      true,
		AfterFileAction: func(file string) {
			expecting--
			if expecting == 0 {
				wait <- true
				//filename = file
			}
		},
		ArchiveDir:           archDir,
		ProcessExistingFiles: true,
		Paranoia : paranoid,
	}

	Watch(&cfg)
	
	tfn := tempDir + string(os.PathSeparator) + "foo"
	f, err := os.Create(tfn)
	if err != nil {
		panic(err)
	}
	f.Write(([]byte)("part one"))
	time.Sleep( 120 * time.Millisecond )
	f.Write(([]byte)("part two"))
	f.Close()

	<-wait

	fileIn := makeFileIn(t, "foo")
	fileIn( tempDir + string(os.PathSeparator) + "arch", true, "foo in archive")
	
	tn := ""
	if paranoid > NoParanoia {
		tn = "foo not in input dir"
	} else {
		tn = "foo still in input dir"
	}

	fileIn( tempDir + string(os.PathSeparator) + "arch", true, tn)

	
}

func skipLong( t *testing.T ){
	if os.Getenv("LONGTESTS") != "1" {
		t.Skip("Not running extended tests set LONGTESTS environment var to include these")
	}
}


func makeFileIn(t *testing.T, fn string) func(string, bool, string) {
	return func(dir string, desired_result bool, describe string) {
		_, err := os.Stat(dir + string(os.PathSeparator) + fn)
		if desired_result == true && err != nil {

			if err != nil {
				log.Println(err)
			}
			t.Fatal(describe)
		} else if desired_result == false && (err == nil || !os.IsNotExist(err)) {
			if err != nil {
				log.Println(err)
			}
			t.Fatal(describe)
		}
	}
}
func makeIs(t *testing.T) func(interface{}, interface{}, string) {
	return func(a interface{}, b interface{}, describe string) {
		if a != b {
			t.Fatal(describe)
		}
	}
}
