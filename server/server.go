package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/sync/semaphore"
	"golang.org/x/xerrors"
)

type service struct {
	processing *semaphore.Weighted
}

func (s *service) reset() error {
	if !s.processing.TryAcquire(1) {
		return xerrors.Errorf("failed to process request since a previous request is running")
	}

	defer s.processing.Release(1)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	errCh := make(chan error)
	go s.process(ctx, errCh)

	select {
	case err := <-errCh:
		if err != nil {
			return xerrors.Errorf("failed to process: %v", err)
		}
		return nil
	case <-ctx.Done():
		return xerrors.Errorf("timed out")
	}
}

// process does the following steps in order:
// 1. Execute the stop script
// 2. Execute replace DB script
// 3. Execute the start script
// 4. Execute the reset DARC script
func (s *service) process(ctx context.Context, errCh chan error) {
	// 1. Stop the agents
	err := s.stopTwins(ctx)
	if err != nil {
		errCh <- err
		return
	}

	// 2. Replace the databases
	err = s.runReplaceDBScript(ctx)
	if err != nil {
		errCh <- err
		return
	}

	// 3. Start the agents
	err = s.startTwins(ctx)
	if err != nil {
		errCh <- err
		return
	}

	// We sleep for 20 seconds to allow the conodes to be up and running
	time.Sleep(20 * time.Second)

	// 4. Reset the DARCs
	err = s.runResetDARCScript(ctx)
	if err != nil {
		errCh <- err
		return
	}

	errCh <- nil
}

func runCmd(ctx context.Context, cmd *command) error {
	cmdInstance := exec.CommandContext(ctx, cmd.Name, cmd.Arguments...)
	if cmd.Dir != "" {
		cmdInstance.Dir = cmd.Dir
	}

	_, err := cmdInstance.Output()
	if err != nil {
		return xerrors.Errorf("failed to reset DARC: %v", err)
	}

	return nil
}

func (s *service) runResetDARCScript(ctx context.Context) error {
	resetCmd := &command{
		Dir:       "/home/dedis/bin",
		Name:      "sh",
		Arguments: []string{"./resetDarc"},
	}

	err := runCmd(ctx, resetCmd)
	if err != nil {
		log.Printf("failed to reset DARCs: %v", err)
		return xerrors.Errorf("failed to reset DARCs")
	}

	return nil
}

func (s *service) stopTwins(ctx context.Context) error {
	stopCmd := &command{
		Name:      "systemctl",
		Arguments: []string{"--user", "stop", "twins.target"},
	}

	err := runCmd(ctx, stopCmd)
	if err != nil {
		return xerrors.Errorf("failed to run command: %v", err)
	}

	return nil
}

func (s *service) runReplaceDBScript(ctx context.Context) error {
	replaceDB := &command{
		Dir:       "/home/dedis/bin",
		Name:      "sh",
		Arguments: []string{"./replaceDB"},
	}

	err := runCmd(ctx, replaceDB)
	if err != nil {
		log.Printf("failed to replace database with snapshots: %v", err)
		return xerrors.Errorf("failed to replace database with snapshots")
	}

	return nil
}

type command struct {
	Dir       string
	Name      string
	Arguments []string
}

func (s *service) startTwins(ctx context.Context) error {
	startCmd := &command{
		Name:      "systemctl",
		Arguments: []string{"--user", "start", "twins.target"},
	}

	err := runCmd(ctx, startCmd)
	if err != nil {
		log.Printf("failed to start twins.target: %v", err)
		return xerrors.Errorf("failed to start twins.target")
	}

	return nil
}

// Serve parses the CLI arguments and spawns an HTTP server
func Serve(c *cli.Context) error {
	port := c.Int("port")

	s := &service{
		processing: semaphore.NewWeighted(1),
	}
	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		err := s.reset()
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte("Reset complete"))
	})

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		fmt.Printf("Starting server at %s...\n", httpServer.Addr)
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start the server: %v", err)
		}
	}(wg)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	fmt.Printf("Stopping the server. Waiting for upto 60s\n")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := httpServer.Shutdown(ctx)
	if err != nil {
		log.Fatalf("failed to shutdown server: %v", err)
	}

	wg.Wait()
	return nil
}
