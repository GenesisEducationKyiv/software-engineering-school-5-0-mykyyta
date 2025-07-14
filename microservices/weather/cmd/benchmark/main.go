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
	concurrentLimit = 50
	nRequests       = 1000
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
	client := &http.Client{Timeout: 3 * time.Second}
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrentLimit)

	start := time.Now()
	for i := 0; i < nRequests; i++ {
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			resp, err := client.Get(httpURL)
			if err != nil {
				log.Println("HTTP error:", err)
				return
			}
			_ = resp.Body.Close()
		}()
	}
	wg.Wait()
	return time.Since(start)
}
