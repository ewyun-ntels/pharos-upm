# Dashboard Schema Documentation

Dashboard 시스템의 JSON Schema 정의와 사용법을 설명합니다.

## 📋 Schema Files

### 1. DashboardConfig.schema
Dashboard 설정 및 레이아웃 정의를 위한 스키마입니다.

**위치**: `shared/schema/dashboard/DashboardConfig.schema`  
**용도**: Dashboard 구성 정보 관리  
**생성 타입**: `shared/types/dashboardconfig.Types`

### 2. DashboardPermission.schema
Dashboard 권한 관리를 위한 스키마입니다.

**위치**: `shared/schema/dashboard/DashboardPermission.schema`  
**용도**: 사용자별 Dashboard 접근 권한 관리  
**생성 타입**: `shared/types/dashboardpermission.Types`

## 🏗️ 아키텍처 패턴

```
Frontend ←→ shared/types ←→ Dashboard Engine ←→ Backend
            ↓ 자동 생성      ↓ 렌더링           ↓ 데이터 처리
        JSON Schema      Chart/Widget      Database/API
```

## 📖 Dashboard Config 사용 예제

### Dashboard 설정 구조
```json
{
  "id": "main-dashboard",
  "name": "Main System Dashboard",
  "description": "System overview and key metrics",
  "layout": "grid",
  "refresh_interval": 30,
  "cards": [
    {
      "id": "cpu-usage",
      "title": "CPU Usage",
      "type": "chart",
      "position": {
        "x": 0,
        "y": 0,
        "w": 6,
        "h": 4
      },
      "chart_options": {
        "chart_type": "line",
        "scales": {
          "y": {
            "min": 0,
            "max": 100
          }
        }
      },
      "data_provider": {
        "dataProviderName": "prometheus",
        "resource": "metrics",
        "chartQuery": {
          "query": "cpu_usage_percent{instance=\"server-1\"}",
          "queryName": "cpu_metrics"
        }
      }
    }
  ],
  "variables": {
    "time_range": "1h",
    "refresh_rate": "30s"
  }
}
```

### Card 타입별 예제

#### 1. Chart Card (차트)
```json
{
  "id": "memory-chart",
  "title": "Memory Usage",
  "type": "chart",
  "chart_options": {
    "chart_type": "area",
    "unit_type": "bytes",
    "show_legend": true
  },
  "data_provider": {
    "dataProviderName": "prometheus", 
    "resource": "metrics",
    "chartQuery": {
      "query": "memory_usage_bytes / memory_total_bytes * 100"
    }
  }
}
```

#### 2. Stat Card (통계)
```json
{
  "id": "user-count",
  "title": "Active Users",
  "type": "stat",
  "chart_options": {
    "show_type": "last",
    "unit_type": "count",
    "show_value": true
  },
  "data_provider": {
    "dataProviderName": "postgresql",
    "resource": "database", 
    "chartQuery": {
      "query": "SELECT COUNT(*) FROM users WHERE status = 'active'"
    }
  }
}
```

#### 3. Table Card (테이블)
```json
{
  "id": "top-errors",
  "title": "Top Errors",
  "type": "table",
  "chart_options": {
    "show_type": "table",
    "max_rows": 10
  },
  "data_provider": {
    "dataProviderName": "clickhouse",
    "resource": "logs",
    "chartQuery": {
      "query": "SELECT error_message, COUNT(*) as count FROM error_logs GROUP BY error_message ORDER BY count DESC LIMIT 10"
    }
  }
}
```

## 📖 Dashboard Permission 사용 예제

### 권한 설정 구조
```json
{
  "dashboard_id": "main-dashboard",
  "user_id": "user123",
  "permissions": {
    "read": true,
    "write": true,
    "delete": false,
    "share": true
  },
  "role": "editor",
  "created_at": "2025-09-27T00:00:00Z",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

### 권한 레벨

#### Admin (관리자)
```json
{
  "role": "admin",
  "permissions": {
    "read": true,
    "write": true, 
    "delete": true,
    "share": true,
    "manage_permissions": true
  }
}
```

#### Editor (편집자)
```json
{
  "role": "editor",
  "permissions": {
    "read": true,
    "write": true,
    "delete": false,
    "share": true
  }
}
```

#### Viewer (조회자)  
```json
{
  "role": "viewer",
  "permissions": {
    "read": true,
    "write": false,
    "delete": false, 
    "share": false
  }
}
```

## 🎯 지원하는 Data Provider

### Prometheus
```json
{
  "dataProviderName": "prometheus",
  "resource": "metrics",
  "chartQuery": {
    "query": "rate(http_requests_total[5m])",
    "queryName": "request_rate"
  }
}
```

### PostgreSQL
```json
{
  "dataProviderName": "postgresql",
  "resource": "database", 
  "chartQuery": {
    "query": "SELECT date_trunc('hour', created_at) as time, COUNT(*) FROM orders GROUP BY time"
  }
}
```

### ClickHouse
```json
{
  "dataProviderName": "clickhouse",
  "resource": "analytics",
  "chartQuery": {
    "query": "SELECT toStartOfHour(timestamp) as time, avg(response_time) FROM requests GROUP BY time"
  }
}
```

## 📊 Chart Options

### Line Chart
```json
{
  "chart_type": "line",
  "scales": {
    "x": {"type": "time"},
    "y": {"min": 0, "max": 100}
  },
  "show_legend": true,
  "unit_type": "percentage"
}
```

### Bar Chart
```json
{
  "chart_type": "bar", 
  "orientation": "vertical",
  "show_values": true,
  "color_palette": ["#FF6384", "#36A2EB", "#FFCE56"]
}
```

### Pie Chart
```json
{
  "chart_type": "pie",
  "show_legend": true,
  "show_percentages": true,
  "donut": false
}
```

## 📱 Layout 시스템

### Grid Layout
```json
{
  "layout": "grid",
  "grid_options": {
    "cols": 12,
    "row_height": 60,
    "margin": [10, 10],
    "container_padding": [20, 20]
  }
}
```

### Position 설정
```json
{
  "position": {
    "x": 0,     // Grid column position (0-11)
    "y": 0,     // Grid row position  
    "w": 6,     // Width (grid units)
    "h": 4      // Height (grid units)
  }
}
```

## ✅ 검증된 기능

- ✅ 다양한 차트 타입 지원 (line, bar, pie, area, table, stat)
- ✅ 여러 데이터소스 통합 (Prometheus, PostgreSQL, ClickHouse)
- ✅ 유연한 Grid 레이아웃 시스템
- ✅ 세밀한 권한 관리 (role-based access control)  
- ✅ 실시간 데이터 refresh
- ✅ Frontend와 Backend 타입 동기화

## 🔗 관련 파일

- **Frontend**: `core/frontend/src/types/charts.ts`
- **Schema**: `shared/schema/dashboard/`
- **Types**: `shared/types/dashboardconfig/`, `shared/types/dashboardpermission/`

## 🎯 장점

1. **통합 설정**: 모든 dashboard 구성을 JSON Schema로 통합 관리
2. **다양한 시각화**: Chart.js 기반 풍부한 차트 옵션 제공
3. **멀티 데이터소스**: 여러 데이터베이스/API를 하나의 대시보드에서 조회
4. **권한 기반 접근**: 사용자별/역할별 세밀한 접근 제어
5. **반응형 레이아웃**: Grid 기반 자유로운 레이아웃 구성
6. **실시간 업데이트**: 설정 가능한 자동 refresh 지원