# migy Database migration helper for MySQL

migy is a simple, standalone tool for managing SQL migrations on MySQL.
It generates up/down migration SQL files, checks their reversibility, and works entirely without requiring a live database connection.
Ideal for teams and individuals who want fine-grained control over their migrations using pure SQL.

Note: This program is still under development.

## Installation

```cmd
go install github.com/makiuchi-d/migy@latest
```


## Usage

migy manages SQL migrations for MySQL with a focus on file-based workflows and SQL consistency.

- `init` generates an initial SQL file for the database schema.
- `create` generates a new pair of up/down migration files.
- Edit the generated SQL files manually as needed.
- `check` verifies that a pair of up/down migrations are reversible.
- `snapshot` creates a SQL snapshot representing the database state at a specific migration point.

```shell
$ migy --help
A standalone database migration helper for MySQL

Usage:
  migy [command]

Available Commands:
  check       Check if an up/down migration pair is reversible
  completion  Generate the autocompletion script for the specified shell
  create      Create a new pair of up/down SQL migration files
  help        Help about any command
  init        Generate the initial migration SQL file
  snapshot    Generate a SQL snapshot at the specified migration point
  version     Show version

Flags:
  -d, --dir string   directory with migration files (default ".")
  -h, --help         help for migy
  -q, --quit         quit stdout

Use "migy [command] --help" for more information about a command.
```
