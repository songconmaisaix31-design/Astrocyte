// Package attentionruntime assembles original attention use cases at process
// entrypoints. It supplies no domain policy, source projection or scheduler.
package attentionruntime

import (
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func NewSleepService(db *sqlite.DB, dynamics app.DynamicsOptions, options app.SleepOptions) *app.SleepService {
	return app.NewSleepService(sqlite.NewSleepRepository(db), app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), dynamics), options)
}
