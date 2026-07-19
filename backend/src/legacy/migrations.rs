use std::collections::HashMap;
use std::fmt;
use sha2::{Sha256, Digest};

#[derive(Debug)]
pub enum MigrationError {
    ChecksumMismatch { expected: String, actual: String },
    MissingChecksum,
    InvalidMigrationContent,
}

impl fmt::Display for MigrationError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            MigrationError::ChecksumMismatch { expected, actual } => {
                write!(f, "Checksum mismatch: expected {}, but got {}", expected, actual)
            }
            MigrationError::MissingChecksum => write!(f, "Checksum is missing"),
            MigrationError::InvalidMigrationContent => write!(f, "Invalid migration content"),
        }
    }
}

impl std::error::Error for MigrationError {}

pub struct Migration {
    pub id: u32,
    pub sql: String,
    pub expected_checksum: String,
}

impl Migration {
    pub fn new(id: u32, sql: &str, expected_checksum: &str) -> Self {
        Migration {
            id,
            sql: sql.to_string(),
            expected_checksum: expected_checksum.to_string(),
        }
    }

    pub fn validate(&self) -> Result<(), MigrationError> {
        let actual_checksum = calculate_checksum(&self.sql);
        if actual_checksum!= self.expected_checksum {
            Err(MigrationError::ChecksumMismatch {
                expected: self.expected_checksum.clone(),
                actual: actual_checksum,
           })
        } else {
            Ok(())
        }
    }
}

fn calculate_checksum(sql: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(sql);
    let result = hasher.finalize();
    format!("{:x}", result)
}

pub struct MigrationRegistry {
    migrations: HashMap<u32, Migration>,
}

impl MigrationRegistry {
    pub fn new() -> Self {
        MigrationRegistry {
            migrations: HashMap::new(),
        }
    }

    pub fn register(&mut self, migration: Migration) -> Result<(), MigrationError> {
        if self.migrations.contains_key(&migration.id) {
            return Err(MigrationError::InvalidMigrationContent);
        }
        migration.validate()?;
        self.migrations.insert(migration.id, migration);
        Ok(())
    }

    pub fn get(&self, id: u32) -> Option<&Migration> {
        self.migrations.get(&id)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_migration() {
        let mut registry = MigrationRegistry::new();
        let valid_sql = "CREATE TABLE users (id INT PRIMARY KEY, name TEXT);";
        let valid_checksum = calculate_checksum(valid_sql);
        let migration = Migration::new(1, valid_sql, &valid_checksum);
        assert!(registry.register(migration).is_ok());
    }

    #[test]
    fn test_checksum_mismatch() {
        let mut registry = MigrationRegistry::new();
        let valid_sql = "CREATE TABLE users (id INT PRIMARY KEY, name TEXT);";
        let invalid_checksum = "invalid_checksum";
        let migration = Migration::new(1, valid_sql, invalid_checksum);
        assert!(matches!(
            registry.register(migration),
            Err(MigrationError::ChecksumMismatch {.. })
        ));
    }

    #[test]
    fn test_missing_checksum() {
        let mut registry = MigrationRegistry::new();
        let valid_sql = "CREATE TABLE users (id INT PRIMARY KEY, name TEXT);";
        let migration = Migration::new(1, valid_sql, "");
        assert!(matches!(
            registry.register(migration),
            Err(MigrationError::MissingChecksum)
        ));
    }

    #[test]
    fn test_invalid_migration_content() {
        let mut registry = MigrationRegistry::new();
        let valid_sql = "CREATE TABLE users (id INT PRIMARY KEY, name TEXT);";
        let valid_checksum = calculate_checksum(valid_sql);
        let migration = Migration::new(1, valid_sql, &valid_checksum);
        assert!(registry.register(migration).is_ok());

        let duplicate_migration = Migration::new(1, valid_sql, &valid_checksum);
        assert!(matches!(
            registry.register(duplicate_migration),
            Err(MigrationError::InvalidMigrationContent)
        ));
    }
}