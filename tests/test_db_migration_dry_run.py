import unittest

from tools import db_migration


class MigrationDryRunPlanTests(unittest.TestCase):
    def test_no_pending_migrations_has_empty_up_plan(self):
        status = [
            {"version": "20240101000000", "description": "Initial schema", "applied": True},
            {"version": "20240102000000", "description": "Add users", "applied": True},
        ]

        self.assertEqual(db_migration.build_migration_plan(status, "up"), [])

    def test_pending_migrations_include_machine_readable_fields(self):
        status = [
            {"version": "20240101000000", "description": "Initial schema", "applied": True},
            {"version": "20240102000000", "description": "Add users", "applied": False},
            {"version": "20240103000000", "description": "Add audit log", "applied": False},
        ]

        plan = db_migration.build_migration_plan(status, "up")

        self.assertEqual(
            plan,
            [
                {
                    "version": "20240102000000",
                    "description": "Add users",
                    "direction": "up",
                    "would_execute": True,
                },
                {
                    "version": "20240103000000",
                    "description": "Add audit log",
                    "direction": "up",
                    "would_execute": True,
                },
            ],
        )

    def test_specific_rollback_target_plans_back_to_target(self):
        status = [
            {"version": "20240101000000", "description": "Initial schema", "applied": True},
            {"version": "20240102000000", "description": "Add users", "applied": True},
            {"version": "20240103000000", "description": "Add audit log", "applied": True},
            {"version": "20240104000000", "description": "Add billing", "applied": False},
        ]

        plan = db_migration.build_migration_plan(
            status,
            "down",
            target_version="20240102000000",
        )

        self.assertEqual(
            [entry["version"] for entry in plan],
            ["20240103000000", "20240102000000"],
        )
        self.assertTrue(all(entry["direction"] == "down" for entry in plan))
        self.assertTrue(all(entry["would_execute"] for entry in plan))

    def test_dry_run_up_does_not_execute_migrations(self):
        original_status = db_migration.get_migration_status
        original_apply = db_migration.apply_migration
        try:
            db_migration.get_migration_status = lambda: [
                {
                    "version": "20240101000000",
                    "description": "Initial schema",
                    "applied": False,
                }
            ]

            def fail_if_called(version, direction="up"):
                raise AssertionError("dry-run must not execute migrations")

            db_migration.apply_migration = fail_if_called
            self.assertTrue(db_migration.run_all_migrations(dry_run=True))
        finally:
            db_migration.get_migration_status = original_status
            db_migration.apply_migration = original_apply


if __name__ == "__main__":
    unittest.main()
