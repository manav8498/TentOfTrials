use sha2::{Sha256, Digest};
use std::fmt;

/// Errors that can occur during migration checksum validation.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ChecksumError {
    /// The migration content is empty or whitespace-only.
    EmptyContent,
    /// The expected checksum does not match the computed checksum.
    Mismatch {
        migration_id: u64,
        expected: String,
        actual: String,
    },
    /// The stored checksum is missing or invalid.
    MissingChecksum {
        migration_id: u64,
    },
}

impl fmt::Display for ChecksumError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ChecksumError::EmptyContent => write!(f, "migration content is empty or whitespace-only"),
            ChecksumError::Mismatch { migration_id, expected, actual } => {
                write!(f, "checksum mismatch for migration {}: expected {}, got {}", migration_id, expected, actual)
            }
            ChecksumError::MissingChecksum { migration_id } => {
                write!(f, "missing or invalid checksum for migration {}", migration_id)
            }
        }
    }
}

impl std::error::Error for ChecksumError {}

/// Compute a deterministic SHA-256 checksum for migration SQL content.
/// 
/// The content is trimmed and normalized (line endings converted to \n)
/// before hashing to ensure stable checksums across platforms.
pub fn compute_checksum(sql: &str) -> Result<String, ChecksumError> {
    let normalized = sql.trim();
    if normalized.is_empty() {
        return Err(ChecksumError::EmptyContent);
    }
    
    // Normalize line endings for cross-platform stability
    let normalized = normalized.replace("\r\n", "\n").replace("\r", "\n");
    
    let mut hasher = Sha256::new();
    hasher.update(normalized.as_bytes());
    let result = hasher.finalize();
    Ok(format!("{:x}", result))
}

/// Validate that a migration's SQL content matches its expected checksum.
/// 
/// Returns Ok(()) if the checksum matches, or a descriptive error otherwise.
pub fn validate_migration_checksum(migration_id: u64, sql: &str, expected_checksum: Option<&str>) -> Result<(), ChecksumError> {
    let expected = expected_checksum.ok_or(ChecksumError::MissingChecksum { migration_id })?;
    
    if expected.trim().is_empty() {
        return Err(ChecksumError::MissingChecksum { migration_id });
    }
    
    let actual = compute_checksum(sql)?;
    
    if actual != expected.trim() {
        return Err(ChecksumError::Mismatch {
            migration_id,
            expected: expected.to_string(),
            actual,
        });
    }
    
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_compute_checksum_valid_sql() {
        let sql = "CREATE TABLE users (id SERIAL PRIMARY KEY);";
        let checksum = compute_checksum(sql).unwrap();
        assert!(!checksum.is_empty());
        assert_eq!(checksum.len(), 64); // SHA-256 hex length
        
        // Deterministic: same input produces same output
        let checksum2 = compute_checksum(sql).unwrap();
        assert_eq!(checksum, checksum2);
    }

    #[test]
    fn test_compute_checksum_normalizes_whitespace() {
        let sql1 = "  CREATE TABLE users (id INT);  ";
        let sql2 = "CREATE TABLE users (id INT);";
        assert_eq!(compute_checksum(sql1).unwrap(), compute_checksum(sql2).unwrap());
    }

    #[test]
    fn test_compute_checksum_normalizes_line_endings() {
        let sql_unix = "SELECT 1;\nSELECT 2;";
        let sql_windows = "SELECT 1;\r\nSELECT 2;";
        assert_eq!(compute_checksum(sql_unix).unwrap(), compute_checksum(sql_windows).unwrap());
    }

    #[test]
    fn test_compute_checksum_empty_content() {
        assert_eq!(compute_checksum(""), Err(ChecksumError::EmptyContent));
        assert_eq!(compute_checksum("   "), Err(ChecksumError::EmptyContent));
        assert_eq!(compute_checksum("\n\t\r\n"), Err(ChecksumError::EmptyContent));
    }

    #[test]
    fn test_validate_checksum_match() {
        let sql = "ALTER TABLE users ADD COLUMN email VARCHAR(255);";
        let checksum = compute_checksum(sql).unwrap();
        assert!(validate_migration_checksum(20210101000000, sql, Some(&checksum)).is_ok());
    }

    #[test]
    fn test_validate_checksum_mismatch() {
        let sql = "ALTER TABLE users ADD COLUMN email VARCHAR(255);";
        let result = validate_migration_checksum(20210101000000, sql, Some("deadbeef"));
        assert!(matches!(result, Err(ChecksumError::Mismatch { .. })));
        
        if let Err(ChecksumError::Mismatch { migration_id, expected, actual }) = result {
            assert_eq!(migration_id, 20210101000000);
            assert_eq!(expected, "deadbeef");
            assert_ne!(actual, "deadbeef");
        }
    }

    #[test]
    fn test_validate_checksum_missing() {
        let sql = "SELECT 1;";
        assert_eq!(
            validate_migration_checksum(20210101000000, sql, None),
            Err(ChecksumError::MissingChecksum { migration_id: 20210101000000 })
        );
    }

    #[test]
    fn test_validate_checksum_empty_expected() {
        let sql = "SELECT 1;";
        assert_eq!(
            validate_migration_checksum(20210101000000, sql, Some("")),
            Err(ChecksumError::MissingChecksum { migration_id: 20210101000000 })
        );
        assert_eq!(
            validate_migration_checksum(20210101000000, sql, Some("   ")),
            Err(ChecksumError::MissingChecksum { migration_id: 20210101000000 })
        );
    }

    #[test]
    fn test_validate_checksum_empty_sql() {
        let result = validate_migration_checksum(20210101000000, "", Some("abc123"));
        assert_eq!(result, Err(ChecksumError::EmptyContent));
    }

    #[test]
    fn test_different_sql_different_checksums() {
        let sql1 = "CREATE TABLE a (id INT);";
        let sql2 = "CREATE TABLE b (id INT);";
        assert_ne!(compute_checksum(sql1).unwrap(), compute_checksum(sql2).unwrap());
    }
}
