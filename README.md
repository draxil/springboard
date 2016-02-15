# springboard
Watch a directory for new files and bounce them on somewhere else
OR Simple "drop folder to POST reqest (or whatever)" type thing!

# usage

For the full range of options always best to do:

> springboard -h
 
And similarly for the options pertaining to the subcommands and actions:
 
> springboard post -h
 
# actions
 
 So far there are two actions:
 
 * post - Send the file content as an HTTP POST request
 * echo - Echo the file path to stdout (good for building shell pipelines)
 
 # status / plans
 
 This is at an early stage of development and is subject to change! Upcoming additions:
 
* Archiving behaviour
* Error handling behaviour
* Picking up files that appeared while springboard was "off" 

Please feel free to shout at me if I'm missing something you need.
