package scron

var (
	defaultCron = New()
)

// AddSingleton adds a func to the Cron to be run on the given schedule.
// The spec is parsed using the time zone of this Cron instance as the default.
// An opaque ID is returned that can be used to later remove it.
func AddSingleton(spec string, cmd func(), cmdName string) (EntryID, error) {
	return defaultCron.AddJob(spec, FuncJob(cmd), cmdName)
}

// Add adds a func to the Cron to be run on the given schedule.
// The spec is parsed using the time zone of this Cron instance as the default.
// An opaque ID is returned that can be used to later remove it.
func Add(spec string, cmd func(), cmdName string) (EntryID, error) {
	return defaultCron.AddJob(spec, FuncJob(cmd), cmdName)
}

// Entries return all timed tasks as slice.
func Entries() []Entry {
	return defaultCron.Entries()
}

// Start all timed tasks as slice.
func Start() {
	defaultCron.Start()
	return
}

// Run all timed tasks as slice
func Run() {
	defaultCron.Run()
	return
}

// Remove all timed tasks as slice.
func Remove(id EntryID) {
	defaultCron.Remove(id)
	return
}

// RemoveByName all timed tasks as slice.
func RemoveByName(name string) {
	defaultCron.RemoveByName(name)
	return
}

// Stop caller can wait for running jobs to complete.
func Stop() {
	defaultCron.Stop()
	return
}
