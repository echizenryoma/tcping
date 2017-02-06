package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
)

func readIPs(ipPool chan string, cfg *Config) {
	file, err := os.Open(cfg.IP)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if len(ip) > 0 && net.ParseIP(ip) != nil {
			ipPool <- ip
		} else {
			log.Printf("%s is not ip\n", ip)
		}
	}
	defer close(ipPool)
}

func writeCSV(resultChan chan Result, wg *sync.WaitGroup, cfg *Config) {
	file, err := os.Create(cfg.Save)
	if err != nil {
		log.Println(err.Error())
	}
	defer file.Close()
	defer wg.Done()
	w := csv.NewWriter(file)
	for {
		select {
		case result, _ := <-resultChan:
			err = w.Write(result.String())
			log.Println(result.String())
			if err != nil {
				log.Println(err)
			}
			w.Flush()
			wg.Done()
		}
	}
}

func main() {
	log.SetOutput(os.Stderr)
	configFile := flag.String("config", "config.json", "Config file[JSON]")
	flag.Parse()
	cfg, err := readConfig(*configFile)
	if err != nil {
		log.Fatalln(err)
	}
	runtime.GOMAXPROCS(cfg.NumCPU)
	ipPool := make(chan string, cfg.Workers)
	resultChan := make(chan Result, cfg.Workers)
	go readIPs(ipPool, cfg)
	wg := &sync.WaitGroup{}
	for i := 0; i < cfg.Workers; i++ {
		go dial(ipPool, resultChan, wg, cfg)
	}
	wg.Add(1)
	go writeCSV(resultChan, wg, cfg)
	wg.Done()
	wg.Wait()
}
