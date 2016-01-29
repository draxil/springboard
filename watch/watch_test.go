package watch

import (
	"io/ioutil"
	"os"
	"testing"
)

func Test_NoAction(t *testing.T) {
	is := make_is(t)
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
	is( filename, temp_dir + string(os.PathSeparator) + "foo",
		"Filename match")
}

func make_is(t *testing.T) func(interface{}, interface{}, string) {
	return func(a interface{}, b interface{}, describe string) {
		if a != b {
			t.Fatal(describe)
		}
	}
}
