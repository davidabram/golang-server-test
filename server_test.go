package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
)

var ts *httptest.Server
var serverShutdown func()

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World")
}

func goodbyeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Goodbye World")
}

func TestMain(m *testing.M) {
	var wg sync.WaitGroup
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/goodbye", goodbyeHandler)

	ts = httptest.NewServer(mux)

	_, cancel := context.WithCancel(context.Background())
	serverShutdown = cancel

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		fmt.Println("Received shutdown signal")
		cancel()
	}()

	exitCode := m.Run()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ts.Close()
	}()

	wg.Wait()

	os.Exit(exitCode)
}

func TestHelloHandler(t *testing.T) {
	resp, err := http.Get(ts.URL + "/hello")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	expected := "Hello World\n"
	if string(body) != expected {
		t.Errorf("Expected response body %q, got %q", expected, string(body))
	}
}

func TestGoodbyeHandler(t *testing.T) {
	resp, err := http.Get(ts.URL + "/goodbye")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	expected := "Goodbye World\n"
	if string(body) != expected {
		t.Errorf("Expected response body %q, got %q", expected, string(body))
	}
}
