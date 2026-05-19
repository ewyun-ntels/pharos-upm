package common

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// 화이트리스트 기반 입력 검증을 위한 패턴들
var (
	UUIDPattern     = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	SimpleIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,50}$`)

	AlertNamePattern    = regexp.MustCompile(`^[a-zA-Z0-9_\-.\s]{1,100}$`)
	ResourceNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_\-]{1,50}$`)
	FileNamePattern     = regexp.MustCompile(`^[a-zA-Z0-9_\-.]{1,255}$`)

	HostnamePattern  = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-.]{0,253}[a-zA-Z0-9]$`)
	IPAddressPattern = regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)

	SQLSafeStringPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-.\s]{0,255}$`)

	DateTimeFormats = []string{
		time.RFC3339,
		time.DateTime,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
)

// ValidationError 검증 오류를 나타내는 구조체
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", v.Field, v.Message)
}

// Validator 입력 검증을 위한 인터페이스
type Validator struct {
	errors []ValidationError
}

// NewValidator 새로운 검증기를 생성합니다
func NewValidator() *Validator {
	return &Validator{errors: make([]ValidationError, 0)}
}

// AddError 검증 오류를 추가합니다
func (v *Validator) AddError(field, value, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	})
}

// HasErrors 오류가 있는지 확인합니다
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// GetErrors 모든 검증 오류를 반환합니다
func (v *Validator) GetErrors() []ValidationError {
	return v.errors
}

// GetFirstError 첫 번째 오류 메시지를 반환합니다
func (v *Validator) GetFirstError() string {
	if len(v.errors) > 0 {
		return v.errors[0].Error()
	}
	return ""
}

// ValidateUUID UUID 형식을 검증합니다
func (v *Validator) ValidateUUID(field, value string, required bool) string {
	if value == "" {
		if required {
			v.AddError(field, value, "UUID is required")
			return ""
		}
		return value
	}

	if !UUIDPattern.MatchString(value) {
		// UUID 파싱으로 2차 검증
		if _, err := uuid.Parse(value); err != nil {
			v.AddError(field, value, "invalid UUID format")
			return ""
		}
	}
	return value
}

// ValidateSimpleID 단순 ID를 검증합니다
func (v *Validator) ValidateSimpleID(field, value string, required bool) string {
	if value == "" {
		if required {
			v.AddError(field, value, "ID is required")
			return ""
		}
		return value
	}

	if !SimpleIDPattern.MatchString(value) {
		v.AddError(field, value, "ID must contain only alphanumeric characters, underscores, and hyphens (max 50 chars)")
		return ""
	}
	return value
}

// ValidateAlertName alert name을 검증합니다
func (v *Validator) ValidateAlertName(field, value string, required bool) string {
	if value == "" {
		if required {
			v.AddError(field, value, "alert name is required")
			return ""
		}
		return value
	}

	// 길이 체크
	if len(value) > 100 {
		v.AddError(field, value, "alert name cannot exceed 100 characters")
		return ""
	}

	// 패턴 체크
	if !AlertNamePattern.MatchString(value) {
		v.AddError(field, value, "alert name contains invalid characters")
		return ""
	}

	// 연속된 공백 제거
	sanitized := regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(value), " ")
	return sanitized
}

// ValidateResourceName 리소스 이름을 검증합니다
func (v *Validator) ValidateResourceName(field, value string, required bool) string {
	if value == "" {
		if required {
			v.AddError(field, value, "resource name is required")
			return ""
		}
		return value
	}

	if !ResourceNamePattern.MatchString(value) {
		v.AddError(field, value, "resource name must contain only alphanumeric characters, underscores, and hyphens (max 50 chars)")
		return ""
	}
	return value
}

// ValidateFileName 파일명을 검증합니다
func (v *Validator) ValidateFileName(field, value string, required bool) string {
	if value == "" {
		if required {
			v.AddError(field, value, "file name is required")
			return ""
		}
		return value
	}

	if !FileNamePattern.MatchString(value) {
		v.AddError(field, value, "file name contains invalid characters")
		return ""
	}

	// 위험한 파일명 패턴 체크
	dangerous := []string{"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, pattern := range dangerous {
		if strings.Contains(value, pattern) {
			v.AddError(field, value, "file name contains dangerous characters")
			return ""
		}
	}

	return value
}

// ValidateHostname 호스트명을 검증합니다
func (v *Validator) ValidateHostname(field, value string, required bool) string {
	if value == "" {
		if required {
			v.AddError(field, value, "hostname is required")
			return ""
		}
		return value
	}

	// IP 주소인지 확인
	if IPAddressPattern.MatchString(value) {
		return value // IP 주소는 유효함
	}

	// 호스트명 패턴 확인
	if !HostnamePattern.MatchString(value) {
		v.AddError(field, value, "invalid hostname format")
		return ""
	}

	// 길이 제한
	if len(value) > 253 {
		v.AddError(field, value, "hostname too long (max 253 characters)")
		return ""
	}

	return strings.ToLower(value)
}

// ValidateDateTime 날짜/시간을 검증합니다
func (v *Validator) ValidateDateTime(field, value string, required bool) time.Time {
	var zero time.Time

	if value == "" {
		if required {
			v.AddError(field, value, "date/time is required")
		}
		return zero
	}

	// 여러 포맷으로 파싱 시도
	for _, format := range DateTimeFormats {
		if t, err := time.Parse(format, value); err == nil {
			return t
		}
	}

	v.AddError(field, value, "invalid date/time format")
	return zero
}

// ValidateCount 카운트 값을 검증합니다
func (v *Validator) ValidateCount(field, value string, min, max int, required bool) int {
	if value == "" {
		if required {
			v.AddError(field, value, "count is required")
		}
		return 0
	}

	count, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(field, value, "count must be a valid integer")
		return 0
	}

	if count < min {
		v.AddError(field, value, fmt.Sprintf("count must be at least %d", min))
		return 0
	}

	if count > max {
		v.AddError(field, value, fmt.Sprintf("count cannot exceed %d", max))
		return 0
	}

	return count
}

// ValidateEnum 열거값을 검증합니다
func (v *Validator) ValidateEnum(field, value string, allowedValues []string, required bool) string {
	if value == "" {
		if required {
			v.AddError(field, value, "value is required")
			return ""
		}
		return value
	}

	if slices.Contains(allowedValues, value) {
		return value
	}

	v.AddError(field, value, fmt.Sprintf("value must be one of: %s", strings.Join(allowedValues, ", ")))
	return ""
}

// ValidateJSONString JSON 문자열을 검증합니다
func (v *Validator) ValidateJSONString(field, value string, required bool) string {
	if value == "" {
		if required {
			v.AddError(field, value, "JSON is required")
			return ""
		}
		return value
	}

	// JSON 유효성 검사
	var temp any
	if err := json.Unmarshal([]byte(value), &temp); err != nil {
		v.AddError(field, value, "invalid JSON format")
		return ""
	}

	// JSON 문자열 크기 제한 (1MB)
	if len(value) > 1024*1024 {
		v.AddError(field, value, "JSON too large (max 1MB)")
		return ""
	}

	return value
}

// ValidateStringLength 문자열 길이를 검증합니다
func (v *Validator) ValidateStringLength(field, value string, minLen, maxLen int, required bool) string {
	if value == "" {
		if required {
			v.AddError(field, value, "value is required")
			return ""
		}
		return value
	}

	if len(value) < minLen {
		v.AddError(field, value, fmt.Sprintf("must be at least %d characters", minLen))
		return ""
	}

	if len(value) > maxLen {
		v.AddError(field, value, fmt.Sprintf("cannot exceed %d characters", maxLen))
		return ""
	}

	// 제어 문자 제거
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1 // 제거
		}
		return r
	}, value)

	return cleaned
}

// ValidateQueryParam 쿼리 매개변수를 안전하게 검증합니다
func (v *Validator) ValidateQueryParam(field, value string, required bool) string {
	if value == "" {
		if required {
			v.AddError(field, value, "query parameter is required")
			return ""
		}
		return value
	}

	// URL 디코딩
	decoded, err := url.QueryUnescape(value)
	if err != nil {
		v.AddError(field, value, "invalid URL encoding")
		return ""
	}

	// SQL 안전한 패턴 체크
	if !SQLSafeStringPattern.MatchString(decoded) {
		v.AddError(field, value, "query parameter contains invalid characters")
		return ""
	}

	return decoded
}
