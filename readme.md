# Go Concurrent Port Scanner

A high-performance TCP connect port scanner written in Go using a worker pool concurrency model.  
This tool scans a target host for open TCP ports and supports CLI configuration, service detection, and exportable results.

---

## 🚀 Features

- Concurrent TCP scanning using worker pools
- Configurable number of workers
- Adjustable timeout per connection
- Fast scan mode (reduced timeout)
- Service name detection (common ports)
- Progress indicator during scan
- JSON and CSV export support
- Clean CLI interface

---

## 🧠 How it works

This scanner uses a worker pool architecture:

Main Goroutine
  ├── Job Channel (ports)
  ├── Worker Goroutines (TCP scanners)
  └── Result Channel (scan results)

Each worker:
1. Receives a port from the job queue
2. Attempts a TCP connection using net.DialTimeout
3. Returns whether the port is open

---

## ⚙️ Usage

### Basic scan
go run . -host scanme.nmap.org

### Custom port range
go run . -host scanme.nmap.org -ports 1-2000

### Increase concurrency
go run . -host scanme.nmap.org -workers 200

### Faster scan mode
go run . -host scanme.nmap.org -fast

### Save results to file

CSV:
go run . -host scanme.nmap.org -out results.csv

JSON:
go run . -host scanme.nmap.org -json -out results.json

---

## 📊 Example Output

====================================
Go Concurrent Port Scanner
====================================
Target     : scanme.nmap.org
Ports      : 1-1024
Workers    : 100
Timeout    : 300ms

[OPEN] 22   /tcp  ssh
[OPEN] 80   /tcp  http

Scanning... 1024/1024 ports

====================================
Scan Complete
====================================
Open Ports    : 2
Ports Scanned : 1024
Elapsed Time  : 1.7s

---

## ⚡ Technical Details

- Worker pool concurrency model
- Buffered channels for job distribution
- net.DialTimeout TCP connect scanning
- Atomic counters for safe progress tracking

---

## ⚠️ Notes

- Only scan systems you own or are authorized to test
- Some ports may appear closed due to timeout limits
- Fast mode may introduce false negatives

---

## 🧪 Test Target

scanme.nmap.org

Provided by Nmap for educational purposes.
