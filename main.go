package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/rpc"
	"net/smtp"
	"os"
	"os/signal"
	"path"
	"strings"

	"golang.org/x/net/websocket"

	"github.com/MJKWoolnough/httpbuffer"
	"github.com/MJKWoolnough/httpgzip"
	"github.com/MJKWoolnough/httplog"
	"github.com/MJKWoolnough/webserver/contact"
	"github.com/MJKWoolnough/webserver/proxy/client"
)

var (
	gedcomFile  = flag.String("g", "./tree.ged", "GEDCOM file")
	templateDir = flag.String("t", "./templates", "template directory")
	filesDir    = flag.String("f", "./files", "files directory")
	logFile     = flag.String("l", "", "file for request logging")
	logName     = flag.String("n", "", "name for logging")
	logger      *log.Logger
)

var templateFuncs = template.FuncMap{
	"mul":  func(i, j uint) uint { return i * j },
	"add":  func(i, j uint) uint { return i + j },
	"sub":  func(i, j uint) uint { return i - j },
	"subr": func(i, j uint) uint { return j - i },
	"uint": func(i int) uint { return uint(i) },
	"int":  func(i uint) int { return int(i) },
	"gtr":  func(i, j uint) bool { return j > i },
}

func main() {
	flag.Parse()
	logger = log.New(os.Stderr, *logName, log.LstdFlags)
	if err := rpc.Register(RPC{}); err != nil {
		logger.Println(err)
		return
	}
	if err := SetupGedcomData(*gedcomFile); err != nil {
		logger.Println("error reading GEDCOM file: ", err)
		return
	}

	from := os.Getenv("contactFormFrom")
	os.Unsetenv("contactFormFrom")
	to := os.Getenv("contactFormTo")
	os.Unsetenv("contactFormTo")
	addr := os.Getenv("contactFormAddr")
	os.Unsetenv("contactFormAddr")
	username := os.Getenv("contactFormUsername")
	os.Unsetenv("contactFormUsername")
	password := os.Getenv("contactFormPassword")
	os.Unsetenv("contactFormPassword")
	p := strings.IndexByte(addr, ':')
	addrMPort := addr
	if p > 0 {
		addrMPort = addrMPort[:p]
	}
	tmpl := template.Must(template.ParseFiles(path.Join(*filesDir, "/FH/contact.html")))
	ec := make(chan error)
	go func() {
		for {
			logger.Println(<-ec)
		}
	}()
	fh := new(http.ServeMux)
	fh.Handle("/FH/contact.html", &contact.Contact{
		Template: tmpl,
		From:     from,
		To:       to,
		Host:     addr,
		Auth:     smtp.PlainAuth("", username, password, addrMPort),
		Err:      ec,
	})

	list := &List{
		ListTemplate:     template.Must(template.ParseFiles(path.Join(*templateDir, "list.html.tmpl"))),
		RelationTemplate: template.Must(template.ParseFiles(path.Join(*templateDir, "relation.html.tmpl"))),
	}
	fh.Handle("/FH/list.html", http.HandlerFunc(list.List))
	fh.Handle("/FH/calc.html", http.HandlerFunc(list.Calculator))
	tree := &Tree{
		HTMLTemplate: template.Must(template.New("tree.html.tmpl").Funcs(templateFuncs).ParseFiles(path.Join(*templateDir, "tree.html.tmpl"))),
	}

	fh.Handle("/FH/tree.html", http.HandlerFunc(tree.HTML))
	http.Handle("/FH/", httpbuffer.Handler{fh})
	http.Handle("/FH/rpc", websocket.Handler(rpcHandler))
	http.Handle("/", httpgzip.FileServer(http.Dir(*filesDir)))

	var (
		lFile  *os.File
		server = &http.Server{
			Handler:  http.DefaultServeMux,
			ErrorLog: logger,
		}
	)
	if *logFile != "" {
		var err error
		lFile, err = os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Printf("Error appending to log file: %s\n", err)
		} else {
			lr, err := httplog.NewWriteLogger(lFile, httplog.DefaultFormat)
			if err != nil {
				logger.Fatalf("error starting request logging: %s\n", err)
			}
			server.Handler = httplog.Wrap(http.DefaultServeMux, lr)
		}

	}

	if err := client.Setup(server); err != nil {
		logger.Fatalf("error setting up server: %s\n", err)
	}

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

	err := client.Run()
	if lFile != nil {
		lFile.Close()
	}

	select {
	case <-cc:
	default:
		logger.Println(err)
		cc <- struct{}{}
	}
	<-cc
}
