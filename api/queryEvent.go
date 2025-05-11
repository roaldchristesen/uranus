package api

import (
	"Uranus/app"
	"Uranus/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"strconv"
	"strings"
	"time"
)

func QueryEvent(c *gin.Context) {

	jsonData, httpStatus, err := queryAsJSON(c, app.GApp.DbPool)
	if err != nil {
		c.JSON(httpStatus, gin.H{"error": err.Error()})
		return
	}

	c.Data(httpStatus, "application/json", jsonData)
}

func queryAsJSON(c *gin.Context, db *pgxpool.Pool) ([]byte, int, error) {
	// TODO:
	// Note on security:
	// This version is still vulnerable to SQL injection if any of the inputs are user-controlled. Safe version using parameterized queries (recommended with database/sql or GORM):

	// TODO:
	// Check for unknown arguments

	start := time.Now() // Start timer
	ctx := c.Request.Context()

	query := app.GApp.SqlQueryEvent

	// Default condition: WHERE ed.start > NOW()
	// Default order: ORDER BY ed.start ASC

	languageStr := c.Query("lang")
	startStr := c.Query("start")
	endStr := c.Query("end")
	timeStr := c.Query("time")
	searchStr := c.Query("search")
	eventIdsStr := c.Query("events")
	venueIdsStr := c.Query("venues")
	spaceIdsStr := c.Query("spaces")
	orgIdsStr := c.Query("organizers")
	countryCodesStr := c.Query("countries")
	// stateCode := c.Query("state_code")
	postalCodeStr := c.Query("postal_code")
	// buildingLevelCodeStr := c.Query("building_level")
	// buildingMinLevelCodeStr := c.Query("building_min_level")
	// buildingMaxLevelCodeStr := c.Query("building_max_level")
	// spaceMinCapacityCodeStr := c.Query("space_min_capacity")
	// spaceMaxCapacityCodeStr := c.Query("space_max_capacity")
	// spaceMinSeatsCodeStr := c.Query("space_min_seats")
	// spaceMaxSeatsCodeStr := c.Query("space_max_seats")
	lonStr := c.Query("lon")
	latStr := c.Query("lat")
	radiusStr := c.Query("radius")
	eventTypesStr := c.Query("event_types")
	genreTypesStr := c.Query("genre_types")
	spaceTypesStr := c.Query("space_types")
	titleStr := c.Query("title")
	cityStr := c.Query("city")
	accessibilityInfosStr := c.Query("accessibility")
	visitorInfosStr := c.Query("visitor_infos")
	ageStr := c.Query("age")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	// TODO: offset, limit

	eventDateConditions := ""
	var conditions []string
	var args []interface{}
	argIndex := 1 // Postgres uses $1, $2, etc.
	var err error

	if languageStr != "" {
		if !app.IsValidIso639_1(languageStr) {
			return nil, 500, fmt.Errorf("lang format error, %s", languageStr)
		}
	} else {
		languageStr = "de"
	}

	args = append(args, languageStr)
	argIndex++

	if app.IsValidDateStr(startStr) {
		eventDateConditions += "WHERE ed.start >= $" + strconv.Itoa(argIndex)
		args = append(args, startStr)
		argIndex++
	} else if startStr != "" {
		return nil, 500, fmt.Errorf("start %s has invalid format", startStr)
	} else {
		eventDateConditions += "WHERE ed.start >= CURRENT_DATE"
	}

	if app.IsValidDateStr(endStr) {
		eventDateConditions += " AND (ed.end <= $" + strconv.Itoa(argIndex) + " OR ed.start <= $" + strconv.Itoa(argIndex) + ")"
		args = append(args, endStr)
		argIndex++
	} else if endStr != "" {
		return nil, 500, fmt.Errorf("end %s has invalid format", endStr)
	}

	argIndex, err = sql.BuildTimeCondition(timeStr, "start", "time", argIndex, &conditions, &args)
	if err != nil {
		return nil, 500, err
	}

	argIndex, err = sql.BuildSanitizedIlikeCondition(searchStr, "e.description", "search", argIndex, &conditions, &args)
	if err != nil {
		return nil, 500, err
	}

	if countryCodesStr != "" {
		format := "v.country_code IN (%s)"
		argIndex, err = sql.BuildInConditionForStringSlice(countryCodesStr, format, "country_codes", argIndex, &conditions, &args)
		if err != nil {
			return nil, 500, err
		}
	}

	if postalCodeStr != "" {
		argIndex, err = sql.BuildLikeConditions(postalCodeStr, "v.postal_code", argIndex, &conditions, &args)
		if err != nil {
			return nil, 500, err
		}
	}

	if eventIdsStr != "" {
		argIndex, err = sql.BuildColumnInIntCondition(eventIdsStr, "e.id", "events", argIndex, &conditions, &args)
		if err != nil {
			return nil, 500, err
		}
	}

	if venueIdsStr != "" {
		argIndex, err = sql.BuildColumnInIntCondition(venueIdsStr, "v.id", "venues", argIndex, &conditions, &args)
		if err != nil {
			return nil, 500, err
		}
	}

	if orgIdsStr != "" {
		argIndex, err = sql.BuildColumnInIntCondition(orgIdsStr, "o.id", "organizers", argIndex, &conditions, &args)
		if err != nil {
			return nil, 500, err
		}
	}

	if spaceIdsStr != "" {
		argIndex, err = sql.BuildColumnInIntCondition(spaceIdsStr, "COALESCE(s.id, es.id)", "spaces", argIndex, &conditions, &args)
		if err != nil {
			return nil, 500, err
		}
	}

	argIndex, err = sql.BuildGeographicRadiusCondition(
		lonStr, latStr, radiusStr, "v.wkb_geometry",
		argIndex, &conditions, &args,
	)
	if err != nil {
		return nil, 500, err
	}

	if eventTypesStr != "" {
		format := "EXISTS (SELECT 1 FROM uranus.event_link_types sub_elt WHERE sub_elt.event_id = e.id AND sub_elt.event_type_id IN (%s))"
		var err error
		argIndex, err = sql.BuildInCondition(eventTypesStr, format, "event_types", argIndex, &conditions, &args)
		if err != nil {
			return nil, 500, err
		}
	}

	if genreTypesStr != "" {
		format := "EXISTS (SELECT 1 FROM uranus.genre_link_types sub_glt WHERE sub_glt.event_id = e.id AND sub_glt.genre_type_id IN (%s))"
		var err error
		argIndex, err = sql.BuildInCondition(genreTypesStr, format, "genre_types", argIndex, &conditions, &args)
		if err != nil {
			return nil, 500, err
		}
	}

	if spaceTypesStr != "" {
		format := "COALESCE(s.space_type_id, es.space_type_id) IN (%s)"
		argIndex, err = sql.BuildInCondition(spaceTypesStr, format, "space_types", argIndex, &conditions, &args)
		if err != nil {
			return nil, 500, err
		}
	}

	argIndex, err = sql.BuildSanitizedIlikeCondition(titleStr, "e.title", "title", argIndex, &conditions, &args)
	if err != nil {
		return nil, 500, err
	}

	argIndex, err = sql.BuildSanitizedIlikeCondition(cityStr, "v.city", "city", argIndex, &conditions, &args)
	if err != nil {
		return nil, 500, err
	}

	argIndex, err = sql.BuildContainedInColumnRangeCondition(ageStr, "min_age", "max_age", argIndex, &conditions, &args)
	if err != nil {
		return nil, 500, err
	}

	argIndex, err = sql.BuildBitmaskCondition(accessibilityInfosStr, "ed.accessibility_flags", "accessibility_flags", argIndex, &conditions, &args)
	if err != nil {
		fmt.Println(".... err", err)
		return nil, 500, err
	}

	argIndex, err = sql.BuildBitmaskCondition(visitorInfosStr, "ed.visitor_info_flags", "visitor_info_flags", argIndex, &conditions, &args)
	if err != nil {
		return nil, 500, err
	}

	conditionsStr := ""
	if len(conditions) > 0 {
		conditionsStr = "WHERE " + strings.Join(conditions, " AND ")
		fmt.Println(conditionsStr)
	}

	order := "ORDER BY ed.start ASC"

	// Add LIMIT and OFFSET
	limitClause, argIndex, err := sql.BuildLimitOffsetClause(limitStr, offsetStr, argIndex, &args)
	if err != nil {
		return nil, 400, err
	}

	query = strings.Replace(query, "{{event-date-conditions}}", eventDateConditions, 1)
	query = strings.Replace(query, "{{conditions}}", conditionsStr, 1)
	query = strings.Replace(query, "{{limit}}", limitClause, 1)
	query = strings.Replace(query, "{{order}}", order, 1)

	fmt.Println(query)
	fmt.Printf("eventDateConditions: %#v\n", eventDateConditions)
	fmt.Printf("conditions: %#v\n", conditions)
	fmt.Printf("args: %d: %#v\n", len(args), args)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, 500, err
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	columnNames := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columnNames[i] = string(fd.Name)
	}

	var results []map[string]interface{}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, 500, err
		}

		rowMap := make(map[string]interface{}, len(values))
		for i, col := range columnNames {
			rowMap[col] = values[i]
		}

		results = append(results, rowMap)
	}

	if rows.Err() != nil {
		return nil, 500, rows.Err()
	}

	type QueryResponse struct {
		Total   int                      `json:"total"`
		Time    string                   `json:"time"`
		Columns []string                 `json:"columns"`
		Results []map[string]interface{} `json:"events"`
	}

	elapsed := time.Since(start)
	milliseconds := int(elapsed.Milliseconds())

	response := QueryResponse{
		Total:   len(results),
		Columns: columnNames,
		Time:    fmt.Sprintf("%d msec", milliseconds),
		Results: results,
	}

	if response.Total < 1 {
		return nil, 200, fmt.Errorf("query returned 0 results")
	} else {
		jsonData, err := json.MarshalIndent(response, "", "  ")
		return jsonData, 200, err
	}
}
