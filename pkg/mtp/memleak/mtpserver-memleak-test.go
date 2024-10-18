package main

import (
	"context"
	"github.com/malumar/zoha/pkg/mtp"
	"github.com/malumar/zoha/pkg/mtp/mockup"
)

// only for the purposes of longer tests regarding memory leaks, after 2 years of operation,
// I did not observe anything like this
func main() {
	lmtpServer := mtp.NewDefaultListenerExt(mtp.DefaultAddress, "test", mtp.ReceiveFromRemoteUsers|mtp.Authorization)
	ctx := context.Background()

	if err := lmtpServer.Listen(); err != nil {

		return
	}

	defer lmtpServer.Stop()

	mockup.Wg.Add(1)
	go func() {
		mockup.Wg.Done()
		lmtpServer.Run(ctx, new(mockup.SupervisorMockup), nil)

	}()
	for {
		select {
		case <-ctx.Done():
			return
		}
	}
	ctx.Done()
	mockup.Wg.Wait()

}
