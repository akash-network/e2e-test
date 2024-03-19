package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	port := "8080"
	e2eDir := "/var/lib/e2e-test"

	if val := os.Getenv("E2E_PORT"); val != "" {
		res, err := strconv.ParseUint(val, 10, 32)
		if err != nil || res > math.MaxUint16 {
			fmt.Printf("E2E_PORT contains invalid value\n")
			return
		}

		port = val
	}

	if val := os.Getenv("E2E_DIR"); val != "" {
		e2eDir = val
	}

	e2eDir = strings.TrimSuffix(e2eDir, "/")

	if err := os.MkdirAll(e2eDir, os.ModeDir); err != nil {
		fmt.Printf("couldn't create work dir: %s\n", err.Error())
		return
	}

	file, err := os.OpenFile(e2eDir+"/testfile", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}

	fmt.Printf("saving to file %s\n", e2eDir+"/testfile")

	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}

	if string(data) == "" {
		if _, err = file.WriteString("default"); err != nil {
			fmt.Printf("%s\n", err.Error())
		}

		_, _ = file.Seek(0, io.SeekStart)
	}

	var lock sync.Mutex

	shutdownCh := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())

	srv := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: newRouter(shutdownCh, file, &lock),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()

		err = srv.ListenAndServe()
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	go func() {
		defer wg.Done()

		select {
		case <-ctx.Done():
			break
		case <-shutdownCh:
			<-time.After(3 * time.Second)

			_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	}()

	fmt.Printf("app started\n")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	sig := <-ch

	cancel()
	fmt.Printf("service received signal: %s. shutting down\n", sig.String())

	go func() {
		_ = srv.Shutdown(context.Background())
	}()

	wg.Wait()
}

func newRouter(shChan chan<- struct{}, file *os.File, lock *sync.Mutex) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/kill",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(http.StatusOK)
			shChan <- struct{}{}
		}).
		Methods("GET")

	router.HandleFunc("/GET/value",
		func(resp http.ResponseWriter, request *http.Request) {
			defer lock.Unlock()
			lock.Lock()

			_, _ = file.Seek(0, io.SeekStart)

			data, _ := ioutil.ReadAll(file)

			_, _ = resp.Write(data)
		}).Methods("GET")

	router.HandleFunc("/SET/value",
		func(resp http.ResponseWriter, request *http.Request) {
			defer lock.Unlock()
			lock.Lock()

			_, _ = file.Seek(0, io.SeekStart)

			data, _ := ioutil.ReadAll(request.Body)

			_, _ = file.Write(data)
			_ = file.Sync()

			resp.WriteHeader(http.StatusOK)
		}).Methods("GET")

	return router
}
