package relay

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/bloXroute-Labs/bdn-operations-relay/config"
	"github.com/bloXroute-Labs/bdn-operations-relay/relay/server"

	"golang.org/x/sync/errgroup"
)

func Run(cfg *config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	eg, gCtx := errgroup.WithContext(ctx)

	s, err := server.NewServer(gCtx, cfg)
	if err != nil {
		return err
	}

	eg.Go(func() error {
		return s.Start(gCtx)
	})

	<-gCtx.Done()

	s.Shutdown()

	return eg.Wait()
}
