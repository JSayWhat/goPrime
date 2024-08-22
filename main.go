package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)


func chanIn[T any](done <-chan int, channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	fannedInStream := make(chan T)
	transfer := func(c <-chan T) {
		defer wg.Done()
		ticker := time.NewTicker(3 * time.Second) // Log every 3 seconds
		for i := range c {
			select {
			case <-done:
				log.Println("chanIn Go routine has finished.")
        	ticker.Stop()
				return
				case <-ticker.C:
                log.Println("chanIn Go routine is still running...")
			case fannedInStream <- i:
			}
		}
}
for _, c := range channels {
	wg.Add(1)
	go transfer(c)
}
go func() {
	wg.Wait()
	close(fannedInStream)
}()
return fannedInStream
}

func repeatFunc[T any, K any](done <-chan K, fn func() T, ctx context.Context) <- chan T {
	stream := make(chan T)
	go func() {
		defer close(stream)
		ticker := time.NewTicker(3 * time.Second) // Log every 3 seconds
		for {
			select {
			case <-done:
				 log.Println("RepeatFunc Go routine has finished.")
        	ticker.Stop()
				return	
			case <- ctx.Done():
				log.Println("RepeatFunc Go routine has finished.")
        	ticker.Stop()
				return
			case <-ticker.C:
                log.Println("RepeatFunc Go routine is still running...")
			case stream <- fn():
			}
		}   
	}()
	return stream
}

func take[T any, K any](done <-chan K, stream <-chan T, n int) <-chan T {
	takenStream := make(chan T)
	go func() {
		defer close(takenStream)
		for i := 1; i < n; i++ {
			select {
			case <-done:
				return
			case takenStream <- <-stream:
			}
		}
	}()
	return takenStream
}

func primeFinder(done, randIntStream <-chan int) <-chan int {
	isPrime := func(randomInt int) bool {
		if randomInt < 2 { // Ensure numbers less than 2 are not considered
			return false
		}
		for i := 2; i*i <= randomInt; i++ {
			if randomInt%i == 0 {
				return false
			}
		} 
		return true
	}
	primes := make(chan int)
	go func() {
		defer close(primes)
		ticker := time.NewTicker(3 * time.Second) // Log every 3 seconds
		for {
			select {
				case <-done:
					log.Println("Primes Go routine has finished.")
        	ticker.Stop()
					return
				case randomInt := <- randIntStream:
					log.Println("Primes Go routine is still running...")
					if isPrime(randomInt) {
						primes <- randomInt
				}
			}
		}
	}()
	return primes
}

// Helper function to count primes in a given range
func countPrimesInRange(min, max int) int {
	count := 0
	for num := min; num <= max; num++ {
		if isPrime(num) {
			count++
		}
		if count >= 100 {
			break
		}
	}
	log.Println("Total primes in range:", count)
	return count
}

// Function to check if a number is prime
func isPrime(num int) bool {
	if num < 2 {
		return false
	}
	for i := 2; i*i <= num; i++ {
		if num%i == 0 {
			return false
		}
	}
	return true
}

type Request struct {
    Min int `json:"min"`
    Max int `json:"max"`
		MaxPrimesCount int    `json:"max_primes_count"`
}

type Response struct {
    RandomNumber int `json:"random_prime_number"`
		Primes       []int `json:"primes,omitempty"`
}

func main() {
	http.HandleFunc("/", handleHTML)
	http.HandleFunc("/random", randomHandler)
  fmt.Println("Server is running on port 8080")
  err := http.ListenAndServe(":8080", nil)
	if err != nil {
			fmt.Println("Server failed:", err)
		}
}

func handleHTML(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func randomHandler(w http.ResponseWriter, r *http.Request) {
var req Request
err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
}
	if req.Min > req.Max {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "min cannot be greater than max"})
        return
    }

	min, max := req.Min, req.Max
	if max > 100000000 || max < min { // Check for valid range
        http.Error(w, "Invalid max value. Please enter a value between 10 and 100,000,000.", http.StatusBadRequest)
        return
    }

	// Ensure min is at least 1
	if min < 2 {
		min = 2
	}
	// Calculate the maximum possible primes in the range
	totalPrimes := countPrimesInRange(min, max)	

	// Validate max_primes_count
	maxPrimesCount := req.MaxPrimesCount
	if maxPrimesCount < 1 {
		maxPrimesCount = 1 // Set to minimum if less than 1
	} else if maxPrimesCount > 100 || max < min { // Check for valid range
		http.Error(w, "Invalid max count. Max count is 100.", http.StatusBadRequest)
		maxPrimesCount = int(math.Min(float64(totalPrimes), 100)) // Set to maximum if greater than 100 
	}

	// Track used numbers to avoid repeats
	usedNumbers := make(map[int]bool)
	var mutex sync.Mutex

	// Function to generate a unique random number
	generateUniqueRandomNumber := func(min, max int) int {
        mutex.Lock()
        defer mutex.Unlock()
        for {
            num := rand.Intn(max-min+1) + min
            if !usedNumbers[num] {
                usedNumbers[num] = true
                return num
            }
        }
    };

	randomNumber := generateUniqueRandomNumber(min, max)

	// Find primes in range 
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Set a timeout for the context
	defer cancel()
	done := make(chan int)
	defer close(done)

  randNumFetcher := func(min, max int) int { return generateUniqueRandomNumber(min, max) }
	
	randIntStream := repeatFunc(done, func() int { return randNumFetcher(min, max) }, ctx)

	// channel fanned out
	CPUCount := runtime.NumCPU()
	log.Println("Number of CPUs:", CPUCount)
	primeFinderChannels := make([]<-chan int, CPUCount)
	for i := 0; i < CPUCount; i++ {
		primeFinderChannels[i] = primeFinder(done, randIntStream)
	}

	// channel fanned in
	fannedInStream := chanIn(done, primeFinderChannels...)

	// collect primes
	primes := make([]int, 0)

	// Use maxPrimesCount to limit the number of primes
	for prime := range take(done, fannedInStream, maxPrimesCount) {
		primes = append(primes, prime)
	}

	// create response
	resp := Response{RandomNumber: randomNumber, Primes: primes} 

	// send response json
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
