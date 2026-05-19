package clickhouse_interface

import (
	"encoding/json"
	"testing"
)

func TestClickhouseBatch_AddBatchJson(t *testing.T) {
	tests := []struct {
		name    string
		data    []any
		wantErr bool
	}{
		{
			name:    "단일 문자열 데이터",
			data:    []any{"test"},
			wantErr: false,
		},
		{
			name:    "단일 숫자 데이터",
			data:    []any{123},
			wantErr: false,
		},
		{
			name:    "단일 구조체 데이터",
			data:    []any{map[string]any{"name": "test", "age": 30}},
			wantErr: false,
		},
		{
			name:    "여러 데이터 추가",
			data:    []any{"test1", "test2", "test3"},
			wantErr: false,
		},
		{
			name:    "nil 데이터",
			data:    []any{nil},
			wantErr: false,
		},
		{
			name:    "빈 배열",
			data:    []any{[]string{}},
			wantErr: false,
		},
		{
			name:    "복잡한 구조체",
			data:    []any{map[string]any{"users": []string{"user1", "user2"}, "count": 2}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batch := &ClickhouseBatch{}

			for _, data := range tt.data {
				err := batch.AddBatchJson(data)
				if (err != nil) != tt.wantErr {
					t.Errorf("ClickhouseBatch.AddBatchJson() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}

			// 데이터가 제대로 추가되었는지 확인
			if !tt.wantErr && len(batch.batchData) == 0 {
				t.Error("AddBatchJson() 후 batchData가 비어있음")
			}
		})
	}
}

func TestClickhouseBatch_AddBatchJson_WithUnmarshallableData(t *testing.T) {
	batch := &ClickhouseBatch{}

	// JSON으로 마샬링할 수 없는 데이터 (채널)
	invalidData := make(chan int)

	err := batch.AddBatchJson(invalidData)
	if err == nil {
		t.Error("마샬링할 수 없는 데이터에 대해 에러가 발생하지 않음")
	}
}

func TestClickhouseBatch_Get(t *testing.T) {
	tests := []struct {
		name      string
		setupData []any
		expectNil bool
	}{
		{
			name:      "데이터가 있는 경우",
			setupData: []any{"test1", "test2"},
			expectNil: false,
		},
		{
			name:      "데이터가 없는 경우",
			setupData: []any{},
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batch := &ClickhouseBatch{}

			// 테스트 데이터 설정
			for _, data := range tt.setupData {
				err := batch.AddBatchJson(data)
				if err != nil {
					t.Fatalf("데이터 설정 중 에러 발생: %v", err)
				}
			}

			// Get() 호출 전 데이터 길이 저장
			dataLengthBefore := len(batch.batchData)

			// Get() 호출
			result := batch.Get()

			// 결과 검증
			if tt.expectNil && len(result) != 0 {
				t.Error("빈 배치에서 빈 결과를 기대했지만 데이터가 반환됨")
			}

			if !tt.expectNil && len(result) == 0 {
				t.Error("데이터가 있는 배치에서 결과를 기대했지만 빈 데이터가 반환됨")
			}

			// Get() 호출 후 batchData가 nil인지 확인
			if batch.batchData != nil {
				t.Error("Get() 호출 후 batchData가 nil로 설정되지 않음")
			}

			// 반환된 데이터의 길이가 원래 데이터 길이와 같은지 확인
			if !tt.expectNil && len(result) != dataLengthBefore {
				t.Errorf("반환된 데이터 길이 = %d, 기대값 = %d", len(result), dataLengthBefore)
			}
		})
	}
}

func TestClickhouseBatch_Get_MultipleCallsReturnNil(t *testing.T) {
	batch := &ClickhouseBatch{}

	// 데이터 추가
	err := batch.AddBatchJson("test")
	if err != nil {
		t.Fatalf("데이터 추가 중 에러 발생: %v", err)
	}

	// 첫 번째 Get() 호출
	firstResult := batch.Get()
	if len(firstResult) == 0 {
		t.Error("첫 번째 Get() 호출에서 데이터가 반환되지 않음")
	}

	// 두 번째 Get() 호출
	secondResult := batch.Get()
	if len(secondResult) != 0 {
		t.Error("두 번째 Get() 호출에서 nil이 아닌 데이터가 반환됨")
	}
}

func TestClickhouseBatch_Integration(t *testing.T) {
	batch := &ClickhouseBatch{}

	// 여러 타입의 데이터 추가
	testData := []any{
		map[string]any{"id": 1, "name": "user1"},
		map[string]any{"id": 2, "name": "user2"},
		map[string]any{"id": 3, "name": "user3"},
	}

	for _, data := range testData {
		err := batch.AddBatchJson(data)
		if err != nil {
			t.Fatalf("데이터 추가 중 에러 발생: %v", err)
		}
	}

	// 배치 데이터 가져오기
	result := batch.Get()

	// 결과가 비어있지 않은지 확인
	if len(result) == 0 {
		t.Fatal("배치 결과가 비어있음")
	}

	// JSON 라인들로 분할 (마지막 개행 문자 제외)
	lines := []byte{}
	lineCount := 0
	for _, b := range result {
		if b == '\n' {
			lineCount++
			// 각 라인이 유효한 JSON인지 확인
			var temp any
			if len(lines) > 0 {
				if err := json.Unmarshal(lines, &temp); err != nil {
					t.Errorf("라인이 유효한 JSON이 아님: %s, 에러: %v", string(lines), err)
				}
			}
			lines = []byte{}
		} else {
			lines = append(lines, b)
		}
	}

	// 라인 수가 추가한 데이터 수와 같은지 확인
	if lineCount != len(testData) {
		t.Errorf("라인 수 = %d, 기대값 = %d", lineCount, len(testData))
	}

	// Get() 후 배치가 비어있는지 확인
	secondResult := batch.Get()
	if len(secondResult) != 0 {
		t.Error("두 번째 Get() 호출에서 데이터가 반환됨")
	}
}

// 벤치마크 테스트
func BenchmarkClickhouseBatch_AddBatchJson(b *testing.B) {
	batch := &ClickhouseBatch{}
	data := map[string]any{
		"id":   1,
		"name": "benchmark_user",
		"tags": []string{"tag1", "tag2", "tag3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch.AddBatchJson(data)
	}
}

func BenchmarkClickhouseBatch_Get(b *testing.B) {
	// 테스트 데이터로 배치 준비
	data := map[string]any{
		"id":   1,
		"name": "benchmark_user",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		batch := &ClickhouseBatch{}
		batch.AddBatchJson(data)
		b.StartTimer()

		batch.Get()
	}
}

// 예제 테스트
func ExampleClickhouseBatch_AddBatchJson() {
	batch := &ClickhouseBatch{}

	// 사용자 데이터 추가
	user1 := map[string]any{
		"id":   1,
		"name": "Alice",
		"age":  30,
	}
	user2 := map[string]any{
		"id":   2,
		"name": "Bob",
		"age":  25,
	}

	batch.AddBatchJson(user1)
	batch.AddBatchJson(user2)

	// 배치 데이터 가져오기
	data := batch.Get()

	// 결과 출력 (실제로는 JSON 라인들이 개행 문자로 구분됨)
	if len(data) > 0 {
		// 출력 예시를 위한 처리
	}
}

func ExampleClickhouseBatch_Get() {
	batch := &ClickhouseBatch{}

	// 데이터 추가
	batch.AddBatchJson(map[string]string{"message": "hello"})

	// 배치 데이터 가져오기 (이후 배치는 비워짐)
	data := batch.Get()

	if len(data) > 0 {
		// 데이터 처리
	}

	// 두 번째 호출은 빈 결과 반환
	emptyData := batch.Get()
	_ = emptyData // 빈 슬라이스
}
