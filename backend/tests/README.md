# Tests
#
# This directory contains integration and end-to-end tests.
#
# Structure:
#   - Unit tests: Located next to implementation files (e.g., config_test.go)
#   - Integration tests: Located in this directory (tests/)
#
# Running tests:
#   make test              # Run all tests
#   make test-coverage     # Run tests with coverage report
#
# Testing best practices (for Phase 0 and beyond):
#   1. Unit tests for business logic (in-package)
#   2. Integration tests for API endpoints (in tests/)
#   3. Table-driven tests for multiple scenarios
#   4. Mock external dependencies (database, cache, APIs)
#   5. Use testify/assert for assertions
#
# Future: In Phase 1+, we'll add:
#   - API endpoint tests
#   - Database integration tests
#   - Cache integration tests
#   - Authentication tests
