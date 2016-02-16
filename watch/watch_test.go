package watch

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func Test_NoAction(t *testing.T) {
	is := make_is(t)
	temp_dir, err := ioutil.TempDir("", "springboard")
	if err != nil {
		panic(err)
	}
	wait := make(chan bool)
	filename := ""
	defer func() { os.Remove(temp_dir) }()
	cfg := Config{
		dont_block: true,
		Dir:        temp_dir,
		Debug:      true,
		AfterFileAction: func(file string) {
			wait <- true
			filename = file
		},
	}

	Watch(&cfg)

	_, err = os.Create(temp_dir + string(os.PathSeparator) + "foo")
	if err != nil {
		panic(err)
	}
	<-wait
	is(filename, temp_dir+string(os.PathSeparator)+"foo",
		"Filename match")
	os.Remove(temp_dir)
}

func TestArchive(t *testing.T) {

	mk_temp_dir := func() string {
		s, e := ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}

	temp_dir := mk_temp_dir()
	arch_dir := mk_temp_dir()

	defer func() { os.Remove(temp_dir) }()
	defer func() { os.Remove(arch_dir) }()

	wait := make(chan bool)
	filename := ""

	cfg := Config{
		dont_block: true,
		Dir:        temp_dir,
		Debug:      true,
		AfterFileAction: func(file string) {
			wait <- true
			filename = file
		},
		ArchiveDir: arch_dir,
	}

	Watch(&cfg)

	_, err := os.Create(temp_dir + string(os.PathSeparator) + "foo")
	if err != nil {
		panic(err)
	}
	<-wait

	file_in := make_file_in(t, "foo")

	file_in(arch_dir, true, "File IS in arch dir")
	file_in(temp_dir, false, "File is NOT in source dir")
}

func TestHandleExistingOff(t *testing.T) {
	mk_temp_dir := func() string {
		s, e := ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}

	temp_dir := mk_temp_dir()
	arch_dir := mk_temp_dir()

	defer func() { os.Remove(temp_dir) }()
	defer func() { os.Remove(arch_dir) }()

	for _, v := range []string{"zip", "zap", "zop"} {
		_, err := os.Create(temp_dir + string(os.PathSeparator) + v)
		if err != nil {
			panic(err)
		}
	}

	wait := make(chan bool)
	filename := ""
	
	expecting := 1
	cfg := Config{
		dont_block: true,
		Dir:        temp_dir,
		Debug:      true,
		AfterFileAction: func(file string) {
			expecting--
			if expecting == 0 {
				wait <- true
				filename = file
			}
		},
		ArchiveDir: arch_dir,
	}

	Watch(&cfg)

	_, err := os.Create(temp_dir + string(os.PathSeparator) + "foo")
	if err != nil {
		panic(err)
	}
	<-wait

	for _, v := range []string{"zip", "zap", "zop"} {
		file_in := make_file_in(t, v)
		file_in(arch_dir, false, "File is NOT in arch dir")
		file_in(temp_dir, true, "File IS in source dir")
	}
	for _, v := range []string{"foo"} {
		file_in := make_file_in(t, v)
		file_in(arch_dir, true, "File IS in arch dir")
		file_in(temp_dir, false, "File is NOT in source dir")
	}
}
func TestHandleExistingOn(t *testing.T) {
	mk_temp_dir := func() string {
		s, e := ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}

	temp_dir := mk_temp_dir()
	arch_dir := mk_temp_dir()

	defer func() { os.Remove(temp_dir) }()
	defer func() { os.Remove(arch_dir) }()

	for _, v := range []string{"zip", "zap", "zop"} {
		_, err := os.Create(temp_dir + string(os.PathSeparator) + v)
		if err != nil {
			panic(err)
		}
	}

	wait := make(chan bool)
	filename := ""
	expecting := 4
	cfg := Config{
		dont_block: true,
		Dir:        temp_dir,
		Debug:      true,
		AfterFileAction: func(file string) {
			expecting--
			if expecting == 0 {
				wait <- true
				filename = file
			}
		},
		ArchiveDir:           arch_dir,
		ProcessExistingFiles: true,
	}

	Watch(&cfg)

	_, err := os.Create(temp_dir + string(os.PathSeparator) + "foo")
	if err != nil {
		panic(err)
	}
	<-wait

	for _, v := range []string{"foo", "zip", "zap", "zop"} {
		file_in := make_file_in(t, v)
		file_in(arch_dir, true, "File IS in arch dir")
		file_in(temp_dir, false, "File is NOT in source dir")
	}
}

func TestIgnoreDir(t *testing.T){
	mk_temp_dir := func() string {
		s, e := ioutil.TempDir("", "springboard")
		if e != nil {
			panic(e)
		}
		return s
	}

	temp_dir := mk_temp_dir()
	arch_dir := temp_dir + string(os.PathSeparator) + "arch"
	derr := os.Mkdir( arch_dir, 0777 )
	
	if derr != nil {
		panic(derr)
	}


	defer func() { os.Remove(temp_dir) }()
	defer func() { os.Remove(arch_dir) }()
	wait := make(chan bool)
	filename := ""
	expecting := 1
	cfg := Config{
		dont_block: true,
		Dir:        temp_dir,
		Debug:      true,
		AfterFileAction: func(file string) {
			expecting--
			if expecting == 0 {
				wait <- true
				filename = file
			}
		},
		ArchiveDir:           arch_dir,
		ProcessExistingFiles: true,
	}

	Watch(&cfg)

	derr = os.Mkdir(  temp_dir + string(os.PathSeparator) + "zing", 0700 )
	if derr != nil {
		panic(derr)
	}
	
	tfn := temp_dir + string(os.PathSeparator) + "foo"
	_, err := os.Create(tfn)
	if err != nil {
		panic(err)
	}
	<-wait

	is := make_is(t)
	is( filename, tfn, "got foo not any of the dirs" )
	file_in := make_file_in(t, "foo")
	file_in( temp_dir + string(os.PathSeparator) + "arch", true, "foo gets archived")
}

func make_file_in(t *testing.T, fn string) func(string, bool, string) {
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
func make_is(t *testing.T) func(interface{}, interface{}, string) {
	return func(a interface{}, b interface{}, describe string) {
		if a != b {
			t.Fatal(describe)
		}
	}
}
