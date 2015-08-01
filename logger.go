package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func WriteFile(ch chan Message) {
	for {
		res := <-ch
		if res.Command != "MSG" {
			continue
		}
		basePath := "public/_public/" + strings.Title(res.Channel) + " chatlog/"
		timestamp := res.Timestamp.UTC()
		msgDate := timestamp.Format("2006-01-02")
		msgTime := timestamp.Format("[2006-01-02 15:04:05 MST] ")
		monthYear := timestamp.Format("January 2006")
		currMonthYear := time.Now().UTC().Format("January 2006")

		checkFolders(basePath + currMonthYear + "/userlogs/")

		f, err := OpenFile(fmt.Sprintf("%s%s/%s.txt", basePath, monthYear, msgDate))
		if err != nil {
			logErr(err, res)
			continue
		}
		f.WriteString(msgTime + res.Nick + ": " + res.Data + "\n")
		f.Close()
		f, err = OpenFile(fmt.Sprintf("%s%s/userlogs/%s.txt", basePath, monthYear, res.Nick))
		if err != nil {
			logErr(err, res)
			continue
		}
		f.WriteString(msgTime + res.Nick + ": " + res.Data + "\n")
		f.Close()
		// log.Printf("%s > %s: %s", res.Channel, res.Nick, res.Data)
	}
}

func logErr(err error, m interface{}) {
	if err != nil {
		log.Printf("ERROR: %s\n%q\n", err, m)
	}
}

func OpenFile(p string) (*os.File, error) {
	return os.OpenFile(fmt.Sprintf(p),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
}

func checkFolders(p string) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logErr(err, "")
		return
	}

	p = filepath.Join(dir, p)
	_, err = os.Stat(p)
	if err != nil {
		createFolders(p)
		return
	}
}

func createFolders(p string) {
	err := os.MkdirAll(p, 0755)
	if err != nil {
		logErr(err, "")
		return
	}
}
