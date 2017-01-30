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
)

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 && net.ParseIP(line) != nil {
			lines = append(lines, line)
		} else {
			log.Printf("%s is not ip\n", line)
		}
	}
	return lines, scanner.Err()
}

func writeCSV(records []Result, cfg *Config) {
	if len(records) <= 0 {
		return
	}
	file, err := os.Create(cfg.Save)
	if err != nil {
		log.Println(err.Error())
	}
	defer file.Close()

	w := csv.NewWriter(file)
	err = w.Write(records[0].getHearders())
	if err != nil {
		log.Println(err)
	}
	for _, result := range records {
		err = w.Write(result.String())
		if err != nil {
			log.Println(err)
		}
	}
	w.Flush()
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
	ips, err := readLines(cfg.IP)
	if err != nil {
		log.Fatalln(err)
	}
	go func(ipPool chan string, ips []string) {
		for _, ip := range ips {
			ipPool <- ip
		}
		close(ipPool)
	}(ipPool, ips)
	for i := 0; i < cfg.Workers; i++ {
		go dial(ipPool, resultChan, cfg)
	}

	records := []Result{}
	count := 0
	for result := range resultChan {
		records = append(records, result)
		log.Println(result.String())
		count = count + 1
		if count >= len(ips) {
			break
		}
	}
	writeCSV(records, cfg)
}
