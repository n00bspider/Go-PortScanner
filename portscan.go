package main

import (
	"fmt"
	"net"
	"runtime"
	"sort"
	"sync"
	"time"
)

func worker(ports <- chan int , results *[]int , addressInput *string , wg *sync.WaitGroup) {
	defer wg.Done()
	for p := range ports {
		address := fmt.Sprintf("%v:%v" , *addressInput , p)
		conn,err := net.Dial("tcp" , address)
		if err == nil {
			*results = append(*results, p)
			conn.Close()
		}
	}
}

func workerCount() int {
	if runtime.NumCPU() * 100 > 1000 {
		return 1000
	} else {
		return  runtime.NumCPU() * 100
	}
}

func main() {
	ports := make(chan int , workerCount())
	wg := sync.WaitGroup{}
	results := []int{}
	addressInput := new(string)
	startPort := new(int)
	endPort := new(int)
	

	key := 0
	for key != 3 {
		switch key {
		case 0:
			fmt.Printf("Enter a host : ")
			fmt.Scanln(addressInput)
			if fmt.Sprintf("%T",*addressInput) == "string" {
				host , err := net.LookupHost(*addressInput)
				if err == nil {
					*addressInput = host[0];
				}
			}
			if net.ParseIP(*addressInput) == nil {
				fmt.Printf("Enter a valid host.\n")
				key = 0
			} else {
				fmt.Printf("HOST : %v\n" , *addressInput)
				key = 1
			}
		case 1:
			fmt.Printf("Start from port (min : 1) : ")
			fmt.Scanln(startPort)
			if *startPort <= 0 {
				fmt.Printf("Enter a valid port number\n")
				key = 1
			} else {
				key = 2
			}
		case 2:
			fmt.Printf("Until port (max : 65535) : ")
			fmt.Scanln(endPort)
			if *endPort > 65535 {
				fmt.Printf("Enter a valid port number\n")
				key = 2
			} else {
				key = 3
			}
		}
	}

	startTime := time.Now()
	
	for w := 0; w < cap(ports); w++ {
		wg.Add(1)
		go worker(ports , &results , addressInput , &wg)
	}

	fmt.Printf("Scanning %v:(%v -> %v) ....\n\n" , *addressInput , *startPort , *endPort)

	for p := *startPort;p <= *endPort; p++ {
		ports <- p
	}
	close(ports)
	wg.Wait()

	fmt.Printf("Scanning took %v\n\n" , time.Since(startTime))

	if len(results) != 0 {
		sort.Ints(results)
		fmt.Printf("Open ports from %v to %v :\n" , *startPort , *endPort)
		for _,r := range results {
			fmt.Println("[*]",r)
		}
	} else {
		fmt.Printf("No ports from %v to %v\n" , *startPort , *endPort)
	}
}