#!/usr/bin/env python3
"""
Drawio 파일 병합 스크립트

여러 개의 .drawio 파일을 하나의 파일로 병합하되, 각 파일을 별도의 페이지(diagram)로 유지합니다.
"""

import xml.etree.ElementTree as ET
from pathlib import Path
import sys
from typing import List


def parse_drawio_file(file_path: Path) -> List[ET.Element]:
    """
    drawio 파일을 파싱하여 모든 diagram 요소를 추출합니다.
    
    Args:
        file_path: drawio 파일 경로
        
    Returns:
        diagram 요소 리스트
    """
    try:
        tree = ET.parse(file_path)
        root = tree.getroot()
        
        # mxfile 루트에서 모든 diagram 요소 추출
        diagrams = root.findall('diagram')
        
        if not diagrams:
            print(f"⚠️  경고: {file_path.name}에서 diagram 요소를 찾을 수 없습니다.")
            return []
        
        print(f"✅ {file_path.name}: {len(diagrams)}개 diagram 발견")
        return diagrams
        
    except ET.ParseError as e:
        print(f"❌ 오류: {file_path.name} 파싱 실패 - {e}")
        return []
    except Exception as e:
        print(f"❌ 오류: {file_path.name} 처리 중 예외 발생 - {e}")
        return []


def merge_drawio_files(input_files: List[Path], output_file: Path) -> bool:
    """
    여러 drawio 파일을 하나로 병합합니다.
    
    Args:
        input_files: 병합할 drawio 파일 경로 리스트
        output_file: 출력 파일 경로
        
    Returns:
        성공 여부
    """
    if not input_files:
        print("❌ 오류: 병합할 파일이 없습니다.")
        return False
    
    print(f"\n📋 총 {len(input_files)}개 파일 병합 시작...\n")
    
    # 새로운 mxfile 루트 생성
    mxfile = ET.Element('mxfile')
    mxfile.set('host', 'app.diagrams.net')
    mxfile.set('modified', '2026-02-23T00:00:00.000Z')
    mxfile.set('agent', '5.0')
    mxfile.set('version', '24.0.0')
    
    total_diagrams = 0
    
    # 각 파일에서 diagram 추출하여 추가
    for input_file in input_files:
        diagrams = parse_drawio_file(input_file)
        
        for diagram in diagrams:
            # diagram의 ID를 고유하게 만들기 위해 파일명 prefix 추가
            original_id = diagram.get('id', '')
            file_prefix = input_file.stem.replace('.', '_')
            new_id = f"{file_prefix}_{original_id}"
            diagram.set('id', new_id)
            
            # 페이지 이름에 파일명 prefix 추가 (선택사항)
            original_name = diagram.get('name', '')
            if original_name:
                # 파일명에서 번호와 제목 추출
                file_parts = input_file.stem.split('-', 1)
                if len(file_parts) > 1:
                    page_number = file_parts[0]
                    diagram.set('name', f"[{page_number}] {original_name}")
            
            mxfile.append(diagram)
            total_diagrams += 1
    
    if total_diagrams == 0:
        print("❌ 오류: 병합할 diagram이 없습니다.")
        return False
    
    # XML 파일 생성
    tree = ET.ElementTree(mxfile)
    ET.indent(tree, space='  ')  # Python 3.9+
    
    try:
        tree.write(output_file, encoding='utf-8', xml_declaration=True)
        print(f"\n✅ 병합 완료!")
        print(f"📄 출력 파일: {output_file}")
        print(f"📊 총 {total_diagrams}개 페이지 포함\n")
        return True
        
    except Exception as e:
        print(f"❌ 오류: 파일 저장 실패 - {e}")
        return False


def main():
    """메인 함수"""
    # 현재 디렉토리
    current_dir = Path(__file__).parent
    
    # 병합할 파일 목록 (순서 지정)
    input_file_names = [
        '01-control-ui-wireframe.drawio',
        '02.01-all-jobs-list-ui-wireframe.drawio',
        '02.02-job-detail-result-ui-wireframe.drawio',
        '03-control-sequence-diagram.drawio',
        '04-state-transition-diagram.drawio',
        '05-system-architecture-diagram.drawio',
    ]
    
    # 파일 존재 여부 확인
    input_files = []
    missing_files = []
    
    for file_name in input_file_names:
        file_path = current_dir / file_name
        if file_path.exists():
            input_files.append(file_path)
        else:
            missing_files.append(file_name)
    
    if missing_files:
        print("⚠️  경고: 다음 파일을 찾을 수 없습니다:")
        for file_name in missing_files:
            print(f"   - {file_name}")
        print()
    
    if not input_files:
        print("❌ 오류: 병합할 파일이 하나도 없습니다.")
        sys.exit(1)
    
    # 출력 파일명
    output_file = current_dir / 'catv-stb-control-all-diagrams.drawio'
    
    # 병합 실행
    success = merge_drawio_files(input_files, output_file)
    
    if success:
        print("사용 방법:")
        print("1. draw.io 또는 diagrams.net에서 파일 열기")
        print("2. 하단의 탭에서 각 페이지 확인")
        print("3. '+' 버튼으로 새 페이지 추가 가능\n")
        sys.exit(0)
    else:
        sys.exit(1)


if __name__ == '__main__':
    main()
