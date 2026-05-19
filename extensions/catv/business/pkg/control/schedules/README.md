# pkg/control/schedules

STB 제어 스케줄 CRUD 및 실행 결과 조회 API.

## 패키지 구조

| 파일                  | 역할                                                                                         |
| --------------------- | -------------------------------------------------------------------------------------------- |
| `api.go`              | `Api` 구조체 — 초기화, 스케줄 로딩/등록/해제, cron 관리                                      |
| `handler.go`          | `getSchedulesResultsListHandler`, `getSchedulesResultsHandler` — 결과 목록/상세 핸들러       |
| `filter.go`           | `SchedulesResultsListQueryFilter`, `SchedulesResultsQueryFilter` — 쿼리 파라미터 바인딩·검증 |
| `schedule_request.go` | `ScheduleRequest` — 스케줄 생성/수정 요청 바디                                               |

## 베이스 경로

`control_common.HttpRelativePath` = `/catv/control`

## API

| 메서드   | 경로                                  | 설명                |
| -------- | ------------------------------------- | ------------------- |
| `GET`    | `/catv/control/schedules`             | 스케줄 목록 조회    |
| `POST`   | `/catv/control/schedules`             | 스케줄 생성         |
| `PUT`    | `/catv/control/schedules`             | 스케줄 수정         |
| `DELETE` | `/catv/control/schedules/:name`       | 스케줄 삭제         |
| `GET`    | `/catv/control/schedules/results`     | 실행 결과 목록 조회 |
| `GET`    | `/catv/control/schedules/results/:id` | 실행 결과 상세 조회 |

## 스케줄 타입

| 타입          | 설명                                                   |
| ------------- | ------------------------------------------------------ |
| `immediately` | 등록 즉시 1회 실행                                     |
| `once`        | `schedule_spec_once` 지정 시각에 1회 실행 후 등록 해제 |
| `repeat`      | `schedule_spec_repeat` cron 표현식으로 반복 실행       |

## 요청 바디 (`ScheduleRequest`)

```json
{
  "name": "test-schedule",
  "schedule_type": "once",
  "schedule_spec_once": "2026-05-01T09:00:00Z",
  "area_type": "settopbox",
  "area_ids": ["AA:BB:CC:DD:EE:FF"],
  "sos": [],
  "l3s": [],
  "cells": [],
  "settopboxes": ["AA:BB:CC:DD:EE:FF"],
  "work_type": "stb_request_info",
  "work_value": null
}
```

## 결과 목록 조회 파라미터 (`SchedulesResultsListQueryFilter`)

| 파라미터        | 타입   | 설명                               |
| --------------- | ------ | ---------------------------------- |
| `id`            | string | 특정 실행 결과 ID                  |
| `work_type`     | string | 제어 명령 타입                     |
| `schedule_id`   | string | 스케줄 ID                          |
| `schedule_name` | string | 스케줄 이름 (LIKE 검색)            |
| `k8s_job_name`  | string | K8s Job 이름                       |
| `from`          | string | 시작 일시 (ISO 8601)               |
| `to`            | string | 종료 일시 (ISO 8601)               |
| `limit`         | int    | 최대 결과 수 (기본 100, 최대 1000) |
| `offset`        | uint64 | 오프셋                             |

## 결과 상세 조회 파라미터 (`SchedulesResultsQueryFilter`)

| 파라미터          | 타입     | 설명                                                  |
| ----------------- | -------- | ----------------------------------------------------- |
| `k8s_job_name`    | string   | K8s Job 이름                                          |
| `cm_mac_addr`     | string   | CM MAC 주소 (후방 와일드카드 `*` 가능)                |
| `stb_mac_addr`    | string   | STB MAC 주소 (후방 와일드카드 `*` 가능)               |
| `stb_mdl_nm`      | string   | STB 모델명 (완전 일치)                                |
| `progress_status` | []string | 상태 필터 (`pending`/`running`/`succeeded`/`failed`)  |
| `result_code`     | string   | 결과 코드 (`1`=성공, `0`=기기 오류, `-1`=시스템 오류) |
| `search`          | string   | 전역 검색 (k8s_job_name, MAC, 모델명, 결과 메시지)    |
| `limit`           | int      | 최대 결과 수 (기본 1000, 최대 10000)                  |
| `offset`          | uint64   | 오프셋                                                |

## 스케줄 실행 흐름

```
Load() → DB에서 스케줄 목록 로드 → registerSchedule()
         ┌─ immediately → 즉시 channel 송신
         ├─ once        → time.AfterFunc → channel 송신 후 unregisterSchedule
         └─ repeat      → cron.AddFunc → channel 송신 (반복)

channelHandler() (goroutine)
  channel 수신 → settopbox.go의 실행 로직 호출

Unload() → cron 중지 → 모든 scheduleEntries 해제 → done 채널 닫기 → 대기
```
