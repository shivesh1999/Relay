# Database Migrations
#
# This directory contains SQL migration files for the PostgreSQL database.
# Migrations are numbered and applied in order.
#
# Format:
#   000001_create_users_table.up.sql
#   000001_create_users_table.down.sql
#
# Convention:
# - Use migrate-cli (github.com/golang-migrate/migrate) for migrations
# - Each migration has an .up.sql (apply) and .down.sql (rollback) file
# - Files are numbered sequentially and immutable once applied
#
# Local workflow:
#   cp configs/.env.example configs/.env
#   update DATABASE_URL in configs/.env
#   make db-migrate-up
#   make db-migrate-status
#   make db-migrate-down
#
# Future: In Phase 1, we'll create migrations for:
#   - AI providers table
#   - Chat sessions table
#   - Rate limiting tables
#   - And more based on requirements
