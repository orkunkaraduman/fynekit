package fynekit

import (
	"context"
	"sync"
)

type Runner struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewRunner() *Runner {
	return NewRunnerWithContext(nil)
}

func NewRunnerWithContext(ctx context.Context) *Runner {
	r := &Runner{}
	ctxParent := context.Background()
	if ctx != nil {
		ctxParent = ctx
	}
	r.ctx, r.cancel = context.WithCancel(ctxParent)
	return r
}

func (r *Runner) do(fn func(context.Context), async bool) error {
	r.wg.Add(1)
	if e := r.ctx.Err(); e != nil {
		r.wg.Done()
		return e
	}
	f := func() {
		defer r.wg.Done()
		if fn != nil {
			fn(r.ctx)
		}
	}
	if async {
		go f()
	} else {
		f()
	}
	return nil
}

func (r *Runner) Do(fn func(context.Context)) error {
	return r.do(fn, false)
}

func (r *Runner) DoAsync(fn func(context.Context)) error {
	return r.do(fn, true)
}

func (r *Runner) Ctx() context.Context {
	return r.ctx
}

func (r *Runner) Cancel() {
	r.cancel()
}

func (r *Runner) Wait() {
	<-r.ctx.Done()
	r.wg.Wait()
}

func (r *Runner) Stop() {
	r.Cancel()
	r.Wait()
}
