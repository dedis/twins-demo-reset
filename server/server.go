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
// 1. Stop the agents
// 2. Replace the database with a snapshot
// 3. Start the agents
// 4. Reset the DARCs
func (s *service) process(ctx context.Context, errCh chan error) {
	// 1. Stop the agents
	err := s.stopAgents(ctx)
	if err != nil {
		errCh <- err
		return
	}

	// 2. Replace the databases
	s.replaceDatabase(ctx)
	if err != nil {
		errCh <- err
		return
	}

	// 3. Start the agents
	s.startAgents(ctx)
	if err != nil {
		errCh <- err
		return
	}

	// 4. Reset the DARCs
	s.resetDARCs(ctx)
	if err != nil {
		errCh <- err
		return
	}

	errCh <- nil
}

func (s *service) resetDARCs(ctx context.Context) error {
	nathDarcCmd := command{
		Name:      "bcadmin",
		Arguments: []string{"darc", "rule", "--darc", "4401770705c2140bb4d0e364f86b586bebc16c1faf768841794f43d420d781b2", "--rule", "'spawn:calypsoRead'", "--replace", "--identity", "'did:sov:Baqh3nz5QX3zVQ6RWTmQGr'"},
	}

	resetCommands := []command{nathDarcCmd}

	for _, resetCmd := range resetCommands {
		cmd := exec.CommandContext(ctx, resetCmd.Name, resetCmd.Arguments...)

		_, err := cmd.Output()
		if err != nil {
			return xerrors.Errorf("failed to reset DARC: %v", err)
		}
	}

	return nil
}

func (s *service) stopAgents(ctx context.Context) error {
	agentNames := []string{"researcher", "biobank", "dotnet"}
	for _, agent := range agentNames {
		_, err := exec.CommandContext(ctx, "pkill", "-SIGINT", agent).Output()
		if err != nil {
			return xerrors.Errorf("failed to execute stop agent command: %v", err)
		}
	}

	return nil
}

func (s *service) replaceDatabase(ctx context.Context) error {
	_, err := exec.CommandContext(ctx, "rm", "-rf",
		"/home/dedis/.indy_wallet/biobank",
		"/home/dedis/.indy_wallet/researcher",
		"/home/dedis/.indy_wallet/mediator.twins-project.org",
	).Output()

	if err != nil {
		return xerrors.Errorf("failed to remove databases: %v", err)
	}

	_, err = exec.CommandContext(ctx, "cp", "-R",
		"/home/dedis/snapshots/biobank",
		"/home/dedis/snapshots/researcher",
		"/home/dedis/snapshots/mediator.twins-project.org",
		"/home/dedis/.indy_wallet/",
	).Output()

	if err != nil {
		return xerrors.Errorf("failed to copy snapshots: %v", err)
	}

	return nil
}

type command struct {
	Dir       string
	Name      string
	Arguments []string
}

func (s *service) startAgents(ctx context.Context) error {
	researcherCmd := command{
		Dir:       "/home/dedis/twins",
		Name:      "./researcher",
		Arguments: []string{"agent", "--path", "~/.indy_wallet/researcher/peerstore", "--vdri-endpoint", "'http://localhost:4001'", "--did", "did:sov:FUp9R3oNxdWAMgB81A22ft", "--bcId", "'9832e2e66e1441b0f0da5011d50882cc49783b64af5371c6ac60b938f8a4e60c'", "--roster", "./roster.toml", "start"},
	}
	biobankCmd := command{
		Dir:       "/home/dedis/twins",
		Name:      "./biobank",
		Arguments: []string{"agent", "--path", "~/.indy_wallet/biobank/peerstore", "--vdri-endpoint", "'http://localhost:4001'", "--did", "did:sov:YKhCeTM8YsR8iQX16BaKdp", "start"},
	}
	mediatorCmd := command{
		Dir:       "/home/dedis/aries-mediator",
		Name:      "dotnet",
		Arguments: []string{"run", "--project", "MediatorAgent"},
	}

	commands := []command{researcherCmd, biobankCmd, mediatorCmd}

	for _, c := range commands {
		cmd := exec.CommandContext(ctx, c.Name, c.Arguments...)
		cmd.Dir = c.Dir

		err := cmd.Start()
		if err != nil {
			return xerrors.Errorf("failed to execute a start agent command: %v", err)
		}
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
