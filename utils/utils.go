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
	"os/signal"
	"os/user"
	"syscall"
	"strings"
	"strconv"
	"path/filepath"

	"github.com/grandcat/zeroconf"
	"github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/atotto/clipboard"
)

const (
	Domain string = "local."
	Service string = "_goshare._http._tcp"
	Default_Port = 8000
)

func CopyClipBoard() (string, error) {
	text, err := clipboard.ReadAll()
	if err != nil {
		return "", err
	}
	return string(text), nil
}

func getAbsolutePath(file string) string {
	if file[0] == '~' {
		User, err := user.Current()
		if err != nil {
			panic(err)
		}
		h := User.HomeDir
		file = h+file[1:]
	}
	absPath, err := filepath.Abs(file)
	if err != nil {
		panic(err)
	}
	return absPath 
}

func getFileName(files []string) (string, string) {
	file := getAbsolutePath(files[0])
	return file, "false"
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

func servicePresent(service_code string, results <- chan *zeroconf.ServiceEntry) bool {
	var service = service_code + "." + Service + "." + Domain
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
	})

	http.HandleFunc("/airshare", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "Text Sender")
	})

	http.HandleFunc("/text", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, text)
		fmt.Println("Resource accessed by:", strings.Split(r.RemoteAddr, ":")[0])
	})

	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		log.Fatal(err)
	}
}

func startFileServer(file string, port int, compress string) {
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		Openfile, err := os.Open("static/download.html")
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
	})

	http.HandleFunc("/airshare", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "File Sender")
	})

	http.HandleFunc("/download", func(rw http.ResponseWriter, r *http.Request) {
		Openfile, err := os.Open(file)
		defer Openfile.Close()
		if err != nil {
			http.Error(rw, "File not found", 404)
			return
		}

		FileHeader := make([]byte, 512)
		Openfile.Read(FileHeader)
		FileContentType := http.DetectContentType(FileHeader)

		FileStat, _ := Openfile.Stat()
		FileSize := strconv.FormatInt(FileStat.Size(), 10)

		rw.Header().Set("Content-Type", FileContentType)
		rw.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(file)+"\"; size="+FileSize)
		rw.Header().Set("Content-Length", FileSize)
		rw.Header().Set("airshare-compress", compress)
		if r.Method == "HEAD" {
			fmt.Println("Resource accessed by:", strings.Split(r.RemoteAddr, ":")[0])
		}
		if r.Method == "GET" {
			fmt.Println("Resource examined by:", strings.Split(r.RemoteAddr, ":")[0])
		}
		Openfile.Seek(0, 0)
		io.Copy(rw, Openfile)
		return
	})

	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		log.Fatal(err)
	}
}

func generateQRForCode(host string) {
	obj := qrcodeTerminal.New()
	obj.Get(host).Print()
}

func startService(service_code string, port int) {
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

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<- c
	fmt.Println("\nExit signal received, shutting down\n")
	service.Shutdown()
	os.Exit(0)
	select {}
}

func CreateTextService(service_code string, text string, port int) {
	if !CheckServicePresent(service_code) {
		go startTextServer(text, port)
		startService(service_code, port)
	} else {
		log.Fatal("A mDNS service with the same name is already running")
		os.Exit(-1)
	}
}

func CreateFileService(service_code string, files []string, port int) {
	if !CheckServicePresent(service_code) {
		file, compress := getFileName(files)
		go startFileServer(file, port, compress)
		startService(service_code, port)
	} else {
		log.Fatal("A mDNS service with the same name is already running")
		os.Exit(-1)
	}
}
