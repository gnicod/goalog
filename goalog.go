package main

import (
	"github.com/howeyc/fsnotify"
	"fmt"
	"regexp"
	"bufio"
	"io"
	"log"
	"os"
)

type Color struct {
	HEADER , OKBLUE , OKGREEN , WARNING , FAIL , ENDC string
}

var color     = Color{"\033[95m","\033[94m","\033[92m", "\033[93m","\033[91m","\033[0m"}
var logfile   = "/var/log/httpd/error_log"
var lineRegex = regexp.MustCompile("\\[(.*?)\\] \\[(.*?)\\] \\[(.*?)\\] (.*)")

func cprint(colorName string,text string) {
	pr := colorName + text + color.ENDC
	fmt.Println(pr)
}

func colorize_pattern(colorName string ,pattern string ,str string) string{
	src      := []byte(str)
	search   := regexp.MustCompile(pattern)
	colorN   := []byte(colorName)
	endcolor  := []byte(color.ENDC)
	i := -1
	src = search.ReplaceAllFunc(src, func(s []byte) []byte {
		if i != 0 {
			i -= 1
			var tmp = append(colorN,s...)
			return append(tmp,endcolor...)
		}
		return s
	})
	return (string(src))
}

func parse_line(input string) (string,string,string,string){
	if lineRegex.MatchString(input) {
		splitt      := lineRegex.FindAllStringSubmatch(input,-1)[0]
		return splitt[1],splitt[2],splitt[3],splitt[4]
	}
	return "","","",""
}

func get_last_line() string{
	f, err := os.Open(logfile)
	if err != nil {
		log.Fatal(err)
	}
	bf := bufio.NewReader(f)
	var lline string
	for {
		line, isPrefix, err := bf.ReadLine()
		if err == io.EOF {
			return lline
			break
		}
		lline = string(line)
		if err != nil {
			log.Fatal(err)
			log.Fatal(lline)
		}
		if isPrefix {
			log.Fatal("Error: Unexpected long line reading", f.Name())
		}
	}
	return string("aa ")
}

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	done := make(chan bool)
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if(ev.IsModify()){
					var line = get_last_line()
					da,ty,cl,msg := parse_line(line)
					fmt.Println(msg)
					fmt.Println(ty)
					fmt.Println(cl)
					fmt.Println(da)
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()
	err = watcher.Watch(logfile)
	if err != nil {
		log.Fatal(err)
	}
	<-done
	watcher.Close()
}
