package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/platon-p/kpodz3/payments/application/services"
	pb "github.com/platon-p/kpodz3/proto"
)

type OutboxWorker struct {
	accountRepo services.AccountRepo
}

func (w *OutboxWorker) Run(ctx context.Context) {
	t := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			evt, err := w.pullEvents(ctx)
			if err != nil {
				fmt.Println(fmt.Errorf("pull events: %w", err))
				continue
			}
			err = w.publishEvent(ctx, evt)
			if err != nil {
				fmt.Println(fmt.Errorf("publish events: %w", err))
				continue
			}
		}
	}
}

func (w *OutboxWorker) pullEvents(ctx context.Context) (*pb.Event, error) {
	evt := new(pb.Event)
	err := w.accountRepo.PopEvent(ctx, "order_success", evt)
	if err != nil {
		return nil, err
	}
	return evt, nil
}

func (w *OutboxWorker) publishEvent(ctx context.Context, evt *pb.Event) error {
	fmt.Println(evt)
	panic("not implemented")
}
