// the run's lifecycle as the master reports it; undefined for masters that
// predate the field
export type RunLifecycle = 'running' | 'finished' | 'not_running' | undefined;
