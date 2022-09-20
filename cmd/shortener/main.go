package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/danilovkiri/dk_go_url_shortener/internal/api/grpc/handlers"
	"github.com/danilovkiri/dk_go_url_shortener/internal/api/grpc/interceptors"
	pb "github.com/danilovkiri/dk_go_url_shortener/internal/api/grpc/proto"
	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/secretary/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/infile"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/inpsql"
	"google.golang.org/grpc"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildMetadata() {
	// print out build parameters
	switch buildVersion {
	case "":
		fmt.Printf("Build version: %s\n", "N/A")
	default:
		fmt.Printf("Build version: %s\n", buildVersion)
	}
	switch buildDate {
	case "":
		fmt.Printf("Build date: %s\n", "N/A")
	default:
		fmt.Printf("Build date: %s\n", buildDate)
	}
	switch buildCommit {
	case "":
		fmt.Printf("Build commit: %s\n", "N/A")
	default:
		fmt.Printf("Build commit: %s\n", buildCommit)
	}
}

func main() {
	// print out build parameters
	printBuildMetadata()
	// make a top-level file logger for logging critical errors
	flog, err := os.OpenFile(`server.log`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer flog.Close()
	mainlog := log.New(flog, `http `, log.LstdFlags|log.Lshortfile)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// add a waiting group
	wg := &sync.WaitGroup{}
	// set number of wg members to 1 (this will be the persistStorage goroutine)
	wg.Add(1)
	// get configuration
	cfg := config.NewDefaultConfiguration()
	err = cfg.Parse()
	if err != nil {
		mainlog.Fatal(err)
	}
	// initialize (or retrieve if present) storage, switch between "infile" and "inpsql" modules
	var errInit error
	var storageInit storage.URLStorage
	switch cfg.DatabaseDSN {
	case "":
		storageInit, errInit = infile.InitStorage(ctx, wg, cfg)
	default:
		storageInit, errInit = inpsql.InitStorage(ctx, wg, cfg)
	}
	if errInit != nil {
		mainlog.Fatal(errInit)
	}

	// switch between HTTP and GRPC protocols
	switch cfg.UseGRPC {
	case false:
		// initialize server
		server, err := rest.InitServer(ctx, cfg, storageInit)
		if err != nil {
			mainlog.Fatal(err)
		}
		// set a listener for os.Signal
		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		go func() {
			<-done
			mainlog.Print("Server shutdown attempted")
			ctxTO, cancelTO := context.WithTimeout(ctx, 5*time.Second)
			defer cancelTO()
			if err := server.Shutdown(ctxTO); err != nil {
				mainlog.Fatal("Server shutdown failed:", err)
			}
			cancel()
		}()
		// start up the server
		mainlog.Print("Server start attempted")
		if !cfg.EnableHTTPS {
			mainlog.Print("Using HTTP")
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				mainlog.Fatal(err)
			}
		} else {
			mainlog.Print("Using HTTPS")
			if err := server.ListenAndServeTLS("../../certs/localhost.crt", "../../certs/localhost.key"); err != nil && err != http.ErrServerClosed {
				mainlog.Fatal(err)
			}
		}
	case true:
		// initialize server
		server, err := handlers.InitServer(ctx, cfg, storageInit)
		if err != nil {
			mainlog.Fatal(err)
		}
		// set a listener for GRPC server
		listen, err := net.Listen("tcp", cfg.ServerAddress)
		if err != nil {
			mainlog.Fatal(err)
		}
		// initialize a secretary service
		secretaryService := secretary.NewSecretaryService(cfg)
		// initialize an interceptor service
		interceptorService := interceptors.NewAuthHandler(secretaryService, cfg)
		// create a new GRPC server
		s := grpc.NewServer(grpc.UnaryInterceptor(interceptorService.UnaryServerInterceptor()))
		// set a listener for os.Signal
		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		go func() {
			<-done
			mainlog.Print("Server shutdown attempted")
			s.GracefulStop()
			cancel()
		}()
		// register a service
		pb.RegisterShortenerServer(s, server)
		mainlog.Print("Server start attempted")
		mainlog.Print("Using GRPC")
		if err := s.Serve(listen); err != nil {
			mainlog.Fatal(err)
		}
	}

	// wait for goroutine in InitStorage to finish before exiting
	wg.Wait()
	mainlog.Print("Server shutdown succeeded")
}
