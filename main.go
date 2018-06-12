package main // import "vimagination.zapto.org/vimagination.zapto.org"

import (
	"crypto/tls"
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/smtp"
	"os"
	"os/signal"
	"path"
	"strings"

	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/websocket"

	"vimagination.zapto.org/httpbuffer"
	_ "vimagination.zapto.org/httpbuffer/deflate"
	_ "vimagination.zapto.org/httpbuffer/gzip"
	"vimagination.zapto.org/httpgzip"
	"vimagination.zapto.org/httplog"
	"vimagination.zapto.org/httprpc"
	"vimagination.zapto.org/webserver/contact"
	"vimagination.zapto.org/webserver/proxy/client"
)

var (
	gedcomFile    = flag.String("g", "./tree.ged", "GEDCOM file")
	templateDir   = flag.String("t", "./templates", "template directory")
	filesDir      = flag.String("f", "./files", "files directory")
	compressedDir = flag.String("c", "./compressedfiles", "compressed files directory")
	logFile       = flag.String("l", "", "file for request logging")
	logName       = flag.String("n", "", "name for logging")
	logger        *log.Logger
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

type http2https struct {
	http.Handler
}

func (hh http2https) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.TLS == nil {
		var url = "https://" + r.Host + r.URL.Path
		if len(r.URL.RawQuery) != 0 {
			url += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, url, http.StatusMovedPermanently)
		return
	}
	hh.Handler.ServeHTTP(w, r)
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
	http.Handle("/FH/contact.html", httpbuffer.Handler{&contact.Contact{
		Template: tmpl,
		From:     from,
		To:       to,
		Host:     addr,
		Auth:     smtp.PlainAuth("", username, password, addrMPort),
		Err:      ec,
	}})

	list := &List{
		ListTemplate:     template.Must(template.ParseFiles(path.Join(*templateDir, "list.html.tmpl"))),
		RelationTemplate: template.Must(template.ParseFiles(path.Join(*templateDir, "relation.html.tmpl"))),
	}
	http.Handle("/FH/list.html", httpbuffer.Handler{http.HandlerFunc(list.List)})
	http.Handle("/FH/calc.html", httpbuffer.Handler{http.HandlerFunc(list.Calculator)})
	tree := &Tree{
		HTMLTemplate: template.Must(template.New("tree.html.tmpl").Funcs(templateFuncs).ParseFiles(path.Join(*templateDir, "tree.html.tmpl"))),
	}

	var compressed []http.FileSystem
	if *compressedDir != "" {
		compressed = []http.FileSystem{http.Dir(*compressedDir)}
	}

	http.Handle("/FH/tree.html", httpbuffer.Handler{http.HandlerFunc(tree.HTML)})
	http.Handle("/FH/rpc", &rpcSwitch{websocket.Handler(rpcWebsocketHandler), httpbuffer.Handler{httprpc.Handle(nil, jsonrpc.NewServerCodec, 1<<12, "application/json; charset=utf-8")}})
	http.Handle("/", httpgzip.FileServer(http.Dir(*filesDir), compressed...))

	var (
		lFile     *os.File
		leManager = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache("./certcache/"),
			HostPolicy: autocert.HostWhitelist("vimagination.zapto.org"),
		}
		server = &http.Server{
			Handler: leManager.HTTPHandler(http2https{http.DefaultServeMux}),
			TLSConfig: &tls.Config{
				GetCertificate: leManager.GetCertificate,
				NextProtos:     []string{"h2", "http/1.1"},
			},
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
