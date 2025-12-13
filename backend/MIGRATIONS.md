# Database Migrations

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema management.

## Overview

Migrations are stored in the `migrations/` directory with the following naming convention:
```
{version}_{description}.up.sql   # Forward migration
{version}_{description}.down.sql # Rollback migration
```

Example:
```
migrations/
├── 000001_create_isos_table.up.sql
└── 000001_create_isos_table.down.sql
```

## Automatic Migrations

Migrations run automatically when the application starts. The database will be migrated to the latest version on startup.

## Creating New Migrations

### 1. Manual Creation

Create two files with the next version number:

**migrations/000002_add_column_example.up.sql:**
```sql
ALTER TABLE isos ADD COLUMN new_field TEXT DEFAULT '';
```

**migrations/000002_add_column_example.down.sql:**
```sql
ALTER TABLE isos DROP COLUMN new_field;
```

### 2. Using migrate CLI (Optional)

Install the migrate CLI tool:
```bash
go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Create a new migration:
```bash
migrate create -ext sql -dir migrations -seq add_column_example
```

This creates:
- `migrations/000002_add_column_example.up.sql`
- `migrations/000002_add_column_example.down.sql`

## Migration Best Practices

### DO:
- ✅ Keep migrations small and focused
- ✅ Test both up and down migrations
- ✅ Use transactions where possible (SQLite supports DDL transactions)
- ✅ Write idempotent migrations when possible (`IF NOT EXISTS`, `IF EXISTS`)
- ✅ Include meaningful descriptions in filenames

### DON'T:
- ❌ Modify existing migration files after they've been deployed
- ❌ Delete migration files
- ❌ Skip version numbers
- ❌ Use database-specific syntax that won't work with SQLite

## Manual Migration Management (Development Only)

For development and testing, you can manually manage migrations:

### Check Current Version
```bash
migrate -database "sqlite://data/db/isos.db" -path migrations version
```

### Migrate Up
```bash
migrate -database "sqlite://data/db/isos.db" -path migrations up
```

### Migrate Down (Rollback)
```bash
migrate -database "sqlite://data/db/isos.db" -path migrations down
```

### Migrate to Specific Version
```bash
migrate -database "sqlite://data/db/isos.db" -path migrations goto 1
```

### Force Version (Use with Caution)
If migrations get into a dirty state:
```bash
migrate -database "sqlite://data/db/isos.db" -path migrations force VERSION
```

## Migration Schema

golang-migrate creates a `schema_migrations` table to track applied migrations:

```sql
CREATE TABLE schema_migrations (
    version bigint NOT NULL PRIMARY KEY,
    dirty boolean NOT NULL
);
```

- `version`: The current migration version
- `dirty`: Indicates if a migration failed mid-execution

## Troubleshooting

### Migration Failed (Dirty State)

If a migration fails partway through, the database will be marked as "dirty":

1. Check the `schema_migrations` table:
   ```sql
   SELECT * FROM schema_migrations;
   ```

2. Manually fix the issue

3. Force the version:
   ```bash
   migrate -database "sqlite://data/db/isos.db" -path migrations force VERSION
   ```

### Starting Fresh (Development Only)

To reset the database and re-run all migrations:

1. Stop the application
2. Delete the database file: `rm data/db/isos.db`
3. Restart the application (migrations will run automatically)

## Example Migrations

### Adding a Column

**up.sql:**
```sql
ALTER TABLE isos ADD COLUMN tags TEXT DEFAULT '';
```

**down.sql:**
```sql
ALTER TABLE isos DROP COLUMN tags;
```

### Creating an Index

**up.sql:**
```sql
CREATE INDEX idx_isos_status ON isos(status);
```

**down.sql:**
```sql
DROP INDEX idx_isos_status;
```

### Adding a Table

**up.sql:**
```sql
CREATE TABLE download_stats (
    id TEXT PRIMARY KEY,
    iso_id TEXT NOT NULL,
    download_count INTEGER DEFAULT 0,
    last_downloaded_at TIMESTAMP,
    FOREIGN KEY (iso_id) REFERENCES isos(id) ON DELETE CASCADE
);
```

**down.sql:**
```sql
DROP TABLE download_stats;
```

## Testing Migrations

Always test migrations before deploying:

1. Create a backup of your database
2. Apply the migration: `go run main.go`
3. Verify the changes
4. Test rollback: `migrate -database "sqlite://data/db/isos.db" -path migrations down 1`
5. Verify rollback worked
6. Re-apply migration: `migrate -database "sqlite://data/db/isos.db" -path migrations up 1`

## Production Deployment

In production, migrations run automatically on application startup. Ensure:

1. Backup database before deploying new version
2. Test migrations in staging environment first
3. Monitor application logs for migration errors
4. Have rollback plan ready

## References

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [SQLite ALTER TABLE](https://www.sqlite.org/lang_altertable.html)
- [SQLite Transactions](https://www.sqlite.org/lang_transaction.html)
