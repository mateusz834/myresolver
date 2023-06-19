package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/mateusz834/dnsmsg"

	"github.com/mateusz834/myresolver"
)

var (
	//go:embed static/index.html
	rawIndexHTML string

	//go:embed static/main.js
	mainJS []byte

	//go:embed static/main.css
	mainCSS []byte
)

type server struct {
	m               sync.Mutex
	queriedMain     map[string]netip.Addr
	queriedFallback map[string]netip.Addr
	domain          string
	baseDomain      string
}

func NewServer(handleDomain string) server {
	m := make(map[string]netip.Addr)
	return server{domain: ".rand.api.get." + handleDomain + ".", baseDomain: handleDomain, queriedMain: m}
}

func (s *server) Run(dnsAddr netip.AddrPort, listenHTTPAddr string) error {
	tmpl, err := template.New("").Parse(rawIndexHTML)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, struct {
		BaseDomain string
	}{s.baseDomain}); err != nil {
		return err
	}
	index := b.Bytes()

	go func() {
		for {
			time.Sleep(time.Second * 25)

			s.m.Lock()
			s.queriedFallback = s.queriedMain
			s.queriedMain = make(map[string]netip.Addr, len(s.queriedMain))
			s.m.Unlock()

			time.Sleep(time.Second * 5)

			s.m.Lock()
			s.queriedFallback = nil
			s.m.Unlock()
		}
	}()

	errChan := make(chan error, 1)

	go func() {
		errChan <- myresolver.ListenUDPDNS(dnsAddr, s.handleQuery)
	}()

	go func() {
		errChan <- myresolver.ListenTCPDNS(dnsAddr, s.handleQuery)
	}()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", httpMethod(http.MethodGet, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write(index)
		}))
		mux.HandleFunc("/main.js", httpMethod(http.MethodGet, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/javascript")
			w.Write(mainJS)
		}))
		mux.HandleFunc("/main.css", httpMethod(http.MethodGet, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/css")
			w.Write(mainCSS)
		}))
		mux.HandleFunc("/api/who-resolved", httpMethod(http.MethodGet, s.whoResolvedHandler))

		errChan <- http.ListenAndServe(listenHTTPAddr, mux)
	}()

	return <-errChan
}

func (s *server) handleQuery(q dnsmsg.Question[dnsmsg.ParserName], srcAddr netip.Addr) {
	domain := q.Name.String()
	// TODO: this is a naive approach, it does not take into account
	// escape chatacters.
	if strings.HasSuffix(domain, s.domain) {
		s.m.Lock()
		s.queriedMain[domain] = srcAddr
		s.m.Unlock()
	}
}

func (s *server) getLastQueriedAddrOfDomain(domain string) (netip.Addr, bool) {
	s.m.Lock()
	defer s.m.Unlock()

	val, ok := s.queriedMain[domain]
	if ok {
		delete(s.queriedMain, domain)
		return val, true
	}
	if s.queriedFallback != nil {
		val, ok = s.queriedFallback[domain]
	}
	return val, ok
}

func (s *server) whoResolvedHandler(rw http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	addr, ok := s.getLastQueriedAddrOfDomain(domain)
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if addr.Is4In6() {
		addr = netip.AddrFrom4(addr.As4())
	}

	json.NewEncoder(rw).Encode(struct {
		Addr string `json:"addr"`
	}{Addr: addr.String()})
}

func httpMethod(method string, handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Add("Allow", method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		handler(w, r)
	})
}
