# go-migration

[![Go Reference](https://pkg.go.dev/badge/github.com/andrianprasetya/go-migration.svg)](https://pkg.go.dev/github.com/andrianprasetya/go-migration)
[![Go Report Card](https://goreportcard.com/badge/github.com/andrianprasetya/go-migration)](https://goreportcard.com/report/github.com/andrianprasetya/go-migration)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A simple and lightweight database migration library for Go applications.

## Features

- 🚀 Simple and easy to use API
- 📦 Support for multiple database drivers
- 🔄 Up and down migrations
- 📝 SQL and Go-based migrations
- 🔍 Migration history tracking
- ⚡ Fast and lightweight
- 🧪 Well-tested and reliable

## Installation

```bash
go get github.com/andrianprasetya/go-migration
```

## Quick Start

### Basic Usage

```go
package main

import (
    "database/sql"
    "log"
    "os"
    
    "github.com/andrianprasetya/go-migration"
    _ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
    // Open database connection
    // Use environment variables for sensitive data
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create migration instance
    m := migration.New(db)
    
    // Run migrations
    if err := m.Up(); err != nil {
        log.Fatal(err)
    }
    
    log.Println("Migrations applied successfully!")
}
```

### Creating Migrations

Create a new migration file:

```bash
migration create create_users_table
```

This will create two files:
- `migrations/YYYYMMDDHHMMSS_create_users_table.up.sql`
- `migrations/YYYYMMDDHHMMSS_create_users_table.down.sql`

Example migration files:

**Up migration** (`YYYYMMDDHHMMSS_create_users_table.up.sql`):
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Down migration** (`YYYYMMDDHHMMSS_create_users_table.down.sql`):
```sql
DROP TABLE IF EXISTS users;
```

## API Reference

### Migration Methods

- `New(db *sql.DB) *Migrator` - Create a new migration instance
- `Up() error` - Run all pending up migrations
- `Down() error` - Run one down migration
- `DownAll() error` - Run all down migrations
- `Steps(n int) error` - Run n migrations (positive for up, negative for down)
- `Goto(version uint) error` - Migrate to a specific version
- `Version() (uint, error)` - Get current migration version

## Configuration

You can configure the migration directory and table name:

```go
m := migration.New(db,
    migration.WithDirectory("db/migrations"),
    migration.WithTableName("schema_migrations"),
)
```

## Supported Databases

- PostgreSQL
- MySQL
- SQLite
- SQL Server
- And any database with a standard `database/sql` driver

## Examples

See the [examples](examples/) directory for more usage examples:

- [PostgreSQL example](examples/postgres/)
- [MySQL example](examples/mysql/)
- [SQLite example](examples/sqlite/)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Setup

1. Clone the repository:
```bash
git clone https://github.com/andrianprasetya/go-migration.git
cd go-migration
```

2. Install dependencies:
```bash
go mod download
```

3. Run tests:
```bash
go test ./...
```

4. Run tests with coverage:
```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Guidelines

- Write clear and descriptive commit messages
- Add tests for new features
- Update documentation as needed
- Follow Go best practices and conventions
- Ensure all tests pass before submitting a PR

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

**Andrian Prasetya**
- GitHub: [@andrianprasetya](https://github.com/andrianprasetya)
- Email: andrianprasetya223@gmail.com

## Acknowledgments

- Inspired by other migration tools in the Go ecosystem
- Thanks to all contributors who have helped improve this project

## Support

If you find this project helpful, please consider giving it a ⭐️ on GitHub!