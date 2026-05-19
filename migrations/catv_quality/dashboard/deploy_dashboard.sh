#!/bin/bash

# Grafana 대시보드 배포 스크립트
GRAFANA_URL="http://192.168.15.102:30071"
GRAFANA_USER="admin"
GRAFANA_PASS="admin"
DASHBOARD_FILE="$1"

if [ -z "$DASHBOARD_FILE" ]; then
    echo "사용법: $0 <dashboard-json-file>"
    exit 1
fi

if [ ! -f "$DASHBOARD_FILE" ]; then
    echo "오류: 파일 '$DASHBOARD_FILE'을 찾을 수 없습니다."
    exit 1
fi

# 대시보드 제목 추출
DASHBOARD_TITLE=$(python3 -c "import json; f=open('$DASHBOARD_FILE'); d=json.load(f); print(d['dashboard']['title'])")

echo "=== 대시보드 배포: $DASHBOARD_TITLE ==="

# 동일한 제목의 대시보드 검색
echo "1. 기존 대시보드 검색 중..."
SEARCH_RESULT=$(curl -s -u "$GRAFANA_USER:$GRAFANA_PASS" "$GRAFANA_URL/api/search?query=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$DASHBOARD_TITLE'))")")

# 기존 대시보드 UID 추출 및 삭제
echo "$SEARCH_RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    uids = [item['uid'] for item in data if item.get('title') == '$DASHBOARD_TITLE']
    if uids:
        print('   기존 대시보드 발견:', len(uids), '개')
        for uid in uids:
            print('   UID:', uid)
    else:
        print('   기존 대시보드 없음')
except:
    print('   검색 결과 파싱 실패')
" > /tmp/dashboard_uids.txt

cat /tmp/dashboard_uids.txt

# UID 추출 및 삭제 실행
UIDS=$(echo "$SEARCH_RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    uids = [item['uid'] for item in data if item.get('title') == '$DASHBOARD_TITLE']
    for uid in uids:
        print(uid)
except:
    pass
")

if [ ! -z "$UIDS" ]; then
    echo "2. 기존 대시보드 삭제 중..."
    for uid in $UIDS; do
        echo "   삭제: $uid"
        curl -s -X DELETE -u "$GRAFANA_USER:$GRAFANA_PASS" "$GRAFANA_URL/api/dashboards/uid/$uid" > /dev/null
    done
else
    echo "2. 삭제할 대시보드 없음"
fi

# 새 대시보드 업로드
echo "3. 새 대시보드 업로드 중..."
RESULT=$(curl -s -X POST -H "Content-Type: application/json" -u "$GRAFANA_USER:$GRAFANA_PASS" "$GRAFANA_URL/api/dashboards/db" -d @"$DASHBOARD_FILE")

# 결과 출력
echo "$RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if data.get('status') == 'success':
        print('\n✅ 배포 성공!')
        print('   URL:', '$GRAFANA_URL' + data['url'])
        print('   UID:', data['uid'])
    else:
        print('\n❌ 배포 실패!')
        print('   메시지:', data.get('message', '알 수 없는 오류'))
except Exception as e:
    print('\n❌ 응답 파싱 실패:', e)
    print(sys.stdin.read())
"

echo ""
