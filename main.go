package main

import (
	log "code.google.com/p/log4go"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"httpdns/handler"
	"httpdns/misc"
)

type Response struct {
	ErrNo  int         `json:"errno"`
	ErrMsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

func responseError(w http.ResponseWriter, errno int, errmsg string) {
	r := &Response{
		ErrNo:  errno,
		ErrMsg: errmsg,
	}
	b, _ := json.Marshal(r)
	fmt.Fprintf(w, string(b))
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	fmt.Fprintf(w, "%s", "OK")
}

func ResolveHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	res := r.URL.Query().Get("res")
	if res == "" {
		http.Error(w, "Bad Request", 400)
		return
	}

	c := &handler.Cache{}
	ip, host, err := c.Get(res)
	if err == nil {
		resp := Response{
			ErrNo:  10000,
			ErrMsg: "",
			Data:   map[string]string{"ip": ip, "host": host},
		}

		b, _ := json.Marshal(resp)
		fmt.Fprintf(w, string(b))
		return
	}

	ip, host, err = handler.DnsDecoder(res)
	if err != nil {
		log.Warn("[ResolveHandler] error: %v", err)
		responseError(w, 10001, fmt.Sprintf("%s", err))
		return
	}

	c.Set(res, ip, host)

	resp := Response{
		ErrNo:  10000,
		ErrMsg: "",
		Data:   map[string]string{"target": ip, "host": host},
	}

	b, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(b))
}

func main() {

	logConfigFile := flag.String("l", "./runtime/log4go.xml", "Log config file")
	configFile := flag.String("c", "./runtime/conf.json", "Config file")

	flag.Parse()

	log.LoadConfiguration(*logConfigFile)
	if err := misc.LoadConf(*configFile); err != nil {
		fmt.Printf("failed to load conf [%s]: (%s)", *configFile, err)
		os.Exit(1)
	}

	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)

	http.HandleFunc("/ping", PingHandler)
	http.HandleFunc("/d", ResolveHandler)
	err := http.ListenAndServe(misc.Conf.Addr, nil)
	if err != nil {
		fmt.Printf("failed to ListenAndServe: (%s)", err)
		os.Exit(1)
	}
}
