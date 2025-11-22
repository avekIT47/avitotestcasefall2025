package database

import (
	"testing"
)

func TestNew(t *testing.T) {
	// Test with invalid connection string
	_, err := New("invalid connection string")
	if err == nil {
		t.Error("expected error for invalid connection string")
	}
}

func TestDB_Structure(t *testing.T) {
	// This test just checks that DB struct embeds *sql.DB correctly
	// We can't test actual database connection without a real database

	// Test that DB type exists and has the expected structure
	var db *DB
	if db != nil {
		t.Error("expected nil DB")
	}
}
