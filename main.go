package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Result struct {
	Port    int    `json:"port"`
	Open    bool   `json:"open"`
	Service string `json:"service"`
}

var services = map[int]string{
	20:  "ftp-data",
	21:  "ftp",
	22:  "ssh",
	23:  "telnet",
	25:  "smtp",
	53:  "dns",
	80:  "http",
	110: "pop3",
	143: "imap",
	443: "https",
	3306: "mysql",
	5432: "postgresql",
}

func getService(port int) string {
	if s, ok := services[port]; ok {
		return s
	}
	return "unknown"
}

func worker(host string, jobs <-chan int, results chan<- Result, timeout time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	for port := range jobs {
		address := fmt.Sprintf("%s:%d", host, port)

		conn, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			results <- Result{Port: port, Open: false}
			continue
		}

		conn.Close()

		results <- Result{
			Port:    port,
			Open:    true,
			Service: getService(port),
		}
	}
}

func parsePortRange(r string) (int, int) {
	parts := strings.Split(r, "-")
	start, _ := strconv.Atoi(parts[0])
	end, _ := strconv.Atoi(parts[1])
	return start, end
}

func writeJSON(results []Result, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func writeCSV(results []Result, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Port", "Open", "Service"})

	for _, r := range results {
		writer.Write([]string{
			strconv.Itoa(r.Port),
			strconv.FormatBool(r.Open),
			r.Service,
		})
	}

	return nil
}

func main() {

	// CLI FLAGS
	host := flag.String("host", "scanme.nmap.org", "Target host")
	portRange := flag.String("ports", "1-1024", "Port range (e.g. 1-1024)")
	workers := flag.Int("workers", 100, "Number of concurrent workers")
	timeout := flag.Duration("timeout", 300*time.Millisecond, "TCP timeout per port")
	jsonOut := flag.Bool("json", false, "Output JSON format")
	outFile := flag.String("out", "", "Save results to file")
	fast := flag.Bool("fast", false, "Faster scan (reduced timeout)")

	flag.Parse()

	if *fast {
		*timeout = 150 * time.Millisecond
	}

	startPort, endPort := parsePortRange(*portRange)
	totalPorts := endPort - startPort + 1

	fmt.Println("====================================")
	fmt.Println("Go Concurrent Port Scanner")
	fmt.Println("====================================")
	fmt.Printf("Target     : %s\n", *host)
	fmt.Printf("Ports      : %d-%d\n", startPort, endPort)
	fmt.Printf("Workers    : %d\n", *workers)
	fmt.Printf("Timeout    : %s\n", *timeout)
	fmt.Println()

	startTime := time.Now()

	jobs := make(chan int, totalPorts)
	results := make(chan Result, totalPorts)

	var wg sync.WaitGroup
	var scanned int64

	// Start workers
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(*host, jobs, results, *timeout, &wg)
	}

	// Send jobs
	go func() {
		for port := startPort; port <= endPort; port++ {
			jobs <- port
		}
		close(jobs)
	}()

	// Close results when done
	go func() {
		wg.Wait()
		close(results)
	}()

	var openPorts []Result

	// Progress printer (SAFE version)
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Printf("\rScanning... %d/%d ports", atomic.LoadInt64(&scanned), totalPorts)
			case <-done:
				return
			}
		}
	}()

	// Collect results
	for res := range results {

		atomic.AddInt64(&scanned, 1)

		if res.Open {
			openPorts = append(openPorts, res)
			fmt.Printf("\n[OPEN] %-5d/tcp  %s\n", res.Port, res.Service)
		}
	}

	// stop progress goroutine
	close(done)

	elapsed := time.Since(startTime)

	// Sort results
	sort.Slice(openPorts, func(i, j int) bool {
		return openPorts[i].Port < openPorts[j].Port
	})

	fmt.Println()
	fmt.Println("====================================")
	fmt.Println("Scan Complete")
	fmt.Println("====================================")
	fmt.Printf("Open Ports    : %d\n", len(openPorts))
	fmt.Printf("Ports Scanned : %d\n", totalPorts)
	fmt.Printf("Elapsed Time  : %s\n", elapsed)

	// Save output
	if *outFile != "" && *jsonOut {
		if err := writeJSON(openPorts, *outFile); err != nil {
			fmt.Println("Error writing JSON:", err)
		} else {
			fmt.Println("Saved JSON to", *outFile)
		}
	}

	if *outFile != "" && !*jsonOut {
		if err := writeCSV(openPorts, *outFile); err != nil {
			fmt.Println("Error writing CSV:", err)
		} else {
			fmt.Println("Saved CSV to", *outFile)
		}
	}
}