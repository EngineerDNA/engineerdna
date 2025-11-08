package plugin

// MaxPluginEvents is the maximum number of events a plugin can process in a single operation
// to prevent memory exhaustion. This applies to sync, export, and analyze operations.
const MaxPluginEvents = 50000
