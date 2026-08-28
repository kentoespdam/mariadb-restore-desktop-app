package scanner

import "strings"

// extractUseDB extracts the database name from "USE dbname;"
func extractUseDB(line string) string {
	s := strings.TrimSpace(line)
	upper := strings.ToUpper(s)
	if !strings.HasPrefix(upper, "USE ") {
		return ""
	}
	s = s[4:] // strip "USE " (4 chars)
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ";")
	s = strings.TrimSpace(s)
	return strings.Trim(s, "`\"'")
}

// parseDatabaseComment extracts database name from standard dump header comments:
// "-- Host: ... Database: dbname" or "-- Current Database: `dbname`" or "-- Database: dbname"
func parseDatabaseComment(line string) string {
	if !strings.HasPrefix(line, "--") {
		return ""
	}
	lower := strings.ToLower(line)
	if idx := strings.Index(lower, "database:"); idx != -1 {
		part := strings.TrimSpace(line[idx+len("database:"):])
		part = strings.Trim(part, "`;\"' \t\r\n")
		fields := strings.Fields(part)
		if len(fields) > 0 {
			return strings.Trim(fields[0], "`;\"'")
		}
	}
	return ""
}

// parseCreateTable detects "CREATE TABLE `dbname`.`tablename"` or "CREATE TABLE tablename"
func parseCreateTable(upper string) (objType, name string) {
	if !strings.HasPrefix(upper, "CREATE TABLE") {
		return "", ""
	}
	s := upper
	s = strings.ReplaceAll(s, "CREATE TABLE IF NOT EXISTS", "CREATE TABLE")
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "CREATE TABLE")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "`", "")
	s = strings.Trim(s, " \t")
	if idx := strings.Index(s, "("); idx > 0 {
		s = s[:idx]
	}
	s = strings.TrimSuffix(s, ";")
	if idx := strings.Index(s, "."); idx > 0 {
		s = s[idx+1:]
	}
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"'")
	if s == "" {
		return "", ""
	}
	return TypeTable, strings.ToLower(s)
}

// parseBlockMarker detects "-- Dumped routines/triggers/events for database `dbname`"
func parseBlockMarker(line string) (objType, dbName string) {
	lower := strings.ToLower(line)
	var prefix string
	switch {
	case strings.Contains(lower, "-- dumped routines for database"):
		objType = TypeRoutines
		prefix = "-- dumped routines for database"
	case strings.Contains(lower, "-- dumped triggers for database"):
		objType = TypeTriggers
		prefix = "-- dumped triggers for database"
	case strings.Contains(lower, "-- dumped events for database"):
		objType = TypeEvents
		prefix = "-- dumped events for database"
	default:
		return "", ""
	}
	suffix := line[len(prefix):]
	suffix = strings.TrimSpace(suffix)
	suffix = strings.TrimSuffix(suffix, ";")
	suffix = strings.TrimSpace(suffix)
	return objType, strings.Trim(suffix, "`\"'")
}
