# pkg/control/topology

STB 제어 대상 선택을 위한 네트워크 토폴로지 계층 조회 API.

## 패키지 구조

| 파일         | 역할                                 |
| ------------ | ------------------------------------ |
| `api.go`     | `Api` 구조체 — 초기화 및 라우트 등록 |
| `handler.go` | SO/L3/Cell/STB 목록 조회 핸들러      |

## 베이스 경로

`control_common.HttpRelativePath` = `/catv/control`

## API

| 메서드 | 경로                                                                      | 설명                      |
| ------ | ------------------------------------------------------------------------- | ------------------------- |
| `GET`  | `/catv/control/topology/sos`                                              | SO 목록 조회              |
| `GET`  | `/catv/control/topology/sos/:so_id/l3s`                                   | 특정 SO의 L3 목록 조회    |
| `GET`  | `/catv/control/topology/sos/:so_id/l3s/:l3_id/cells`                      | 특정 L3의 Cell 목록 조회  |
| `GET`  | `/catv/control/topology/sos/:so_id/l3s/:l3_id/cells/:cell_id/settopboxes` | 특정 Cell의 STB 목록 조회 |

## 응답 형식

```json
{
  "items": ["SO-01", "SO-02"],
  "count": 2
}
```

## 계층 구조

```
SO (Service Office)
└── L3 (L3 네트워크 영역)
    └── Cell (기지국/셀)
        └── Settopbox (셋탑박스, MAC 주소 기준)
```

각 계층은 상위 계층 ID를 경로 파라미터로 받아 하위 목록을 반환합니다.  
내부적으로 `tables.StbInformationTable`을 통해 ClickHouse DB에서 조회합니다.
