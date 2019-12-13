package app

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type testTask struct {
}

func (t testTask) Run(ctx context.Context, logger zerolog.Logger) {}

func TestAddTask(t *testing.T) {
	app := NewApp(NewAppConfig("MyApp"))

	task := &testTask{}

	app.AddTask(task)

	require.Len(t, app.tasks, 1)
	assert.Equal(t, task, app.tasks[0].t)
}

func TestAddTaskFunc(t *testing.T) {
	app := NewApp(NewAppConfig("MyApp"))

	task := func(ctx context.Context, logger zerolog.Logger) {}

	app.AddTaskFunc(task)

	require.Len(t, app.tasks, 1)
}

func TestStartTasks(t *testing.T) {
	app := NewApp(NewAppConfig("MyApp"))
	var ran bool

	app.AddTaskFunc(func(ctx context.Context, logger zerolog.Logger) {
		ran = true
	})

	app.startTasks(context.TODO())

	assert.Eventually(t, func() bool { return ran }, 1*time.Second, 10*time.Millisecond)
}
