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
	mongoUri := os.Getenv("MONGO_URI")

	mongoclient, err := database.NewDatastore(mongoUri, "case", "trips", ctx)
	if err != nil {
		log.Fatalln("cannot reach mongodb", err)
	}

	token := os.Getenv("TOKEN")
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	authenticator := authentication.NewHardCodedAuthenticator(token, username, password)

	restController := controller.NewController(mongoclient, authenticator, authenticator)
	r := http.NewServeMux()
	r.Handle("/", restController.RegisterHandlers())
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr: ":" + port,
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
