package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"

	"github.com/MJKWoolnough/webserver/proxy/client"
)

var (
	gedcomFile  = flag.String("-g", "./tree.ged", "GEDCOM file")
	templateDir = flag.String("-t", "./templates", "template directory")
	filesDir    = flag.String("-f", "./files", "files directory")
	logName     = flag.String("-n", "", "name for logging")
	logger      *log.Logger
)

func main() {
	flag.Parse()
	logger = log.New(os.Stderr, *logName, log.LstdFlags)
	err := SetupGedcomData(*gedcomFile)
	if err != nil {
		logger.Println("error reading GEDCOM file: ", err)
		return
	}

	http.Handle("/FH/list.html", &List{
		Template: template.Must(template.ParseFiles(path.Join(*templateDir, "list.html.tmpl"))),
	})

	tree := &Tree{
		HTMLTemplate: template.Must(template.ParseFiles(path.Join(*templateDir, "tree.html.tmpl"))),
		JSTemplate:   template.Must(template.ParseFiles(path.Join(*templateDir, "tree.js.tmpl"))),
	}

	http.Handle("/FH/tree.html", http.HandlerFunc(tree.HTML))
	http.Handle("/FH/tree.js", http.HandlerFunc(tree.JS))

	http.Handle("/", http.FileServer(http.Dir(*filesDir)))

	cc := make(chan struct{})
	go func() {
		logger.Println("Server Started")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, os.Interrupt)
		select {
		case <-sc:
			logger.Println("Closing")
		case <-cc:
		}
		signal.Stop(sc)
		close(sc)
		client.Close()
		client.Wait()
		close(cc)
	}()

	err = client.Run()

	select {
	case <-cc:
	default:
		logger.Println(err)
		cc <- struct{}{}
	}
	<-cc
}
