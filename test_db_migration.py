import unittest
from tools.db_migration import generate_dry_run_plan, print_dry_run_plan

class TestDBMigration(unittest.TestCase):
    def setUp(self):
        self.current_version = 1
        self.target_version = 5

    def test_no_pending_migrations(self):
        plan = generate_dry_run_plan(self.current_version, self.current_version)
        self.assertEqual(len(plan['migrations']), 0)

    def test_pending_migrations(self):
        plan = generate_dry_run_plan(self.current_version, self.target_version)
        self.assertEqual(len(plan['migrations']), self.target_version - self.current_version)
        for i, migration in enumerate(plan['migrations']):
            self.assertEqual(migration['version'], self.current_version + i + 1)
            self.assertEqual(migration['direction'], 'up')

    def test_specific_rollback_target(self):
        plan = generate_dry_run_plan(self.target_version, self.current_version, direction='down')
        self.assertEqual(len(plan['migrations']), self.target_version - self.current_version)
        for i, migration in enumerate(plan['migrations']):
            self.assertEqual(migration['version'], self.target_version - i)
            self.assertEqual(migration['direction'], 'down')

if __name__ == '__main__':
    unittest.main()