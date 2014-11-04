/* GPLv3 */
/* deanraccoon@gmail.com */
/* vim: set ts=4 smarttab noet : */

package main

import (
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/hydrogen18/stoppableListener"
	"github.com/thesues/radoshttpd/rados"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	LOGPATH = "/var/log/wuzei.log"
	PIDFILE = "/var/run/wuzei.pid"
	slog    *log.Logger
)

type RadosDownloader struct {
	striper *rados.StriperPool
	soid    string
	offset  uint64
}

func (rd *RadosDownloader) Read(p []byte) (n int, err error) {
	count, err := rd.striper.Read(rd.soid, p, uint64(rd.offset))
	if count == 0 {
		return 0, io.EOF
	}
	rd.offset += uint64(count)
	return count, err
}

/* copied from  io package */
/* default buf is too small for inner web */
func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}
	buf := make([]byte, 4<<20)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return written, err
}

func main() {
	/* pid */
	if err := CreatePidfile(PIDFILE); err != nil {
		fmt.Printf("can not create pid file %s\n", PIDFILE)
		return
	}
	defer RemovePidfile(PIDFILE)

	/* log  */
	f, err := os.OpenFile(LOGPATH, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("failed to open log\n")
		return
	}
	defer f.Close()

	m := martini.Classic()
	slog = log.New(f, "[wuzei]", log.LstdFlags)
	m.Map(slog)

	conn, err := rados.NewConn("admin")
	if err != nil {
		return
	}
	conn.ReadConfigFile("/etc/ceph/ceph.conf")

	err = conn.Connect()
	if err != nil {
		return
	}
	defer conn.Shutdown()

	var wg sync.WaitGroup

	m.Get("/whoareyou", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("I AM WUZEI"))
	})
	m.Get("/(?P<pool>[A-Za-z0-9]+)/(?P<soid>[A-Za-z0-9-\\.]+)", func(params martini.Params, w http.ResponseWriter, r *http.Request) {
		wg.Add(1)
		defer wg.Done()

		poolname := params["pool"]
		soid := params["soid"]
		pool, err := conn.OpenPool(poolname)
		if err != nil {
			slog.Println("open pool failed")
			ErrorHandler(w, r, http.StatusNotFound)
			return
		}
		defer pool.Destroy()

		striper, err := pool.CreateStriper()
		if err != nil {
			slog.Println("open pool failed")
			ErrorHandler(w, r, http.StatusNotFound)
			return
		}
		defer striper.Destroy()

		filename := fmt.Sprintf("%s-%s", poolname, soid)
		size, err := striper.State(soid)
		if err != nil {
			slog.Println("failed to get object " + soid)
			ErrorHandler(w, r, http.StatusNotFound)
			return
		}

		rd := RadosDownloader{&striper, soid, 0}
		/* set content-type */
		/* Content-Type would be others */
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		/* set the stream */
		Copy(w, &rd)
	})

	originalListener, err := net.Listen("tcp", ":3000")
	sl, err := stoppableListener.New(originalListener)

	server := http.Server{}
	http.HandleFunc("/", m.ServeHTTP)

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT)
	signal.Notify(stop, syscall.SIGTERM)

	go func() {
		server.Serve(sl)
	}()

	slog.Printf("Serving HTTP\n")
	select {
	case signal := <-stop:
		slog.Printf("Got signal:%v\n", signal)
	}
	sl.Stop()
	slog.Printf("Waiting on server\n")
	wg.Wait()
	slog.Printf("Server shutdown\n")
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {
	switch status {
	case http.StatusNotFound:
		w.WriteHeader(status)
		w.Write([]byte("object not found"))
	}
}
