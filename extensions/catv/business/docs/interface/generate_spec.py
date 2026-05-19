#!/usr/bin/env python3
"""
INTERFACE.spec.md 생성 스크립트
엑셀 원본과 100% 일치하는 규격서 생성 (모든 컬럼 포함)
"""

import json
from datetime import datetime


def generate_spec():
    """엑셀 원본 기반 SPEC 문서 생성"""

    with open('./interface_data.json', 'r', encoding='utf-8') as f:
        data = json.load(f)

    lines = []

    # ========================================
    # 헤더 및 경고
    # ========================================
    lines.append("# SKB CATV Log Agent 단말 인터페이스 정의서 v2.8")
    lines.append("")
    lines.append("> **⚠️ 자동 생성된 문서입니다**  ")
    lines.append("> 이 파일은 엑셀 원본에서 자동 생성되었습니다. 직접 수정하지 마세요.  ")
    lines.append(
        "> 수정이 필요한 경우 엑셀 파일을 업데이트한 후 `python generate_spec.py`를 실행하세요.")
    lines.append("")
    lines.append("---")
    lines.append("")

    # ========================================
    # 문서 정보
    # ========================================
    lines.append("## 문서 정보")
    lines.append("")
    lines.append("| 항목 | 내용 |")
    lines.append("|------|------|")
    lines.append("| 문서 버전 | v2.8 |")
    lines.append("| 최종 수정일 | 2024년 12월 8일 |")
    lines.append(
        f"| 생성일 | {datetime.now().strftime('%Y년 %m월 %d일 %H:%M:%S')} |")
    lines.append(
        "| 원본 파일 | 1_COMS_SKB_CATV_Log Agent 단말 인터페이스 정의서_v2.8(251208)_해제.xlsx |"
    )
    lines.append("")
    lines.append("---")
    lines.append("")

    # ========================================
    # 각 시트 처리
    # ========================================

    for sheet in data['sheets']:
        sheet_name = sheet['name']

        # 표지는 스킵
        if sheet_name == '표지':
            continue

        # ========================================
        # 개정이력
        # ========================================
        if sheet_name == '개정이력':
            lines.append("## 개정이력")
            lines.append("")

            # Row 2가 헤더 (Col 1부터)
            if len(sheet['rows']) > 2:
                header_row = sheet['rows'][2]
                headers = []
                for cell in header_row['cells'][1:5]:
                    val = cell.get('value', '')
                    if val:
                        headers.append(str(val))

                if headers:
                    lines.append("| " + " | ".join(headers) + " |")
                    lines.append("|" + "|".join([
                        ':----:' if i == 0 else ':--------:' if i ==
                        1 else '----------' if i == 2 else ':------:'
                        for i in range(len(headers))
                    ]) + "|")

                    # Row 3부터 데이터
                    for row_idx in range(3, len(sheet['rows'])):
                        row = sheet['rows'][row_idx]
                        if not row['cells'] or len(row['cells']) <= 1:
                            continue

                        first_val = row['cells'][1].get('value', '') if len(
                            row['cells']) > 1 else ''
                        if not first_val:
                            continue

                        row_data = []
                        for i in range(1, 5):
                            if i < len(row['cells']):
                                val = row['cells'][i].get('value', '')
                                if val:
                                    val = str(val).replace('\n', '<br>')
                                row_data.append(val if val else '')
                            else:
                                row_data.append('')

                        lines.append("| " + " | ".join(row_data) + " |")

            lines.append("")
            lines.append("---")
            lines.append("")
            continue

        # ========================================
        # 제어 응답 메시지 (특수 구조)
        # ========================================
        if sheet_name == '제어 응답 메시지':
            # Row 0: 타이틀
            if len(sheet['rows']) > 0:
                title = sheet['rows'][0]['cells'][0].get('value', '')
                if title:
                    lines.append(f"## {title}")
                    lines.append("")

            # Row 1: 헤더
            if len(sheet['rows']) > 1:
                header_row = sheet['rows'][1]
                headers = []
                for i in range(3):
                    if i < len(header_row['cells']):
                        val = header_row['cells'][i].get('value', '')
                        if val:
                            headers.append(str(val))

                if headers:
                    lines.append("| " + " | ".join(headers) + " |")
                    lines.append("|" + "|".join([
                        ':----:' if i == 0 else '------'
                        for i in range(len(headers))
                    ]) + "|")

                    # Row 2부터 데이터
                    for row_idx in range(2, len(sheet['rows'])):
                        row = sheet['rows'][row_idx]
                        if not row['cells']:
                            continue

                        # Col 1에 값이 있으면 데이터 행
                        message_val = row['cells'][1].get('value', '') if len(
                            row['cells']) > 1 else ''
                        if not message_val:
                            continue

                        row_data = []
                        for i in range(3):
                            if i < len(row['cells']):
                                val = row['cells'][i].get('value', '')
                                row_data.append(str(val) if val else '')
                            else:
                                row_data.append('')

                        lines.append("| " + " | ".join(row_data) + " |")

            lines.append("")
            lines.append("---")
            lines.append("")
            continue

        # ========================================
        # 데이터 전송 및 제어 인터페이스
        # ========================================

        # Row 0: 타이틀
        title = ""
        if len(sheet['rows']) > 0 and sheet['rows'][0]['cells']:
            title = sheet['rows'][0]['cells'][0].get('value', '')

        if title:
            lines.append(f"## {title}")
            lines.append("")

        # Row 2: 헤더 확인
        if len(sheet['rows']) <= 2:
            continue

        header_row = sheet['rows'][2]
        row3 = sheet['rows'][3] if len(sheet['rows']) > 3 else None
        headers = []
        header_indices = []

        # 제어 시트 여부 확인
        is_control_sheet = sheet_name in ['전송요청', '단말제어', '스마트리부팅', 'STB재시작']

        # 모든 헤더 추출 (검토 컬럼은 Android/OCAP 구분)
        max_col = 15
        idx = 0
        while idx < min(max_col, len(header_row['cells'])):
            cell = header_row['cells'][idx]
            val = cell.get('value', '')

            if val:
                # "검토" 헤더를 발견하면 다음 컬럼도 확인 (Android/OCAP 구분)
                if val == '검토' and row3:
                    # 데이터 전송 시트: Col 10-11, 제어 시트: Col 11-12
                    platform_col1 = idx
                    platform_col2 = idx + 1

                    platform1 = row3['cells'][platform_col1].get(
                        'value', '') if len(
                            row3['cells']) > platform_col1 else ''
                    platform2 = row3['cells'][platform_col2].get(
                        'value', '') if len(
                            row3['cells']) > platform_col2 else ''

                    # Android/OCAP 플랫폼 정보가 있으면 분리
                    if platform1 in ['Android', 'Andriod'
                                     ] or platform2 in ['OCAP']:
                        # 첫 번째 검토 컬럼 (Android)
                        if platform1:
                            headers.append(f'검토 ({platform1})')
                            header_indices.append(platform_col1)
                        # 두 번째 검토 컬럼 (OCAP)
                        if platform2:
                            headers.append(f'검토 ({platform2})')
                            header_indices.append(platform_col2)
                        idx += 2  # 두 컬럼을 처리했으므로 건너뜀
                        continue

                headers.append(str(val))
                header_indices.append(idx)

            idx += 1

        if not headers:
            continue

        # Row 3: 연동 정보 및 플랫폼 정보
        if len(sheet['rows']) > 3:
            row3 = sheet['rows'][3]

            # 연동구분 (Col 1)
            protocol = row3['cells'][1].get('value', '') if len(
                row3['cells']) > 1 else ''
            if protocol:
                lines.append(f"**연동구분**: {protocol}")

            # 연동 (Col 2)
            endpoint = row3['cells'][2].get('value', '') if len(
                row3['cells']) > 2 else ''
            if endpoint:
                lines.append(f"**연동**: {endpoint}")

            # 플랫폼 정보 (제어 시트는 Col 11-12, 데이터 전송 시트는 Col 10-11)
            is_control_sheet = sheet_name in [
                '전송요청', '단말제어', '스마트리부팅', 'STB재시작'
            ]
            platform_col1 = 11 if is_control_sheet else 10
            platform_col2 = 12 if is_control_sheet else 11

            platform1 = row3['cells'][platform_col1].get('value', '') if len(
                row3['cells']) > platform_col1 else ''
            platform2 = row3['cells'][platform_col2].get('value', '') if len(
                row3['cells']) > platform_col2 else ''

            platforms = []
            if platform1:
                platforms.append(platform1)
            if platform2:
                platforms.append(platform2)

            if platforms:
                lines.append(f"**플랫폼**: {', '.join(platforms)}")

            lines.append("")

        # Row 4: 전송조건 (데이터 전송 인터페이스만)
        if sheet_name in ['주기전송', '일일전송', '품질계측전송', '자가진단전송', '망품질전환전송']:
            if len(sheet['rows']) > 4:
                row4 = sheet['rows'][4]
                condition = row4['cells'][0].get('value', '') if len(
                    row4['cells']) > 0 else ''
                if condition:
                    lines.append(f"**전송조건**: {condition}")
                    lines.append("")

        # 필드 테이블 헤더
        lines.append("| " + " | ".join(headers) + " |")

        # 테이블 구분선
        alignments = []
        for h in headers:
            if h in ['타입', '필수여부']:
                alignments.append(':----:')
            else:
                alignments.append('------')
        lines.append("|" + "|".join(alignments) + "|")

        # 데이터 행 (Row 4부터)
        start_row = 4

        for row_idx in range(start_row, len(sheet['rows'])):
            row = sheet['rows'][row_idx]
            if not row['cells']:
                continue

            # 필드명 확인 (KEY명1 또는 KEY명2)
            key_col_idx = 2  # KEY명1 (Col 2)
            field_name = row['cells'][key_col_idx].get('value', '') if len(
                row['cells']) > key_col_idx else ''

            # KEY명1이 비어있으면 KEY명2 확인 (제어 시트의 Response 필드용)
            if not field_name and len(row['cells']) > 3:
                key_col_idx = 3  # KEY명2 (Col 3)
                field_name = row['cells'][key_col_idx].get('value', '')

            # 필드명이 없거나 헤더 행이면 스킵
            if not field_name or field_name in ['KEY명1', 'KEY명2']:
                continue

            # 각 헤더에 해당하는 값 추출
            row_data = []
            for idx in header_indices:
                if idx < len(row['cells']):
                    val = row['cells'][idx].get('value', '')
                    if val:
                        # 줄바꿈 처리
                        val = str(val).replace('\n', '<br>')
                        # 필수여부 체크 마크로 변환
                        if '필수여부' in headers and header_indices.index(
                                idx) == headers.index('필수여부'):
                            if str(val).upper() in ['Y', 'O']:
                                val = '✓'
                        # 필드명에 백틱 추가 (KEY명1 또는 KEY명2 컬럼)
                        if ('KEY명1' in headers and header_indices.index(
                                idx) == headers.index('KEY명1')) or \
                           ('KEY명2' in headers and header_indices.index(
                                idx) == headers.index('KEY명2')):
                            val = f'`{val}`'
                        # json 샘플은 코드 블록으로
                        if 'json 샘플' in headers and header_indices.index(
                                idx) == headers.index('json 샘플'):
                            if val and len(val) > 20:
                                val = f'<details><summary>JSON 샘플</summary><pre>{val}</pre></details>'
                    row_data.append(val if val else '')
                else:
                    row_data.append('')

            lines.append("| " + " | ".join(row_data) + " |")

        lines.append("")
        lines.append("---")
        lines.append("")

    # ========================================
    # 문서 끝
    # ========================================
    lines.append("## 문서 생성 정보")
    lines.append("")
    lines.append(f"- 생성 시각: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    lines.append("- 생성 스크립트: `generate_spec.py`")
    lines.append("- 원본 데이터: `interface_data.json`")
    lines.append("")

    return '\n'.join(lines)


if __name__ == '__main__':
    print("INTERFACE.spec.md 생성 중...")

    spec_content = generate_spec()

    output_path = 'analysis/INTERFACE.spec.md'

    with open(output_path, 'w', encoding='utf-8') as f:
        f.write(spec_content)

    print(f"✅ 생성 완료: {output_path}")
    print(f"   파일 크기: {len(spec_content)} bytes")
    print(f"   라인 수: {len(spec_content.splitlines())} lines")
