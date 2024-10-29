package pool

import (
	"context"
	"log"
	"time"

	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
	"golang.org/x/xerrors"
)

func (v *vmixPool) startRetry(ctx context.Context, host string) {
	v.logger.LogMessage(ctx, "start retry for %s", host)
	go func() {
		// 一定間隔で接続を試みる
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				if err := v.retry(ctx, host); err != nil {
					log.Printf("failed to retry: %v", err)
				}
			}
		}
	}()

}

func (v *vmixPool) retry(ctx context.Context, host string) error {
	v.logger.LogMessage(ctx, "retrying for %s", host)

	// vMixのインスタンスが削除されている場合、再接続処理を行わない
	vi, ok := v.vs.Load(host)
	if !ok {
		return nil
	}

	if vi.vmix.IsConnected() {
		return nil
	}

	// 1.接続処理を行う
	if err := vi.vmix.Connect(vi.ctx, time.Second); err != nil {
		return xerrors.Errorf("failed to connect to vMix: %w", err)
	}
	// 2: コールバックを登録する
	vi.vmix.OnTally(func(tr *vmixtcp.TallyResponse) {
		// タリーリクエストを受け取ったときの処理
		// (controllerに処理を委譲する)
		// → つまりControllerを内部に持つ必要がある？
		log.Println("TallyResponse: ", tr)
	})

	// 3: 接続に成功したら、vMixの状態を監視する
	if err := vi.vmix.Run(vi.ctx); err != nil {
		return xerrors.Errorf("failed to run vMix: %w", err)
	}

	return nil
}
