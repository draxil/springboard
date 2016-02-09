package main

import (
	"testing"
	"github.com/draxil/springboard/watch"
	"github.com/codegangsta/cli"
)
func Test_http_post_command( t * testing.T ){
	app := cli.NewApp()
	var our_wc watch.Config
	posted := false
	app.Commands = []cli.Command{
		http_post_command(&our_wc, func( wc * watch.Config){
			posted = true
		}),
	}
	is := make_is(t);
	app.Run([]string{"", "post", "http://goo.com", "x"})
	is( posted, true, "Post executed")
	is( len( our_wc.Actions), 1, "One action generated"  )
	a := our_wc.Actions[0]
	pa, ok := a.(*watch.PostAction)
	is( ok, true, "Not a post action")
	is( pa.To, "http://goo.com", "Post to stored")
	is( our_wc.Dir, "x", "Dir stored" )
}

func Test_http_post_command_opts( t * testing.T ){
	app := cli.NewApp()
	var our_wc watch.Config
	posted := false
	app.Commands = []cli.Command{
		http_post_command(&our_wc, func( wc * watch.Config){
			posted = true
		}),
	}
	is := make_is(t);
	app.Run([]string{"", "post", "--uname", "x", "--pass", "y", "--mime=x/y", "http://goo.com", "x"})
	is( posted, true, "Post executed")
	is( len( our_wc.Actions), 1, "One action generated"  )
	a := our_wc.Actions[0]
	pa, ok := a.(*watch.PostAction)
	is( ok, true, "Not a post action")
	is( pa.To, "http://goo.com", "Post to stored")
	is( our_wc.Dir, "x", "Dir stored" )
	is( pa.BasicAuthUsername, "x", "BasicAuthUsername")
	is( pa.BasicAuthPwd, "y", "BasicAuthPwd")
	is( pa.Mime, "x/y", "Mime")
}


func make_is( t *testing.T) func(interface{}, interface{}, string){
	return func(a interface{}, b interface{}, describe string){
		if( a != b ){
			t.Fatal( describe )
		}
	}
}
