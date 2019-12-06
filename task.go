package app

import (
	"context"
	"github.com/rs/zerolog"
)

type Task interface {
	Run(ctx context.Context, logger zerolog.Logger)
}

type TaskFunc func(ctx context.Context, logger zerolog.Logger)

func (t TaskFunc) Run(ctx context.Context, logger zerolog.Logger) {
	t(ctx, logger)
}

type taskState struct {
	t Task
}

func (a *App) AddTask(t Task) {
	a.tasks = append(a.tasks, &taskState{t: t})
}

func (a *App) AddTaskFunc(t TaskFunc) {
	a.tasks = append(a.tasks, &taskState{t: t})
}

func (a *App) startTasks(ctx context.Context) {
	for _, t := range a.tasks {
		go func() {
			t.t.Run(ctx, a.logger)
			a.wg.Done()
		}()
		a.wg.Add(1)
	}
}
