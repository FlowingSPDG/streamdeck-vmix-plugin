package pool

import (
	"context"
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
					v.logger.LogMessage(ctx, "failed to retry: %v", err)
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
		v.logger.LogMessage(ctx, "destination %s is probably deleted. Abort!", host)

		return nil
	}

	if vi.vmix.IsConnected() {
		v.logger.LogMessage(ctx, "destination %s is already connected. Abort!", host)

		return nil
	}

	v.logger.LogMessage(ctx, "Trying to connect destination %s ...", host)

	// 1.接続処理を行う
	if err := vi.vmix.Connect(vi.ctx, time.Second); err != nil {
		return xerrors.Errorf("failed to connect to vMix: %w", err)
	}
	v.logger.LogMessage(ctx, "connected to vMix destination %s. Register callbacks...", host)

	// 2: コールバックを登録する
	vi.vmix.OnVersion(func(vr *vmixtcp.VersionResponse) {
		// バージョン情報を受け取ったときの処理
		v.logger.LogMessage(ctx, "VersionResponse: %v", vr)
		if err := vi.vmix.Subscribe(vmixtcp.EventTally, ""); err != nil {
			panic(err)
		}
	})
	vi.vmix.OnTally(func(tr *vmixtcp.TallyResponse) {
		// タリーリクエストを受け取ったときの処理
		// (controllerに処理を委譲する)
		// → つまりControllerを内部に持つ必要がある？
		v.logger.LogMessage(ctx, "TallyResponse: %v", tr)
	})

	v.logger.LogMessage(ctx, "Registered callbacks %s", host)

	// 3: 接続に成功したら、vMixの状態を監視する
	if err := vi.vmix.Run(vi.ctx); err != nil {
		return xerrors.Errorf("failed to run vMix: %w", err)
	}

	return nil
}
