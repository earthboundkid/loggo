package server

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"

	"github.com/carlmjohnson/crockford"
	"github.com/carlmjohnson/flagx"
	"github.com/carlmjohnson/versioninfo"
)

const AppName = "leggo"

func CLI(args []string) error {
	var app appEnv
	err := app.ParseArgs(args)
	if err != nil {
		return err
	}
	if err = app.Exec(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}

func (app *appEnv) ParseArgs(args []string) error {
	fl := flag.NewFlagSet(AppName, flag.ContinueOnError)
	app.Logger = log.New(os.Stderr, AppName+" ", log.LstdFlags)
	flagx.BoolFunc(fl, "silent", "don't log debug output", func() error {
		app.Logger.SetOutput(io.Discard)
		return nil
	})
	fl.IntVar(&app.port, "port", 50039, "port to use for listening to connections")
	fl.StringVar(&app.dest, "dest", "loggo", "destination folder for log files")
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), `loggo - %s

Loggo logs requests for when you need to debug a webhook

Usage:

	loggo [options]

Options:
`, versioninfo.Short())
		fl.PrintDefaults()
	}
	versioninfo.AddFlag(fl)
	if err := fl.Parse(args); err != nil {
		return err
	}
	if err := flagx.ParseEnv(fl, AppName); err != nil {
		return err
	}
	return nil
}

type appEnv struct {
	*log.Logger
	port int
	dest string
}

func (app *appEnv) Exec() (err error) {
	app.Printf("starting on :%d", app.port)
	http.HandleFunc("/", app.loggo)
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", app.port), nil)
	return err
}

func (app *appEnv) loggo(w http.ResponseWriter, r *http.Request) {
	app.Printf("%s %q", r.Method, r.URL.Path)
	now := time.Now()
	req, err := httputil.DumpRequest(r, true)
	if err != nil {
		app.Println("read error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, http.StatusText(http.StatusInternalServerError))
		return
	}
	name := crockford.Time(crockford.Lower, now) + crockford.Random(crockford.Lower)
	name = crockford.Partition(name, 4) + ".txt"
	name = filepath.Join(app.dest, name)
	app.Println("writing", name)
	_ = os.MkdirAll(app.dest, 0700)
	if err = os.WriteFile(name, req, 0600); err != nil {
		app.Println("write error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, http.StatusText(http.StatusInternalServerError))
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
