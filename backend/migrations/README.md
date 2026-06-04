# Database Migrations
#
# This directory contains SQL migration files for the PostgreSQL database.
# Migrations are numbered and applied in order.
#
# Format:
#   001_initial_schema.up.sql
#   001_initial_schema.down.sql
#   002_add_users_table.up.sql
#   002_add_users_table.down.sql
#
# Convention:
# - Use migrate-cli (github.com/golang-migrate/migrate) for migrations
# - Each migration has an .up.sql (apply) and .down.sql (rollback) file
# - Files are numbered sequentially and immutable once applied
#
# Future: In Phase 1, we'll create migrations for:
#   - Users table
#   - AI Providers table
#   - Chat sessions table
#   - Rate limiting tables
#   - And more based on requirements
