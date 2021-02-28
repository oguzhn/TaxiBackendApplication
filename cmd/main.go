package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/oguzhn/TaxiBackendApplication/authentication"
	"github.com/oguzhn/TaxiBackendApplication/controller"
	"github.com/oguzhn/TaxiBackendApplication/database"
)

func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	mongoclient, err := database.NewDatastore("mongodb+srv://user9281:sUvNzESn48M2AWk@assignmentcluster-doq69.mongodb.net/case?ssl=true&retryWrites=true", "case", "trips", ctx)
	if err != nil {
		log.Fatalln("cannot reach mongodb", err)
	}

	authenticator := authentication.NewHardCodedAuthenticator("token", "admin", "password")

	restController := controller.NewController(mongoclient, authenticator, authenticator)
	r := http.NewServeMux()
	r.Handle("/", restController.RegisterHandlers())

	server := &http.Server{
		Addr: ":8080",
		Handler: handlers.RecoveryHandler()(
			handlers.ProxyHeaders(
				handlers.LoggingHandler(os.Stdout,
					r,
				),
			),
		),
	}

	wg.Add(1)
	go func() {
		defer cancel()
		defer wg.Done()
		log.Println("Server started to listen")
		err := server.ListenAndServe()
		if err != nil {
			log.Println("server erred: ", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	select {
	case <-sigCh:
		cancel()
		log.Println("os signal caught ending")
	case <-ctx.Done():
		log.Println("context done")
	}

	sctx, sctxcancel := context.WithTimeout(context.Background(), time.Second)
	defer sctxcancel()
	err = server.Shutdown(sctx)
	if err != nil {
		log.Println("error closing server:", err)
	}
	wg.Wait()

}
