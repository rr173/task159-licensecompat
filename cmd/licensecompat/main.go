package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/rr173/task159-licensecompat/internal/httpapi"
	"github.com/rr173/task159-licensecompat/internal/service"
	"github.com/rr173/task159-licensecompat/internal/store"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	dbPath := flag.String("db", "licensecompat.db", "SQLite database path")
	address := flag.String("addr", ":8080", "HTTP address")
	smoke := flag.Bool("smoke-test", false, "run deterministic self-check and exit")
	flag.Parse()
	repository, err := store.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer repository.Close()
	app := service.New(repository)
	if *smoke {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.RunDemo(ctx); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(os.Stdout, "licensecompat smoke test passed")
		return
	}
	if _, err := app.Recover(context.Background()); err != nil {
		log.Fatal(err)
	}
	server := &http.Server{Addr: *address, Handler: httpapi.New(app).Handler(), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("licensecompat listening on %s", *address)
	log.Fatal(server.ListenAndServe())
}
