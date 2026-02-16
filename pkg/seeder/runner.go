package seeder

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// Logger defines a minimal logging interface for the seeder runner.
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// Runner resolves seeder dependencies and executes seeders in the correct order.
type Runner struct {
	registry *Registry
	db       *sql.DB
	logger   Logger
}

// NewRunner creates a new seeder Runner.
// The logger parameter may be nil, in which case logging is silently skipped.
func NewRunner(registry *Registry, db *sql.DB, logger Logger) *Runner {
	return &Runner{
		registry: registry,
		db:       db,
		logger:   logger,
	}
}

// RunAll executes all registered seeders in dependency-resolved order.
// If any seeder fails, execution stops and the error is returned.
func (r *Runner) RunAll() error {
	all := r.registry.GetAll()
	if len(all) == 0 {
		return nil
	}

	order, err := r.resolveAllOrder(all)
	if err != nil {
		return err
	}

	for _, name := range order {
		s := all[name]
		r.logInfo("Running seeder: %s", name)
		if err := s.Run(r.db); err != nil {
			r.logError("Seeder %s failed: %v", name, err)
			return fmt.Errorf("seeder %q: %w", name, err)
		}
		r.logInfo("Seeder %s completed", name)
	}

	return nil
}

// Run executes a specific seeder and its transitive dependencies in order.
// Returns ErrSeederNotFound if the named seeder is not registered.
func (r *Runner) Run(name string) error {
	if _, err := r.registry.Get(name); err != nil {
		return err
	}

	order, err := r.resolveOrder(name)
	if err != nil {
		return err
	}

	all := r.registry.GetAll()
	for _, sName := range order {
		s := all[sName]
		r.logInfo("Running seeder: %s", sName)
		if err := s.Run(r.db); err != nil {
			r.logError("Seeder %s failed: %v", sName, err)
			return fmt.Errorf("seeder %q: %w", sName, err)
		}
		r.logInfo("Seeder %s completed", sName)
	}

	return nil
}

// resolveOrder resolves the execution order for a single seeder and its
// transitive dependencies using DFS-based topological sort.
func (r *Runner) resolveOrder(name string) ([]string, error) {
	all := r.registry.GetAll()

	var order []string
	visited := make(map[string]bool)    // permanent mark
	inProgress := make(map[string]bool) // temporary mark (cycle detection)
	var path []string                   // current DFS path for cycle reporting

	var visit func(n string) error
	visit = func(n string) error {
		if visited[n] {
			return nil
		}
		if inProgress[n] {
			// Build cycle description from path
			cycle := buildCyclePath(path, n)
			return fmt.Errorf("%s: %w", cycle, ErrCircularDependency)
		}

		inProgress[n] = true
		path = append(path, n)

		s, ok := all[n]
		if !ok {
			return fmt.Errorf("dependency %q: %w", n, ErrSeederNotFound)
		}

		if ds, ok := s.(DependentSeeder); ok {
			deps := ds.DependsOn()
			// Sort dependencies for deterministic order
			sort.Strings(deps)
			for _, dep := range deps {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}

		path = path[:len(path)-1]
		inProgress[n] = false
		visited[n] = true
		order = append(order, n)
		return nil
	}

	if err := visit(name); err != nil {
		return nil, err
	}

	return order, nil
}

// resolveAllOrder resolves the execution order for all seeders using
// DFS-based topological sort.
func (r *Runner) resolveAllOrder(all map[string]Seeder) ([]string, error) {
	var order []string
	visited := make(map[string]bool)
	inProgress := make(map[string]bool)
	var path []string

	var visit func(n string) error
	visit = func(n string) error {
		if visited[n] {
			return nil
		}
		if inProgress[n] {
			cycle := buildCyclePath(path, n)
			return fmt.Errorf("%s: %w", cycle, ErrCircularDependency)
		}

		inProgress[n] = true
		path = append(path, n)

		s, ok := all[n]
		if !ok {
			return fmt.Errorf("dependency %q: %w", n, ErrSeederNotFound)
		}

		if ds, ok := s.(DependentSeeder); ok {
			deps := ds.DependsOn()
			sort.Strings(deps)
			for _, dep := range deps {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}

		path = path[:len(path)-1]
		inProgress[n] = false
		visited[n] = true
		order = append(order, n)
		return nil
	}

	// Sort names for deterministic iteration order
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return order, nil
}

// buildCyclePath constructs a human-readable cycle description.
func buildCyclePath(path []string, target string) string {
	// Find where the cycle starts in the path
	start := -1
	for i, n := range path {
		if n == target {
			start = i
			break
		}
	}
	if start == -1 {
		return target + " -> " + target
	}
	cycle := append(path[start:], target)
	return strings.Join(cycle, " -> ")
}

func (r *Runner) logInfo(msg string, args ...interface{}) {
	if r.logger != nil {
		r.logger.Info(msg, args...)
	}
}

func (r *Runner) logError(msg string, args ...interface{}) {
	if r.logger != nil {
		r.logger.Error(msg, args...)
	}
}
