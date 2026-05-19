# UI 설정 API

이 패키지는 UI 설정 관리 및 정적 이미지 자산을 위한 REST API 엔드포인트를 제공합니다.

## 개요

UI Config 패키지는 두 가지 주요 리소스를 관리합니다:
- **설정**: 데이터베이스에 저장된 JSON 설정 데이터
- **이미지**: 임베디드 파일시스템에서 제공되는 정적 이미지 파일

## API 엔드포인트

### 설정 관리

#### GET /ui-config/config
데이터베이스에서 현재 UI 설정을 조회합니다.

**응답:**
- **200 OK**: 설정 JSON 객체 반환
- **204 No Content**: 설정이 없는 경우 (`null` 반환)
- **500 Internal Server Error**: 데이터베이스 또는 파싱 오류

```json
{
  "theme": "dark",
  "language": "ko",
  "features": {
    "enableNotifications": true,
    "maxItems": 100
  }
}
```

#### POST /ui-config/config
새로운 UI 설정 항목을 생성합니다.

**요청 본문:**
- Content-Type: `application/json`
- 원시 JSON 설정 객체

**응답:**
- **200 OK**: 설정이 성공적으로 생성됨
- **500 Internal Server Error**: 데이터베이스 오류

**예제:**
```bash
curl -X 'POST' ${address}/ui-config/config \
     -H 'accept: application/json' \
     -H 'Content-Type: application/json' \
     -d '{
  "theme": "light",
  "language": "ko",
  "features": {
    "enableNotifications": false,
    "maxItems": 50
  }
}'
```

#### PUT /ui-config/config
기존 UI 설정을 업데이트합니다.

**요청 본문:**
- Content-Type: `application/json`
- 원시 JSON 설정 객체

**응답:**
- **200 OK**: 설정이 성공적으로 업데이트됨
- **500 Internal Server Error**: 데이터베이스 오류

**예제:**
```bash
curl -X 'PUT' ${address}/ui-config/config \
     -H 'accept: application/json' \
     -H 'Content-Type: application/json' \
     -d '{
  "theme": "dark",
  "language": "ko"
}'
```

#### DELETE /ui-config/config
모든 UI 설정 데이터를 삭제합니다.

**응답:**
- **200 OK**: 설정이 성공적으로 삭제됨
- **500 Internal Server Error**: 데이터베이스 오류

**예제:**
```bash
curl -X 'DELETE' ${address}/ui-config/config
```

### 이미지 자산

#### GET /ui-config/images
사용 가능한 모든 이미지 파일 목록을 조회합니다.

**응답:**
- **200 OK**: 이미지 파일명 배열
- **500 Internal Server Error**: 파일시스템 오류

```json
["eye-slash.svg", "eye.svg", "logo.png", "icons/home.svg"]
```

#### GET /ui-config/images/:name
특정 이미지 파일을 제공합니다.

**매개변수:**
- `name` (경로): 이미지 파일명 (하위 디렉토리가 있는 경우 포함)

**응답:**
- **200 OK**: 적절한 Content-Type 헤더와 함께 이미지 파일 반환
- **500 Internal Server Error**: 파일을 찾을 수 없거나 읽기 오류

**Content-Type 감지:**
- `.svg` 파일: `image/svg+xml`
- 기타 파일: 내용을 기반으로 자동 감지

**예제:**
```bash
curl ${address}/ui-config/images/eye.svg
# Content-Type: image/svg+xml로 SVG 내용 반환

curl ${address}/ui-config/images/logo.png
# Content-Type: image/png로 PNG 내용 반환
```

## 데이터베이스 스키마

이 패키지는 다음 구조의 `ui_config_config` 테이블을 사용합니다:
- `config` (TEXT): JSON 설정 데이터

## 오류 응답

모든 엔드포인트는 다음 형식의 오류 응답을 반환할 수 있습니다:
```json
{
  "message": "오류 설명"
}
```

## 구현 참고사항

- 설정 데이터는 데이터베이스에 원시 JSON으로 저장됩니다
- 이미지는 임베디드 파일시스템(`pharos.UiConfigImagesFS`)에서 제공됩니다
- 패키지는 JSON 키 순서를 유지하기 위해 정렬된 맵을 사용합니다
- JSON 응답의 형식을 유지하기 위해 HTML 이스케이프가 비활성화됩니다
