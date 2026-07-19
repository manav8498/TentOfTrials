import os
import subprocess
import sys

def get_pending_migrations(current_version, target_version):
    migrations = []
    for version in range(current_version + 1, target_version + 1):
        migration_file = f"migrations/{version:04d}_migration.sql"
        if os.path.exists(migration_file):
            with open(migration_file, 'r') as f:
                description = f.readline().strip()
            migrations.append({
                'version': version,
                'description': description,
                'direction': 'up',
                'execute': True
            })
    return migrations

def get_rollback_migrations(current_version, target_version):
    migrations = []
    for version in range(current_version - 1, target_version - 1, -1):
        migration_file = f"migrations/{version:04d}_migration.sql"
        if os.path.exists(migration_file):
            with open(migration_file, 'r') as f:
                description = f.readline().strip()
            migrations.append({
               'version': version,
                'description': description,
                'direction': 'down',
                'execute': True
            })
    return migrations

def generate_dry_run_plan(current_version, target_version, direction='up'):
    if direction == 'up':
        migrations = get_pending_migrations(current_version, target_version)
    else:
        migrations = get_rollback_migrations(current_version, target_version)
    
    plan = {
        'current_version': current_version,
        'target_version': target_version,
        'direction': direction,
       'migrations': migrations
    }
    return plan

def print_dry_run_plan(plan):
    print(f"Current Version: {plan['current_version']}")
    print(f"Target Version: {plan['target_version']}")
    print(f"Direction: {plan['direction']}")
    print("Pending Migrations:")
    for migration in plan['migrations']:
        print(f"Version: {migration['version']}, Description: {migration['description']}, Direction: {migration['direction']}, Execute: {migration['execute']}")

def main():
    import argparse
    parser = argparse.ArgumentParser(description="Database migration tool")
    parser.add_argument('--status', action='store_true', help='Show current migration status')
    parser.add_argument('--up', type=int, help='Migrate up to the specified version')
    parser.add_argument('--down', type=int, help='Rollback to the specified version')
    parser.add_argument('--dry-run', action='store_true', help='Generate a dry-run migration plan')
    args = parser.parse_args()

    current_version = int(open('db_version.txt', 'r').read().strip())

    if args.status:
        print(f"Current Version: {current_version}")
    elif args.up:
        if args.dry_run:
            plan = generate_dry_run_plan(current_version, args.up)
            print_dry_run_plan(plan)
        else:
            # Existing migration logic
            pass
    elif args.down:
        if args.dry_run:
            plan = generate_dry_run_plan(current_version, args.down, direction='down')
            print_dry_run_plan(plan)
        else:
            # Existing rollback logic
            pass
    else:
        print("No action specified. Use --help for more information.")

if __name__ == "__main__":
    main()
