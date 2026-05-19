package orm

import (
	"database/sql/driver"
	"testing"
	"time"
)

func TestDatetime_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    time.Time
		wantErr bool
	}{
		{
			name:    "time.Time value",
			value:   time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC),
			want:    time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "string value with valid datetime",
			value:   "2023-10-15 14:30:45",
			want:    time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "byte slice value with valid datetime",
			value:   []byte("2023-10-15 14:30:45"),
			want:    time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "string value with invalid datetime format",
			value:   "invalid-datetime",
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "byte slice value with invalid datetime format",
			value:   []byte("invalid-datetime"),
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "unsupported type int",
			value:   123,
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "unsupported type float64",
			value:   123.45,
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "nil value",
			value:   nil,
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := &Datetime{}
			err := dt.Scan(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Datetime.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !dt.Time.Equal(tt.want) {
				t.Errorf("Datetime.Scan() = %v, want %v", dt.Time, tt.want)
			}
		})
	}
}

func TestDatetime_Value(t *testing.T) {
	tests := []struct {
		name    string
		dt      Datetime
		want    driver.Value
		wantErr bool
	}{
		{
			name: "valid datetime",
			dt: Datetime{
				Time: time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC),
			},
			want:    "2023-10-15 14:30:45",
			wantErr: false,
		},
		{
			name: "zero time",
			dt: Datetime{
				Time: time.Time{},
			},
			want:    "0001-01-01 00:00:00",
			wantErr: false,
		},
		{
			name: "time with timezone - should convert to UTC",
			dt: Datetime{
				Time: time.Date(2023, 10, 15, 14, 30, 45, 0, time.FixedZone("KST", 9*60*60)),
			},
			want:    "2023-10-15 05:30:45", // UTC equivalent
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.dt.Value()

			if (err != nil) != tt.wantErr {
				t.Errorf("Datetime.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Datetime.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatetime_ScanValue_RoundTrip(t *testing.T) {
	// Test round-trip conversion: Value() -> Scan()
	original := time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC)
	dt1 := Datetime{Time: original}

	// Convert to driver value
	value, err := dt1.Value()
	if err != nil {
		t.Fatalf("Value() failed: %v", err)
	}

	// Scan back from driver value
	dt2 := &Datetime{}
	err = dt2.Scan(value)
	if err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}

	// Compare times (should be equal)
	if !dt1.Time.Equal(dt2.Time) {
		t.Errorf("Round-trip failed: original %v, result %v", dt1.Time, dt2.Time)
	}
}

func TestDatetime_ScanEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
		},
		{
			name:    "empty byte slice",
			value:   []byte(""),
			wantErr: true,
		},
		{
			name:    "valid datetime with different format",
			value:   "2023/10/15 14:30:45", // Different separator
			wantErr: true,                  // Should fail because it's not the expected format
		},
		{
			name:    "partial datetime",
			value:   "2023-10-15",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := &Datetime{}
			err := dt.Scan(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Datetime.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Benchmark tests
func BenchmarkDatetime_Scan_String(b *testing.B) {
	dt := &Datetime{}
	value := "2023-10-15 14:30:45"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dt.Scan(value)
	}
}

func BenchmarkDatetime_Scan_Bytes(b *testing.B) {
	dt := &Datetime{}
	value := []byte("2023-10-15 14:30:45")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dt.Scan(value)
	}
}

func BenchmarkDatetime_Scan_Time(b *testing.B) {
	dt := &Datetime{}
	value := time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dt.Scan(value)
	}
}

func BenchmarkDatetime_Value(b *testing.B) {
	dt := Datetime{Time: time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = dt.Value()
	}
}
