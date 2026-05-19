package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKSTDatetime(t *testing.T) {
	kst, err := time.LoadLocation("Asia/Seoul")
	require.NoError(t, err)

	tests := []struct {
		name       string
		input      string
		wantUTC    time.Time
		wantErrMsg string
	}{
		{
			name:    "오전 시각 KST→UTC 변환",
			input:   "2024/06/15 09:00",
			wantUTC: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "자정 KST는 전날 UTC로 변환",
			input:   "2024/06/15 00:00",
			wantUTC: time.Date(2024, 6, 14, 15, 0, 0, 0, time.UTC),
		},
		{
			name:    "오후 시각 KST→UTC 변환",
			input:   "2024/06/15 23:59",
			wantUTC: time.Date(2024, 6, 15, 14, 59, 0, 0, time.UTC),
		},
		{
			name:    "KST 정오 변환",
			input:   "2024/01/01 12:00",
			wantUTC: time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC),
		},
		{
			name:       "잘못된 날짜 구분자",
			input:      "2024-06-15 09:00",
			wantErrMsg: "failed to parse time",
		},
		{
			name:       "잘못된 포맷",
			input:      "2024/06/15",
			wantErrMsg: "failed to parse time",
		},
		{
			name:       "빈 문자열",
			input:      "",
			wantErrMsg: "failed to parse time",
		},
		{
			name:       "숫자가 아닌 입력",
			input:      "abcd/ef/gh 00:00",
			wantErrMsg: "failed to parse time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseKSTDatetime(tt.input)

			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUTC, result.Time)
			assert.Equal(t, time.UTC, result.Time.Location())

			// 원본 KST 시각과 오프셋 9시간 차이 검증
			parsedKST, _ := time.ParseInLocation("2006/01/02 15:04", tt.input, kst)
			assert.Equal(t, parsedKST.UTC(), result.Time)
		})
	}
}

func TestParseRunningTimeSec(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantSec    uint64
		wantErrMsg string
	}{
		{
			name:    "모든 단위 포함",
			input:   "1day 2hour 3min 4sec",
			wantSec: 1*86400 + 2*3600 + 3*60 + 4,
		},
		{
			name:    "전부 0",
			input:   "0day 0hour 0min 0sec",
			wantSec: 0,
		},
		{
			name:    "일(day)만 존재",
			input:   "2day 0hour 0min 0sec",
			wantSec: 2 * 86400,
		},
		{
			name:    "시(hour)만 존재",
			input:   "0day 5hour 0min 0sec",
			wantSec: 5 * 3600,
		},
		{
			name:    "분(min)만 존재",
			input:   "0day 0hour 30min 0sec",
			wantSec: 30 * 60,
		},
		{
			name:    "초(sec)만 존재",
			input:   "0day 0hour 0min 45sec",
			wantSec: 45,
		},
		{
			name:    "큰 값 (365일)",
			input:   "365day 23hour 59min 59sec",
			wantSec: 365*86400 + 23*3600 + 59*60 + 59,
		},
		{
			name:       "빈 문자열",
			input:      "",
			wantErrMsg: "failed to parse running time",
		},
		{
			name:       "포맷 불일치",
			input:      "1d 2h 3m 4s",
			wantErrMsg: "failed to parse running time",
		},
		{
			name:       "숫자 없음",
			input:      "day hour min sec",
			wantErrMsg: "failed to parse running time",
		},
		{
			name:       "일부 단위만 있음",
			input:      "1day 2hour",
			wantErrMsg: "failed to parse running time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseRunningTimeSec(tt.input)

			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSec, result)
		})
	}
}
