package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc/credentials/insecure"

	pb "weather/internal/proto"

	"google.golang.org/grpc"
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
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("gRPC dial error:", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

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
	ctx := context.Background()

	start := time.Now()
	for i := 0; i < nRequests; i++ {
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, httpURL, nil)
			if err != nil {
				log.Println("Error creating HTTP request:", err)
				return
			}

			resp, err := client.Do(req)
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
