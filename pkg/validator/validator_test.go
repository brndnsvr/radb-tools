package validator

import (
	"testing"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid absolute path", "/home/user/config.yaml", false},
		{"valid relative path", "config.yaml", false},
		{"empty path", "", true},
		{"path traversal", "/home/user/../../etc/passwd", true},
		{"null byte", "/home/user\x00/config", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidateASN(t *testing.T) {
	tests := []struct {
		name    string
		asn     string
		wantErr bool
	}{
		{"valid with AS prefix", "AS64500", false},
		{"valid without prefix", "64500", false},
		{"valid large ASN", "AS4294967295", false},
		{"empty", "", true},
		{"too large", "AS4294967296", true},
		{"invalid format", "AS64500X", true},
		{"negative", "AS-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateASN(tt.asn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateASN(%q) error = %v, wantErr %v", tt.asn, err, tt.wantErr)
			}
		})
	}
}

func TestValidateIPPrefix(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		wantErr bool
	}{
		{"valid IPv4", "192.0.2.0/24", false},
		{"valid IPv6", "2001:db8::/32", false},
		{"empty", "", true},
		{"host bits set", "192.0.2.1/24", true},
		{"invalid CIDR", "192.0.2.0", true},
		{"invalid IP", "999.999.999.999/24", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPPrefix(tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPPrefix(%q) error = %v, wantErr %v", tt.prefix, err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"valid with subdomain", "user@mail.example.com", false},
		{"empty", "", true},
		{"no at sign", "userexample.com", true},
		{"no domain", "user@", true},
		{"no local part", "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestValidateMaintainer(t *testing.T) {
	tests := []struct {
		name    string
		mntner  string
		wantErr bool
	}{
		{"valid", "MAINT-AS64500", false},
		{"valid simple", "MAINT", false},
		{"empty", "", true},
		{"lowercase", "maint-as64500", true},
		{"starts with hyphen", "-MAINT", true},
		{"ends with hyphen", "MAINT-", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaintainer(tt.mntner)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMaintainer(%q) error = %v, wantErr %v", tt.mntner, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSource(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{"valid RADB", "RADB", false},
		{"valid lowercase", "radb", false},
		{"unsupported RIPE", "RIPE", true},
		{"unknown", "UNKNOWN", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSource(tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSource(%q) error = %v, wantErr %v", tt.source, err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal string", "hello world", "hello world"},
		{"with null byte", "hello\x00world", "helloworld"},
		{"with control chars", "hello\x01\x02world", "helloworld"},
		{"with whitespace", "  hello world  ", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
