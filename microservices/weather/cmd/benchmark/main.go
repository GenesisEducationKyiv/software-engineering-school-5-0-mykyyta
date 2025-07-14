package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
	pb "weather/internal/proto"
)

const (
	concurrentLimit = 10
	nRequests       = 50000
	httpURL         = "http://localhost:8082/weather?city=Kyiv"
	grpcAddr        = "localhost:50051"
)

func main() {
	fmt.Printf("Starting benchmark with %d requests each...\n", nRequests)

	grpcDuration := benchmarkGRPC()
	fmt.Printf("gRPC finished in: %v\n", grpcDuration)

	httpDuration := benchmarkHTTP()
	fmt.Printf("HTTP finished in: %v\n", httpDuration)
}

func benchmarkGRPC() time.Duration {
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal("gRPC dial error:", err)
	}
	defer conn.Close()

	client := pb.NewWeatherServiceClient(conn)
	var wg sync.WaitGroup
	wg.Add(nRequests)

	start := time.Now()
	for i := 0; i < nRequests; i++ {
		go func() {
			defer wg.Done()
			_, err := client.GetWeather(context.Background(), &pb.WeatherRequest{City: "Kyiv"})
			if err != nil {
				log.Println("gRPC error:", err)
			}
		}()
	}
	wg.Wait()
	return time.Since(start)
}

func benchmarkHTTP() time.Duration {
	tr := &http.Transport{
		MaxIdleConns:        concurrentLimit,
		MaxIdleConnsPerHost: concurrentLimit,
		MaxConnsPerHost:     concurrentLimit,
		IdleConnTimeout:     30 * time.Second,
	}
	client := &http.Client{Transport: tr, Timeout: 5 * time.Second}

	jobs := make(chan struct{}, nRequests)
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < concurrentLimit; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for range jobs {
				resp, err := client.Get(httpURL)
				if err != nil {
					log.Printf("[HTTP wrk %d] %v", id, err)
					continue
				}
				_ = resp.Body.Close()
			}
		}(i)
	}

	for i := 0; i < nRequests; i++ {
		jobs <- struct{}{}
	}
	close(jobs)
	wg.Wait()

	return time.Since(start)
}
