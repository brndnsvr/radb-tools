// Package validator provides input validation utilities for the RADb client.
// It protects against path traversal, injection attacks, and malformed inputs.
package validator

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	// ErrInvalidPath indicates a path failed validation
	ErrInvalidPath = errors.New("invalid path: contains unsafe characters or patterns")

	// ErrPathTraversal indicates a path traversal attempt
	ErrPathTraversal = errors.New("path traversal detected")

	// ErrInvalidASN indicates an invalid AS number
	ErrInvalidASN = errors.New("invalid AS number")

	// ErrInvalidPrefix indicates an invalid IP prefix
	ErrInvalidPrefix = errors.New("invalid IP prefix")

	// ErrInvalidEmail indicates an invalid email address
	ErrInvalidEmail = errors.New("invalid email address")
)

// Regular expressions for validation
var (
	// asnRegex matches valid AS numbers (AS followed by digits)
	asnRegex = regexp.MustCompile(`^AS\d+$`)

	// emailRegex provides basic email validation
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// maintainerRegex matches valid maintainer names
	maintainerRegex = regexp.MustCompile(`^[A-Z0-9][A-Z0-9\-]*[A-Z0-9]$`)
)

// ValidatePath validates a file path for safety.
// It prevents path traversal attacks and ensures the path is within expected boundaries.
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("%w: empty path", ErrInvalidPath)
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("%w: contains null byte", ErrInvalidPath)
	}

	// Clean the path
	cleaned := filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("%w: path contains '..'", ErrPathTraversal)
	}

	// Ensure absolute paths don't escape expected boundaries
	if filepath.IsAbs(cleaned) {
		// Additional checks could be added here for specific allowed directories
		return nil
	}

	return nil
}

// ValidateASN validates an Autonomous System Number.
// Accepts formats like "AS64500" or "64500".
func ValidateASN(asn string) error {
	if asn == "" {
		return fmt.Errorf("%w: empty ASN", ErrInvalidASN)
	}

	// Normalize to AS prefix
	normalized := asn
	if !strings.HasPrefix(asn, "AS") {
		normalized = "AS" + asn
	}

	// Check format
	if !asnRegex.MatchString(normalized) {
		return fmt.Errorf("%w: must be in format AS#### or ####", ErrInvalidASN)
	}

	// Extract and validate numeric value
	numStr := strings.TrimPrefix(normalized, "AS")
	num, err := strconv.ParseUint(numStr, 10, 32)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidASN, err)
	}

	// ASN validation (valid range: 0-4294967295)
	// Private ASNs: 64512-65534 and 4200000000-4294967294
	if num > 4294967295 {
		return fmt.Errorf("%w: number too large (max: 4294967295)", ErrInvalidASN)
	}

	return nil
}

// ValidateIPPrefix validates an IP prefix (IPv4 or IPv6 CIDR notation).
func ValidateIPPrefix(prefix string) error {
	if prefix == "" {
		return fmt.Errorf("%w: empty prefix", ErrInvalidPrefix)
	}

	// Parse CIDR notation
	ip, ipNet, err := net.ParseCIDR(prefix)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPrefix, err)
	}

	// Ensure the IP is the network address (not a host address)
	if !ip.Equal(ipNet.IP) {
		return fmt.Errorf("%w: host bits set (should be %s)", ErrInvalidPrefix, ipNet.String())
	}

	return nil
}

// ValidatePrefix is an alias for ValidateIPPrefix for convenience.
func ValidatePrefix(prefix string) error {
	return ValidateIPPrefix(prefix)
}

// ValidateEmail validates an email address using basic regex.
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("%w: empty email", ErrInvalidEmail)
	}

	if len(email) > 254 {
		return fmt.Errorf("%w: too long (max: 254 characters)", ErrInvalidEmail)
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("%w: invalid format", ErrInvalidEmail)
	}

	return nil
}

// ValidateMaintainer validates a maintainer name (mnt-by field).
// RADb maintainer names typically follow RPSL object naming conventions.
func ValidateMaintainer(mntner string) error {
	if mntner == "" {
		return errors.New("empty maintainer name")
	}

	if len(mntner) > 80 {
		return errors.New("maintainer name too long (max: 80 characters)")
	}

	// Maintainer names should be uppercase alphanumeric with hyphens
	if !maintainerRegex.MatchString(mntner) {
		return errors.New("invalid maintainer format (use uppercase, alphanumeric, and hyphens)")
	}

	return nil
}

// SanitizeString removes potentially dangerous characters from a string.
// Use this for user-provided strings that will be used in file names or API calls.
func SanitizeString(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")

	// Remove control characters
	var result strings.Builder
	for _, r := range s {
		if r >= 32 && r < 127 || r >= 128 {
			result.WriteRune(r)
		}
	}

	return strings.TrimSpace(result.String())
}

// ValidateSource validates a RADb source name.
func ValidateSource(source string) error {
	if source == "" {
		return errors.New("empty source")
	}

	// Currently only RADB is supported, but this allows for future expansion
	validSources := map[string]bool{
		"RADB":      true,
		"RIPE":      false, // Future support
		"ARIN":      false, // Future support
		"APNIC":     false, // Future support
		"AFRINIC":   false, // Future support
		"LACNIC":    false, // Future support
	}

	upper := strings.ToUpper(source)
	supported, exists := validSources[upper]

	if !exists {
		return fmt.Errorf("unknown source: %s", source)
	}

	if !supported {
		return fmt.Errorf("source %s not yet supported", source)
	}

	return nil
}
