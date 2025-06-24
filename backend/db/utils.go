package db

import (
	"database/sql"
	"fmt"
)

// Generic scanner interface for any type that can be scanned from database rows
type Scannable interface {
	Scan(rows *sql.Rows) error
}

// QueryList executes a query and scans results into a slice using the provided scanner function
func QueryList[T any](db *sql.DB, query string, scanner func(*sql.Rows) (*T, error), args ...interface{}) ([]*T, error) {
	// Always return an array, even if empty
	results := []*T{}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item, err := scanner(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// QuerySingle executes a query and scans a single result
func QuerySingle[T any](db *sql.DB, query string, scanner func(*sql.Row) (*T, error), args ...interface{}) (*T, error) {
	row := db.QueryRow(query, args...)
	return scanner(row)
}

// QueryFirstFromRows executes a query and returns the first result using the rows scanner
func QueryFirstFromRows[T any](db *sql.DB, query string, scanner func(*sql.Rows) (*T, error), args ...interface{}) (*T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	return scanner(rows)
}

// ValidateOwnership checks if the current user has permission to modify the resource
func ValidateOwnership(currentUserID, resourceUserID int, resourceType string) error {
	if currentUserID != resourceUserID {
		return fmt.Errorf("user does not have permission to update this %s", resourceType)
	}
	return nil
}

// ExecWithValidation executes a query with validation
func ExecWithValidation(db *sql.DB, query string, args ...interface{}) error {
	_, err := db.Exec(query, args...)
	return err
}

// CheckResourceExists checks if a resource exists and returns an error if used by other entities
func CheckResourceExists(db *sql.DB, checkQuery string, resourceType string, args ...interface{}) error {
	rows, err := db.Query(checkQuery, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return fmt.Errorf("%s is used in at least one transaction", resourceType)
	}

	return nil
}
