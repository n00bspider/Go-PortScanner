# TCP Port Scanner

A concurrent TCP port scanner written in Go. Uses a worker pool of goroutines to scan port ranges fast without overwhelming the target or the local OS with too many open file descriptors.

---

## Features

- Concurrent scanning via a configurable goroutine worker pool
- Accepts IPv4, IPv6, and domain names (with DNS resolution)
- Configurable port range
- Clean sorted output — open ports listed in order regardless of which worker found them first
- Cross-platform builds (Linux, macOS, Windows)

---

## Usage

Run the binary and follow the prompts:

```
$ ./portscanner

Enter host (IPv4, IPv6, or domain): scanme.nmap.org
Enter start port: 1
Enter end port: 1024

Resolved scanme.nmap.org → 45.33.32.156

Scanning 45.33.32.156 ports 1–1024 with 200 workers...

[+] Port 22   open
[+] Port 80   open

Scan complete. 2 open ports found in 1.24s
```

---

## How It Works

### Worker Pool

Spawning one goroutine per port doesn't scale — scanning ports 1–65535 would create 65,535 goroutines and likely exhaust the system's file descriptor limit. Instead, a fixed pool of `N` worker goroutines is created upfront. A single jobs channel is fed with port numbers; workers pick up jobs as they finish, keeping concurrency bounded.

```
main goroutine
  │
  ├─ sends port numbers → [jobs channel]
  │
  ├─ worker 1 ──┐
  ├─ worker 2 ──┤──── reads from jobs, dials TCP, sends result → [results channel]
  ├─ worker 3 ──┤
  └─ worker N ──┘
                       │
              results collector goroutine
                       │
                  collects open ports, signals done
```

Each worker calls `net.DialTimeout` with a short timeout. A successful dial means the port is open (TCP three-way handshake completed). A refused connection or timeout means closed or filtered.

### DNS Resolution

If the user provides a hostname instead of an IP, `net.LookupHost` resolves it before scanning begins. The resolved IP is shown so the user knows exactly what is being scanned.

### Result Ordering

Workers finish in arbitrary order depending on which ports respond fastest. Results are collected into a slice and sorted numerically before printing, so the output is always clean and readable regardless of concurrency.

---

## Running It

Pre-built binaries are in the `Builds/` folder.

**Linux / macOS**
```bash
chmod +x ./Builds/portscanner-linux
./Builds/portscanner-linux
```

**Windows**
```
.\Builds\portscanner-windows.exe
```

### Build from source

Requires Go 1.21+

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o Builds/portscanner-linux ./code/

# macOS
GOOS=darwin GOARCH=amd64 go build -o Builds/portscanner-macos ./code/

# Windows
GOOS=windows GOARCH=amd64 go build -o Builds/portscanner-windows.exe ./code/
```

---

## What I Learned

The first version just launched a goroutine per port. It worked on small ranges but fell apart on anything above a few hundred ports — the OS started refusing connections because too many file descriptors were open simultaneously. That failure led directly to the worker pool design, which was a useful lesson in why bounded concurrency matters in real network tools.

Sorting results also wasn't obvious at first. A naive approach of printing as results came in produced output in random order depending on which port responded fastest. Collecting everything first and sorting made the output actually useful.

The DNS step taught me that `net.LookupHost` can return multiple addresses (e.g. when a domain has both A and AAAA records), and you have to decide which one to use — right now the scanner takes the first result, which is a simplification worth noting.

## Disclaimer

Only scan hosts you own or have explicit permission to scan. Unauthorized port scanning may be illegal in your jurisdiction.
