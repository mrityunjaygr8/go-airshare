package utils

import (
	"context"
	"io"
	"time"
	"log"
	"net"
	"net/http"
	"fmt"
	"os"
	"strings"
	"strconv"

	"github.com/grandcat/zeroconf"
	"github.com/Baozisoftware/qrcode-terminal-go"
)

const (
	Domain string = "local."
	Service string = "_goshare._http._tcp"
	Default_Port = 8000
)

func servicePresent(service_code string, results <- chan *zeroconf.ServiceEntry) bool {
	var service = service_code + "." + Service + "." + Domain
	for entry := range results {
		if entry.ServiceInstanceName() == service {
			return true
		}
	}
	return false
}

func getIPAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr).IP.String()
	return localAddr
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

func startTextServer(text string, port int) {
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		Openfile, err := os.Open("static/text.html")
		defer Openfile.Close()

		if err != nil {
			http.Error(rw, "File not found", 404)
			return
		}

		FileHeader := make([]byte, 512)
		Openfile.Read(FileHeader)

		rw.Header().Set("Content-Type", "text/html")
		Openfile.Seek(0, 0)
		io.Copy(rw, Openfile)
		fmt.Println("Resource accessed by:", strings.Split(r.RemoteAddr, ":")[0])
	})

	http.HandleFunc("/airshare", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "Text Sender")
	})

	http.HandleFunc("/text", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, text)
	})

	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		log.Fatal(err)
	}
}

func generateQRForCode(host string) {
	obj := qrcodeTerminal.New()
	obj.Get(host).Print()
}

func CreateService(service_code string, text string, port int) {
	if !CheckServicePresent(service_code) {
		go startTextServer(text, port)

		meta := []string{
			"version=0.1.0",
			"hello=world",
		}

		ips := []string{
			getIPAddress(),
		}

		service, err := zeroconf.RegisterProxy(
			service_code,
			Service,
			Domain,
			port,
			service_code,
			ips,
			meta,
			nil,
		)

		if err != nil {
			log.Fatal(err)
		}

		ip_addr := "http://"+ips[0]+":"+strconv.Itoa(port)
		service_addr := "http://"+service_code+".local"+":"+strconv.Itoa(port)

		fmt.Printf("Access the service at: %s or at: %s, press \"Ctrl+C\" to stop sharing\n", ip_addr, service_addr)
		generateQRForCode(ip_addr)
		defer service.Shutdown()

		select {}
	} else {
		log.Fatal("A mDNS service with the same name is already running")
		os.Exit(-1)
	}
}
