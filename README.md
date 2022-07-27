# Springboard
Watch a directory (drop folder) for new files and "bounce" them on somewhere else, via several means such as HTTP post or another command.

# Usage

Example:

> springboard post --mime text/xml --uname homer --pass s1mps0n https://my.server.com/service ./incoming

Which basically says, when new files appear in the "./incoming" directory send them as an http POST request to the url provided, using basic auth to let us in and force the mimetype to expect XML files.

For the full range of options always best to do:

> springboard -h
 
And similarly for the options pertaining to the subcommands and actions:
 
> springboard post -h

# Installing

## With go 
> go install github.com/draxil/springboard@latest

## With an older go
1. Get go (version 1.7+)
2. Make sure you have set $GOPATH to a sensible place you can write to. 
3. Issue:
> go get -u github.com/draxil/springboard

You will want to make sure that your $GOPATH/bin is in your $PATH or else move the executable from $GOPATH/bin/springboard somewhere in your $PATH.

## Github release builds
See [the github release page](https://github.com/draxil/springboard/releases)

# Actions
 
 So far the available actions are:
 
 * post - Send the file content as an HTTP POST request
 * run  - Execute a command with the new filename as an argument  
 * echo - Echo the file path to stdout (good for building shell pipelines)
 
# API

The code effective funtionality could be useful to a go coder independent of the command itself. I'll post a godoc link here once the documentation is in any kind of shape. If you particularly want this, please shout at me.

# Fiddly details

## Directories

springboard ignores directories, so it's completely safe to have subdirectories which you can use for your archives etc.

# Status / plans
 
 This is at an early stage of development and is subject to change! Upcoming additions:
 
* Error handling behaviour
* Filtering / regex for being selective
* Base paranoia desisions on fsnotify events rather than updated times?

Please feel free to contact at me if I'm missing something you need.

# Versions

* v0.3.2 - Version to cause github builds to start
* v0.3.1 - Cleanup filepath on "existing files" behaviour to remove double slashes
* v0.3.0 - run action, fix to archive behaviour, fixed disappearing flock library
* v0.2.1 - go 1.6 compile fixes
* v0.2.0 - Better error handing, better logging
* v0.1.0 - First useful version


# Credit

Much of the development time for this tool comes from printevolved, if you need print or print technology:

[http://www.printevolved.co.uk](http://www.printevolved.co.uk)
