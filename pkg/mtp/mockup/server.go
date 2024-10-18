package mockup

import (
	"context"
	"github.com/malumar/zoha/pkg/bitmask"
	"github.com/malumar/zoha/pkg/mtp"
)

func Run(flags bitmask.Flag64, fn func(lmtpServer *mtp.Listener)) {
	lmtpServer := mtp.NewDefaultListenerExt("test", "testhost", flags)
	ctx := context.Background()

	if err := lmtpServer.Listen(); err != nil {

		return
	}

	defer lmtpServer.Stop()

	Wg.Add(1)
	go func() {
		lmtpServer.Run(ctx, new(SupervisorMockup), nil)
		Wg.Done()

	}()

	Wg.Add(1)
	go func() {
		fn(lmtpServer)
		Wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		}
	}
	ctx.Done()
	Wg.Wait()

}
