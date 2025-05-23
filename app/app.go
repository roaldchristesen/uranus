package app

import (
	"Uranus/config"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5" // PostgreSQL driver
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type App struct {
	DbPool        *pgxpool.Pool
	DbConfig      config.DBConfig
	SqlQueryEvent string
}

var GApp App

func (app *App) LoadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &app.DbConfig)
	if err != nil {
		return err
	}

	app.DbConfig.Print()
	return nil
}

func (app *App) PrepareSql() error {

	// Read the SQL file
	content, err := os.ReadFile("queries/queryEvent.sql")
	if err != nil {
		panic(fmt.Errorf("failed to read file: %w", err))
	}

	// Convert to string and replace {{schema}} with actual schema name
	query := string(content)
	query = strings.ReplaceAll(query, "{{schema}}", app.DbConfig.DBSchema)
	app.SqlQueryEvent = query
	fmt.Println(app.SqlQueryEvent)

	return nil
}

func (app *App) InitDB() error {

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", app.DbConfig.User, app.DbConfig.Password, app.DbConfig.Host, app.DbConfig.Port, app.DbConfig.DBName)

	var err error
	app.DbPool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
		return err
	}

	fmt.Println("Database connection pool initialized!")
	return nil
}

// EncryptPassword hashes a password and returns the hashed string along with any error
func EncryptPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12) // bcrypt.DefaultCost
	if err != nil {
		return "", err // Return an empty string and the error
	}
	return string(hashedPassword), nil // Return the hashed password and nil error
}

// ComparePasswords compares a plain password with a bcrypt hash
func ComparePasswords(storedHash, password string) error {
	// Compare the plain password to the bcrypt hash
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
}

func ReadSVG(path string) string {
	svgContent, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(svgContent)
}

// TruncateAtWord truncates the string at the word boundary
func TruncateAtWord(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	words := strings.Fields(s)
	var truncated string
	for _, word := range words {
		// Add the word and a space if it doesn't exceed the max length
		if len(truncated)+len(word)+1 <= maxLength {
			if truncated == "" {
				truncated = word
			} else {
				truncated += " " + word
			}
		} else {
			break
		}
	}
	if len(truncated) < len(s) {
		truncated += " ..."
	}
	return truncated
}

// Function to convert database errors to HTTP status codes
func (app *App) DbErrorToHTTP(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Check for "no rows" error (record not found)
	if errors.Is(err, pgx.ErrNoRows) {
		return http.StatusNotFound
	}

	// Check for PostgreSQL-specific errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique constraint violation
			return http.StatusConflict
		case "23503": // Foreign key violation
			return http.StatusBadRequest
		case "42P01": // Undefined table
			return http.StatusInternalServerError
		default:
			return http.StatusInternalServerError
		}
	}

	// Default to 500 Internal Server Error
	return http.StatusInternalServerError
}

// Utility function to extract string from sql.NullString
func SqlNullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL" // Return empty string if NULL
}

// Utility function to extract time from sql.NullTime
func SqlNullTimeToString(nt sql.NullTime) string {
	if nt.Valid {
		return nt.Time.Format(time.RFC3339)
	}
	return "NULL" // Return empty string if NULL
}

// Utility function to extract int64 from sql.NullInt64
func SqlNullInt64ToInt(n sql.NullInt64) int64 {
	if n.Valid {
		return n.Int64
	}
	return 0 // Return 0 if NULL (you can choose another default value)
}

// Utility function to extract string from sql.NullInt64
func SqlNullInt64ToString(n sql.NullInt64) string {
	if n.Valid {
		return fmt.Sprintf("%d", n.Int64)
	}
	return "NULL" // Return "NULL" if NULL
}

// Utility function to extract bool from sql.NullBool
func SqlNullBoolToBool(nb sql.NullBool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return false // Return false if NULL (you can choose another default value)
}

// Utility function to extract string from sql.NullBool
func SqlNullBoolToString(nb sql.NullBool) string {
	if nb.Valid {
		return fmt.Sprintf("%t", nb.Bool)
	}
	return "NULL" // Return "NULL" if NULL
}

func IsValidDateStr(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

func IsValidIso639_1(languageStr string) bool {
	if languageStr != "" {
		match, _ := regexp.MatchString("^[a-z]{2}$", languageStr)
		return match
	}
	return false
}
