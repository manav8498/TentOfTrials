import unittest
import json
import subprocess
import os

class TestDBMigrationPlan(unittest.TestCase):
    def test_dry_run_up(self):
        result = subprocess.run(["python3", "tools/db_migration.py", "--up", "--dry-run"], capture_output=True, text=True)
        self.assertEqual(result.returncode, 0)
        
        # Parse the JSON plan from the output
        try:
            # find where json starts
            lines = result.stdout.strip().split('\n')
            json_str = '\n'.join([l for l in lines if l.startswith('[') or l.startswith('{') or l.startswith('  ') or l.startswith(']')])
            plan = json.loads(json_str)
            self.assertTrue(isinstance(plan, list))
            self.assertGreater(len(plan), 0)
            self.assertEqual(plan[0]["direction"], "up")
            self.assertTrue(plan[0]["will_execute"])
        except Exception as e:
            self.fail(f"Failed to parse dry-run up plan: {e}\nOutput was: {result.stdout}")

    def test_dry_run_down(self):
        result = subprocess.run(["python3", "tools/db_migration.py", "--down", "--version", "20210101000000", "--dry-run"], capture_output=True, text=True)
        self.assertEqual(result.returncode, 0)
        
        try:
            plan = json.loads(result.stdout.strip())
            self.assertEqual(len(plan), 1)
            self.assertEqual(plan[0]["version"], "20210101000000")
            self.assertEqual(plan[0]["direction"], "down")
            self.assertTrue(plan[0]["will_execute"])
        except Exception as e:
            self.fail(f"Failed to parse dry-run down plan: {e}\nOutput was: {result.stdout}")

if __name__ == '__main__':
    unittest.main()
