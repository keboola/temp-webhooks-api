// nolint: gocritic
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/keboola/temp-webhooks-api/internal/pkg/env"
	"github.com/keboola/temp-webhooks-api/internal/pkg/webhooks/api/gen/webhooks"
	"github.com/keboola/temp-webhooks-api/internal/pkg/webhooks/api/service"
)

func main() {
	// Flags.
	httpHostF := flag.String("http-host", "0.0.0.0", "HTTP host")
	httpPortF := flag.String("http-port", "8888", "HTTP port")
	debugF := flag.Bool("debug", false, "Log request and response bodies")
	flag.Parse()

	// Setup logger.
	logger := log.New(os.Stderr, "[templatesApi][server]", 0)

	// Envs.
	envs, err := env.FromOs()
	if err != nil {
		logger.Println("cannot load envs: " + err.Error())
		os.Exit(1)
	}

	// Start server
	start(*httpHostF, *httpPortF, *debugF, logger, envs)
}

func start(host, port string, debug bool, logger *log.Logger, envs *env.Map) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize the service.
	svc := service.New(envs)

	// Wrap the services in endpoints that can be invoked from other services
	// potentially running in different processes.
	endpoints := webhooks.NewEndpoints(svc)

	// Create channel used by both the signal handler and server goroutines
	// to notify the main goroutine when to stop the server.
	errCh := make(chan error)

	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the services to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errCh <- fmt.Errorf("%s", <-c)
	}()

	// Create server URL.
	serverUrl := &url.URL{Scheme: "http", Host: net.JoinHostPort(host, port)}

	// Start HTTP server.
	var wg sync.WaitGroup
	handleHTTPServer(ctx, &wg, serverUrl, endpoints, errCh, logger, debug)

	// Wait for signal.
	logger.Printf("exiting (%v)", <-errCh)

	// Send cancellation signal to the goroutines.
	cancel()

	// Wait for goroutines.
	wg.Wait()
	logger.Println("exited")
}
