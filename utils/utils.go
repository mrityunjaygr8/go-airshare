package utils

import (
	"context"
	"time"
	"log"
	"net/http"
	"fmt"
	"os"
	"strconv"

	"github.com/grandcat/zeroconf"
)

const (
	Domain string = "local."
	Service string = "_testimg._tcp"
	Default_Port = 8000
)

func servicePresent(service_code string, results <- chan *zeroconf.ServiceEntry) bool {
	var service = service_code + "." + Service + "." + Domain
	fmt.Println(service)
	for entry := range results {
		if entry.ServiceInstanceName() == service {
			return true
		}
	}
	return false
}

func CheckServicePresent(service_code string) bool {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatal(err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = resolver.Lookup(ctx, service_code, Service, Domain, entries)
	if err != nil {
		log.Fatalln("Failed to lookup:", err.Error())
	}

	a := servicePresent(service_code, entries)
	<- ctx.Done()
	return a
}

func startService(port int) {
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "Hello World")
	})

	log.Println("starting http service ...")
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		log.Fatal(err)
	}
}

func CreateService(service_code string, port int) {
	log.Println(port)
	if !CheckServicePresent(service_code) {
		go startService(port)

		meta := []string{
			"version=0.1.0",
			"hello=world",
		}

		service, err := zeroconf.Register(
			service_code,
			Service,
			Domain,
			port,
			meta,
			nil,
		)

		if err != nil {
			log.Fatal(err)
		}

		defer service.Shutdown()

		select {}
	} else {
		log.Fatal("A mDNS service with the same name is already running")
		os.Exit(-1)
	}
}
