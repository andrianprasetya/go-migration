package seeder

import "database/sql"

// Seeder defines the contract for a database seeder.
// Implementations populate database tables with data.
type Seeder interface {
	Run(db *sql.DB) error
}

// DependentSeeder extends Seeder with dependency declaration.
// DependsOn returns the names of seeders that must run before this one.
type DependentSeeder interface {
	Seeder
	DependsOn() []string
}
