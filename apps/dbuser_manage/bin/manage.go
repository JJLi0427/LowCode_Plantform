package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type UserPermissionManager struct {
	db *sql.DB
}

func NewUserPermissionManager(dsn string) (*UserPermissionManager, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &UserPermissionManager{db: db}, nil
}

func (upm *UserPermissionManager) AddUser(username, password string) error {
	query := fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s';", username, password)
	_, err := upm.db.Exec(query)
	if err != nil {
		return err
	}
	fmt.Printf("User %s added successfully.\n", username)
	return nil
}

func (upm *UserPermissionManager) DeleteUser(username string) error {
	query := fmt.Sprintf("DROP USER '%s'@'%%';", username)
	_, err := upm.db.Exec(query)
	if err != nil {
		return err
	}
	fmt.Printf("User %s deleted successfully.\n", username)
	return nil
}

func (upm *UserPermissionManager) GrantPermission(username, dbName, tableName, permission string) error {
	query := fmt.Sprintf("GRANT %s ON %s.%s TO '%s'@'%%';", permission, dbName, tableName, username)
	_, err := upm.db.Exec(query)
	if err != nil {
		return err
	}
	fmt.Printf("Granted %s permission on %s.%s to user %s.\n", permission, dbName, tableName, username)
	return nil
}

func (upm *UserPermissionManager) RevokePermission(username, dbName, tableName, permission string) error {
	query := fmt.Sprintf("REVOKE %s ON %s.%s FROM '%s'@'%%';", permission, dbName, tableName, username)
	_, err := upm.db.Exec(query)
	if err != nil {
		return err
	}
	fmt.Printf("Revoked %s permission on %s.%s from user %s.\n", permission, dbName, tableName, username)
	return nil
}

func main() {
	// Replace with your MySQL DSN
	dsn := "root:password@tcp(127.0.0.1:3306)/"
	manager, err := NewUserPermissionManager(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer manager.db.Close()

	// Example usage
	err = manager.AddUser("testuser", "testpassword")
	if err != nil {
		log.Printf("Error adding user: %v", err)
	}

	err = manager.GrantPermission("testuser", "testdb", "*", "SELECT, INSERT")
	if err != nil {
		log.Printf("Error granting permission: %v", err)
	}

	err = manager.RevokePermission("testuser", "testdb", "*", "INSERT")
	if err != nil {
		log.Printf("Error revoking permission: %v", err)
	}

	err = manager.DeleteUser("testuser")
	if err != nil {
		log.Printf("Error deleting user: %v", err)
	}
}
