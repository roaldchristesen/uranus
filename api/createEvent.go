package api

import (
	"Uranus/app"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"net/http"
	"strings"
	"time"
)

type EventDate struct {
	Start              *time.Time `json:"start"`
	End                *time.Time `json:"end"`
	AccessibilityFlags []int      `json:"accessibility_flags"`
	VisitorInfoFlags   []int      `json:"visitor_info_flags"`
	SpaceId            *int       `json:"space_id"`
	EntryTime          *string    `json:"entry_time"`
}

type EventRequest struct {
	OrganizerId int         `json:"organizer_id"`
	SpaceId     *int        `json:"space_id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	ImageURL    *string     `json:"image_url,omitempty"`
	EventTypes  []int       `json:"event_types"`
	GenreTypes  []int       `json:"genre_types"`
	EventDates  []EventDate `json:"event_dates"`
}

func CreateEventHandler(c *gin.Context) {
	var req EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schema := pq.QuoteIdentifier(app.GApp.DbConfig.DBSchema)

	db := app.GApp.DbPool

	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not start transaction"})
		return
	}
	defer tx.Rollback(ctx)

	var eventId int
	query := fmt.Sprintf(`INSERT INTO %s.event (organizer_id, space_id, title, description) VALUES ($1, $2, $3, $4) RETURNING id`, schema)
	err = tx.QueryRow(ctx, query, req.OrganizerId, req.SpaceId, req.Title, req.Description).Scan(&eventId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert event"})
		return
	}

	//
	for _, eventDate := range req.EventDates {
		accessibilityFlags := app.CombineFlags(eventDate.AccessibilityFlags)
		visitorFlags := app.CombineFlags(eventDate.VisitorInfoFlags)

		query := ""
		args := []interface{}{eventId, eventDate.Start}
		argIndex := 3

		if eventDate.End != nil {
			query += "end, "
			args = append(args, eventDate.End)
			argIndex++
		}
		query += "accessibility_flags, visitor_info_flags"
		args = append(args, accessibilityFlags, visitorFlags)
		argIndex += 2

		if eventDate.SpaceId != nil {
			query += ", space_id"
			args = append(args, eventDate.SpaceId)
			argIndex++
		}
		if eventDate.EntryTime != nil {
			query += ", entry_time"
			args = append(args, eventDate.EntryTime)
			argIndex++
		}

		// Construct placeholder string
		placeholders := []string{"$1", "$2"}
		for i := 3; i <= len(args); i++ {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		}

		columns := "event_id, start, "
		columns += query

		fullQuery := fmt.Sprintf(`INSERT INTO %s.event_date (%s) VALUES (%s)`, schema, columns, strings.Join(placeholders, ", "))

		_, err := tx.Exec(ctx, fullQuery, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert event date"})
			return
		}
	}

	// TODO: !!!
	if req.ImageURL != nil {
		var imageId int
		err = tx.QueryRow(ctx,
			`INSERT INTO image (url) VALUES ($1) RETURNING id`,
			*req.ImageURL,
		).Scan(&imageId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert image"})
			return
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO event_link_images (event_id, image_id) VALUES ($1, $2)`,
			eventId, imageId,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link image"})
			return
		}
	}

	// Types
	for _, typeId := range req.EventTypes {
		query = fmt.Sprintf(`INSERT INTO %s.event_link_types (event_id, event_type_id) VALUES ($1, $2)`, schema)
		_, err := tx.Exec(ctx, query, eventId, typeId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert event types"})
			return
		}
	}

	// GenreTypes
	for _, genreId := range req.GenreTypes {
		query = fmt.Sprintf(`INSERT INTO %s.genre_link_types (event_id, genre_type_id) VALUES ($1, $2)`, schema)
		_, err := tx.Exec(ctx, query, eventId, genreId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert genres"})
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"event_id": eventId})
}
