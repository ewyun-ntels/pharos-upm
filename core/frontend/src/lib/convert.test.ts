import { responseConvert } from './convert';
import { ChartMetricData } from '@types';
import {Index} from '@pharos/shared/types/datasource';

describe('responseConvert', () => {
  const mockEmptyResult: ChartMetricData = {
    chartMetric: [],
    chartType: 'timeseries',
    endTime: 0,
    startTime: 0,
    step: 0,
    uniqueKeys: [],
  };

  describe('Korean column names data conversion', () => {
    it('should convert Korean percentage columns correctly', () => {
      const inputData: Index = {
        meta: [
          { name: "timestamp" },
          { name: "불량률(%)" },
          { name: "위험률(%)" },
          { name: "정상률(%)" }
        ],
        data: [
          {
            "timestamp": "2025-12-29T06:05:00Z",
            "불량률(%)": 13.39,
            "위험률(%)": 3.88,
            "정상률(%)": 82.45
          },
          {
            "timestamp": "2025-12-29T06:06:00Z",
            "불량률(%)": 13.82,
            "위험률(%)": 3.95,
            "정상률(%)": 81.85
          }
        ],
        rows: 2,
        statistics: { elapsed: 0.1 }
      };

      const result = responseConvert(
        { ...mockEmptyResult },
        inputData,
        'test query',
        'test name',
        'test label',
        1735456300,  // start time
        1735456360,  // end time
        60           // step
      );

      // 검증: chartMetric이 생성되었는지
      expect(result.chartMetric.length).toBeGreaterThan(0);

      // 검증: uniqueKeys가 3개인지 (각 퍼센트 컬럼당 하나씩)
      expect(result.uniqueKeys).toHaveLength(3);

      // 검증: 각 uniqueKey에 한국어 컬럼명이 포함되어 있는지
      const uniqueKeysString = (result.uniqueKeys || []).join('');
      expect(uniqueKeysString).toContain('불량률(%)');
      expect(uniqueKeysString).toContain('위험률(%)');
      expect(uniqueKeysString).toContain('정상률(%)');

      // 검증: chartMetricMap이 올바른 크기인지
      expect(result.chartMetricMap).toBeDefined();
      expect(result.chartMetricMap!.size).toBe(2); // 2개 타임스탬프

      // 검증: 시간 설정이 올바른지 (밀리초로 반환됨)
      expect(result.startTime).toBe(1735456300000);
      expect(result.endTime).toBe(1735456360000);
      expect(result.step).toBe(60000);
    });

    it('should handle timestamps correctly', () => {
      const inputData: Index = {
        meta: [
          { name: "timestamp" },
          { name: "value" }
        ],
        data: [
          {
            "timestamp": "2025-12-29T06:05:00Z",
            "value": 100
          }
        ],
        rows: 1,
        statistics: { elapsed: 0.1 }
      };

      const result = responseConvert(
        { ...mockEmptyResult },
        inputData,
        'test query',
        'test name',
        'test label',
        1735456300,
        1735456360,
        60
      );

      // 검증: timestamp가 올바르게 변환되었는지 (UTC ISO string → unix timestamp 밀리초)
      const firstMetric = result.chartMetric[0];
      expect(firstMetric.timestamp).toBe(1766988300000); // 2025-12-29T06:05:00Z의 unix timestamp (밀리초)
    });

    it('should filter out non-numeric columns as labels', () => {
      const inputData: Index = {
        meta: [
          { name: "timestamp" },
          { name: "region" },
          { name: "cpu_usage" },
          { name: "status" }
        ],
        data: [
          {
            "timestamp": "2025-12-29T06:05:00Z",
            "region": "seoul",
            "cpu_usage": 75.5,
            "status": "normal"
          }
        ],
        rows: 1,
        statistics: { elapsed: 0.1 }
      };

      const result = responseConvert(
        { ...mockEmptyResult },
        inputData,
        'test query',
        'test name',
        'test label',
        1735456300,
        1735456360,
        60
      );

      // 검증: 숫자 값만 차트 데이터로 변환되었는지 (cpu_usage만)
      expect(result.uniqueKeys).toHaveLength(1);

      // uniqueKeys는 JSON 문자열이므로 파싱해서 검증
      const parsedLabel = JSON.parse((result.uniqueKeys || [])[0]);

      // keyColumn이 있을 때는 문자열 필드들이 라벨로 포함됨
      expect(parsedLabel.metric).toHaveProperty('region', 'seoul');
      expect(parsedLabel.metric).toHaveProperty('status', 'normal');

      // columnName은 keyColumn이 없을 때만 설정됨
      // 현재는 region, status가 문자열이므로 keyColumn에 포함되어 columnName이 설정되지 않음
    });

    it('should keep hexadecimal string labels out of numeric fields', () => {
      const inputData: Index = {
        meta: [
          { name: "timestamp" },
          { name: "msg_type" },
          { name: "reason" },
          { name: "value" }
        ],
        data: [
          {
            "timestamp": "2026-05-26T08:49:00Z",
            "msg_type": "0x0f",
            "reason": "nats_request_failed",
            "value": "0"
          }
        ],
        rows: 1,
        statistics: { elapsed: 0.1 }
      };

      const result = responseConvert(
        { ...mockEmptyResult },
        inputData,
        'test query',
        'test name',
        'test label',
        1779785340,
        1779785400,
        30
      );

      expect(result.uniqueKeys).toHaveLength(1);

      const parsedLabel = JSON.parse((result.uniqueKeys || [])[0]);
      expect(parsedLabel.metric).toHaveProperty('msg_type', '0x0f');
      expect(parsedLabel.metric).toHaveProperty('reason', 'nats_request_failed');
      expect(result.chartMetric[0][(result.uniqueKeys || [])[0]]).toBe(0);
    });

    it('should handle empty data gracefully', () => {
      const inputData: Index = {
        meta: [],
        data: [],
        rows: 0,
        statistics: { elapsed: 0.0 }
      };

      const result = responseConvert(
        { ...mockEmptyResult },
        inputData,
        'test query',
        'test name',
        'test label',
        1735456300,
        1735456360,
        60
      );

      // 검증: 빈 데이터 처리
      expect(result.chartMetric).toHaveLength(0);
      expect(result.uniqueKeys).toHaveLength(0);
      expect(result.chartMetricMap!.size).toBe(0);
    });

    it('should handle null data gracefully', () => {
      const inputData: Index = {
        meta: [{ name: "timestamp" }, { name: "value" }],
        data: [
          {
            "timestamp": "2025-12-29T06:05:00Z",
            "value": null
          }
        ],
        rows: 1,
        statistics: { elapsed: 0.1 }
      };

      const result = responseConvert(
        { ...mockEmptyResult },
        inputData,
        'test query',
        'test name',
        'test label',
        1735456300,
        1735456360,
        60
      );

      // 검증: null 값은 제외되었는지
      expect(result.uniqueKeys).toHaveLength(0);
      expect(result.chartMetric).toHaveLength(0);
    });
  });

  describe('Multi-query merge', () => {
    it('should merge multiple queries with same timestamps', () => {
      // 첫 번째 쿼리 결과
      const firstQueryData: Index = {
        meta: [{ name: "timestamp" }, { name: "cpu" }],
        data: [
          { "timestamp": "2025-12-29T06:05:00Z", "cpu": 50 },
          { "timestamp": "2025-12-29T06:06:00Z", "cpu": 60 }
        ],
        rows: 2,
        statistics: { elapsed: 0.1 }
      };

      // 첫 번째 쿼리 처리
      const result = responseConvert(
        { ...mockEmptyResult },
        firstQueryData,
        'query1',
        'name1',
        'label1',
        1735456300,
        1735456360,
        60
      );

      // 두 번째 쿼리 결과 (같은 timestamp)
      const secondQueryData: Index = {
        meta: [{ name: "timestamp" }, { name: "memory" }],
        data: [
          { "timestamp": "2025-12-29T06:05:00Z", "memory": 70 },
          { "timestamp": "2025-12-29T06:06:00Z", "memory": 80 }
        ],
        rows: 2,
        statistics: { elapsed: 0.1 }
      };

      // 두 번째 쿼리 병합
      const mergedResult = responseConvert(
        result,
        secondQueryData,
        'query2',
        'name2',
        'label2',
        1735456300,
        1735456360,
        60
      );

      // 검증: 2개 timestamp 유지
      expect(mergedResult.chartMetric.length).toBe(2);
      expect(mergedResult.chartMetricMap!.size).toBe(2);

      // 검증: 각 timestamp에 두 쿼리의 메트릭이 모두 존재
      const firstTimestamp = mergedResult.chartMetric[0];
      const keys = Object.keys(firstTimestamp).filter(k => k !== 'timestamp');
      expect(keys.length).toBe(2); // cpu + memory 메트릭

      // 검증: uniqueKeys가 병합됨 (2개 쿼리 × 1개 메트릭 = 2개)
      expect(mergedResult.uniqueKeys!.length).toBe(2);
    });

    it('should handle different timestamp ranges between queries', () => {
      // 첫 번째 쿼리: 06:05, 06:06
      const firstQueryData: Index = {
        meta: [{ name: "timestamp" }, { name: "metric_a" }],
        data: [
          { "timestamp": "2025-12-29T06:05:00Z", "metric_a": 10 },
          { "timestamp": "2025-12-29T06:06:00Z", "metric_a": 20 }
        ],
        rows: 2,
        statistics: { elapsed: 0.1 }
      };

      const result = responseConvert(
        { ...mockEmptyResult },
        firstQueryData,
        'query1',
        'name1',
        'label1',
        1735456300,
        1735456360,
        60
      );

      // 두 번째 쿼리: 06:07, 06:08 (다른 timestamp 구간)
      const secondQueryData: Index = {
        meta: [{ name: "timestamp" }, { name: "metric_b" }],
        data: [
          { "timestamp": "2025-12-29T06:07:00Z", "metric_b": 30 },
          { "timestamp": "2025-12-29T06:08:00Z", "metric_b": 40 }
        ],
        rows: 2,
        statistics: { elapsed: 0.1 }
      };

      const mergedResult = responseConvert(
        result,
        secondQueryData,
        'query2',
        'name2',
        'label2',
        1735456300,
        1735456480,
        60
      );

      // 검증: 4개 timestamp (각 쿼리에서 2개씩)
      expect(mergedResult.chartMetric.length).toBe(4);
      expect(mergedResult.chartMetricMap!.size).toBe(4);

      // 검증: timestamp 정렬 확인
      const timestamps = mergedResult.chartMetric.map(m => m.timestamp);
      expect(timestamps).toEqual([...timestamps].sort((a, b) => a - b));

      // 검증: 각 timestamp에는 해당 쿼리의 메트릭만 존재
      // 첫 번째 timestamp (06:05)에는 metric_a만 있음
      const firstMetric = mergedResult.chartMetric[0];
      const firstKeys = Object.keys(firstMetric).filter(k => k !== 'timestamp');
      expect(firstKeys.length).toBe(1); // metric_a만

      // 마지막 timestamp (06:08)에는 metric_b만 있음
      const lastMetric = mergedResult.chartMetric[3];
      const lastKeys = Object.keys(lastMetric).filter(k => k !== 'timestamp');
      expect(lastKeys.length).toBe(1); // metric_b만
    });

    it('should handle partially overlapping timestamp ranges', () => {
      // 첫 번째 쿼리: 06:05, 06:06, 06:07
      const firstQueryData: Index = {
        meta: [{ name: "timestamp" }, { name: "cpu" }],
        data: [
          { "timestamp": "2025-12-29T06:05:00Z", "cpu": 10 },
          { "timestamp": "2025-12-29T06:06:00Z", "cpu": 20 },
          { "timestamp": "2025-12-29T06:07:00Z", "cpu": 30 }
        ],
        rows: 3,
        statistics: { elapsed: 0.1 }
      };

      const result = responseConvert(
        { ...mockEmptyResult },
        firstQueryData,
        'query1',
        'name1',
        'label1',
        1735456300,
        1735456420,
        60
      );

      // 두 번째 쿼리: 06:06, 06:07, 06:08 (부분 겹침)
      const secondQueryData: Index = {
        meta: [{ name: "timestamp" }, { name: "memory" }],
        data: [
          { "timestamp": "2025-12-29T06:06:00Z", "memory": 50 },
          { "timestamp": "2025-12-29T06:07:00Z", "memory": 60 },
          { "timestamp": "2025-12-29T06:08:00Z", "memory": 70 }
        ],
        rows: 3,
        statistics: { elapsed: 0.1 }
      };

      const mergedResult = responseConvert(
        result,
        secondQueryData,
        'query2',
        'name2',
        'label2',
        1735456300,
        1735456480,
        60
      );

      // 검증: 4개 unique timestamps (06:05, 06:06, 06:07, 06:08)
      expect(mergedResult.chartMetric.length).toBe(4);

      // 검증: 겹치는 timestamp (06:06, 06:07)에는 두 메트릭 모두 존재
      const metric0606 = mergedResult.chartMetric[1]; // 06:06
      const keys0606 = Object.keys(metric0606).filter(k => k !== 'timestamp');
      expect(keys0606.length).toBe(2); // cpu + memory

      // 검증: 겹치지 않는 timestamp (06:05)에는 cpu만
      const metric0605 = mergedResult.chartMetric[0]; // 06:05
      const keys0605 = Object.keys(metric0605).filter(k => k !== 'timestamp');
      expect(keys0605.length).toBe(1); // cpu only

      // 검증: 겹치지 않는 timestamp (06:08)에는 memory만
      const metric0608 = mergedResult.chartMetric[3]; // 06:08
      const keys0608 = Object.keys(metric0608).filter(k => k !== 'timestamp');
      expect(keys0608.length).toBe(1); // memory only
    });

    it('should preserve uniqueKeys from all queries', () => {
      const firstQueryData: Index = {
        meta: [{ name: "timestamp" }, { name: "a" }, { name: "b" }],
        data: [{ "timestamp": "2025-12-29T06:05:00Z", "a": 1, "b": 2 }],
        rows: 1,
        statistics: { elapsed: 0.1 }
      };

      const result = responseConvert(
        { ...mockEmptyResult },
        firstQueryData,
        'query1',
        'name1',
        'label1',
        1735456300,
        1735456360,
        60
      );

      const secondQueryData: Index = {
        meta: [{ name: "timestamp" }, { name: "c" }],
        data: [{ "timestamp": "2025-12-29T06:05:00Z", "c": 3 }],
        rows: 1,
        statistics: { elapsed: 0.1 }
      };

      const mergedResult = responseConvert(
        result,
        secondQueryData,
        'query2',
        'name2',
        'label2',
        1735456300,
        1735456360,
        60
      );

      // 검증: 3개 uniqueKeys (a, b from query1 + c from query2)
      expect(mergedResult.uniqueKeys!.length).toBe(3);
    });
  });

  describe('Time handling', () => {
    it('should handle different time field names', () => {
      const testCases = [
        { field: 'time', value: '2025-12-29T06:05:00Z' },
        { field: 'TIME', value: '2025-12-29T06:05:00Z' },
        { field: 'timestamp', value: '2025-12-29T06:05:00Z' }
      ];

      testCases.forEach(({ field, value }) => {
        const inputData: Index = {
          meta: [{ name: field }, { name: 'value' }],
          data: [
            {
              [field]: value,
              'value': 100
            }
          ],
          rows: 1,
          statistics: { elapsed: 0.1 }
        };

        const result = responseConvert(
          { ...mockEmptyResult },
          inputData,
          'test query',
          'test name',
          'test label',
          1735456300,
          1735456360,
          60
        );

        expect(result.chartMetric).toHaveLength(1);
        expect(result.chartMetric[0].timestamp).toBe(1766988300000); // 밀리초
      });
    });
  });
});
