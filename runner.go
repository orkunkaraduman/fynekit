package fynekit

import (
	"context"
	"sync"
)

type runner struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func newRunner() *runner {
	return newRunnerWithContext(nil)
}

func newRunnerWithContext(ctx context.Context) *runner {
	r := &runner{}
	ctxParent := context.Background()
	if ctx != nil {
		ctxParent = ctx
	}
	r.ctx, r.cancel = context.WithCancel(ctxParent)
	return r
}

func (r *runner) run(fn func(context.Context), async bool) error {
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

func (r *runner) Run(fn func(context.Context)) error {
	return r.run(fn, false)
}

func (r *runner) RunAsync(fn func(context.Context)) error {
	return r.run(fn, true)
}

func (r *runner) Ctx() context.Context {
	return r.ctx
}

func (r *runner) Cancel() {
	r.cancel()
}

func (r *runner) Wait() {
	<-r.ctx.Done()
	r.wg.Wait()
}

func (r *runner) Stop() {
	r.Cancel()
	r.Wait()
}
