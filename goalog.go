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

var color     = Color{"\033[1m\033[95m","\033[1m\033[94m","\033[1m\033[92m", "\033[1m\033[93m","\033[1m\033[91m","\033[0m"}

var lineErrorRegex = regexp.MustCompile("\\[(.*?)\\] \\[error\\] \\[(.*?)\\] (.*)")
var lineWarnRegex = regexp.MustCompile("\\[(.*?)\\] \\[warn\\] (.*)")
var lineNoticeRegex = regexp.MustCompile("\\[(.*?)\\] \\[notice\\] (.*)")

var msgParseErrorRegex = regexp.MustCompile("(.*?):(.*) in (.*) on (line [0-9]{1,}), referer: (.*)")

func colorize(colorName string,text string) string {
	return colorName + text + color.ENDC
}

func get_colorized_msg(input string) string{
	if msgParseErrorRegex.MatchString(input) {
		splitt      := msgParseErrorRegex.FindStringSubmatch(input)
		return colorize(color.WARNING,splitt[1]) +":" +splitt[2]+" \n\t=> "+colorize(color.OKBLUE,splitt[3])+" on "+colorize(color.OKBLUE,splitt[4])+"\n\t"+splitt[5]
	}
	return input
}

func parse_line(input string) {
	//Line [error]
	if lineErrorRegex.MatchString(input) {
		splitt      := lineErrorRegex.FindAllStringSubmatch(input,-1)[0]
		fmt.Printf("[%s] %s\n",colorize(color.FAIL,"error"),get_colorized_msg(splitt[3]))
	}
	//Line [warn]
	if lineWarnRegex.MatchString(input) {
		splitt      := lineWarnRegex.FindAllStringSubmatch(input,-1)[0]
		fmt.Printf("[%s] %s \n",colorize(color.WARNING,"warn"),splitt[2])
	}
	//Line [notice]
	if lineNoticeRegex.MatchString(input) {
		splitt      := lineNoticeRegex.FindAllStringSubmatch(input,-1)[0]
		fmt.Printf("[%s]  %s\n",colorize(color.HEADER,"notice"),splitt[2])
	}
}

func get_last_line(logfile string) string{
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
	return string("")
}

func main() {
	logfile := os.Args[1]
	fmt.Println("Watching ",logfile)
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
					var line = get_last_line(logfile)
					parse_line(line)
				}
			case err := <-watcher.Error:
				log.Fatal("error:", err)
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
