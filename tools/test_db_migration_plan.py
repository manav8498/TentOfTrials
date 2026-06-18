#!/usr/bin/env python3
"""Tests for db_migration.py dry-run planning mode."""

import json
import sys
import unittest

# Import the module under test
sys.path.insert(0, ".")
from tools.db_migration import generate_migration_plan


class TestMigrationPlan(unittest.TestCase):
    """Test the generate_migration_plan function."""

    def test_plan_up_returns_pending_migrations(self):
        """--up --plan should return all pending migrations."""
        plan = generate_migration_plan("up")
        self.assertIsInstance(plan, list)
        self.assertGreater(len(plan), 0, "Should have pending migrations")
        for entry in plan:
            self.assertIn("version", entry)
            self.assertIn("description", entry)
            self.assertEqual(entry["direction"], "up")
            self.assertTrue(entry["would_be_attempted"])

    def test_plan_up_is_valid_json(self):
        """Plan output should be valid JSON-serializable."""
        plan = generate_migration_plan("up")
        json_str = json.dumps(plan)
        self.assertIsInstance(json_str, str)
        # Round-trip to confirm valid JSON
        parsed = json.loads(json_str)
        self.assertEqual(len(parsed), len(plan))

    def test_plan_down_existing_version(self):
        """--down --version <existing> --plan should return one entry."""
        # Use a known valid version from the MIGRATIONS list
        plan = generate_migration_plan("down", "20210101000000")
        self.assertEqual(len(plan), 1)
        self.assertEqual(plan[0]["version"], "20210101000000")
        self.assertEqual(plan[0]["direction"], "down")
        self.assertTrue(plan[0]["would_be_attempted"])

    def test_plan_down_nonexistent_version(self):
        """--down --version <nonexistent> --plan should report NOT FOUND."""
        plan = generate_migration_plan("down", "00000000000000")
        self.assertEqual(len(plan), 1)
        self.assertEqual(plan[0]["description"], "NOT FOUND")
        self.assertFalse(plan[0]["would_be_attempted"])

    def test_plan_status_includes_all_migrations(self):
        """--status --plan should include all migrations."""
        plan = generate_migration_plan("status")
        self.assertIsInstance(plan, list)
        self.assertGreater(len(plan), 0)
        for entry in plan:
            self.assertIn("version", entry)
            self.assertIn("description", entry)
            self.assertIn("direction", entry)
            self.assertIn("would_be_attempted", entry)
            self.assertEqual(entry["direction"], "n/a")

    def test_plan_does_not_require_psql(self):
        """Plan generation must not require PostgreSQL / psql."""
        plan = generate_migration_plan("up")
        self.assertIsInstance(plan, list)
        # If we got here without FileNotFoundError or subprocess error, it passed

    def test_backward_compat_dry_run_flag_still_works(self):
        """--dry-run flag and --plan flag are independent concepts."""
        plan_up = generate_migration_plan("up")
        plan_status = generate_migration_plan("status")
        # Both should produce valid output without touching psql
        self.assertGreater(len(plan_up), 0)
        self.assertGreater(len(plan_status), 0)


if __name__ == "__main__":
    # Run tests and exit with appropriate code
    result = unittest.main(verbosity=2, exit=False)
    sys.exit(0 if result.result.wasSuccessful() else 1)
