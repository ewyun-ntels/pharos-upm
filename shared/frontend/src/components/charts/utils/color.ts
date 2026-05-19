/* tailwind4: oklch를 리챠트에서 지원 하지 않을 수 있어서 아래 getCSSColor 방법으로 수정함 */
const getCSSColor = (variable: string) =>
  getComputedStyle(document.documentElement).getPropertyValue(variable).trim();

/**
 * SVG stroke/fill 속성은 oklch 등 최신 CSS 색상 포맷을 지원하지 않으므로
 * Canvas API로 브라우저가 계산한 rgb() 값으로 변환합니다.
 */
function resolveColorToRgb(cssValue: string): string {
  if (!cssValue) return '#3b82f6';
  // 이미 SVG가 이해하는 포맷이면 그대로 사용
  if (/^#/.test(cssValue) || /^rgba?\(/.test(cssValue) || /^hsla?\(/.test(cssValue)) {
    return cssValue;
  }
  try {
    const canvas = document.createElement('canvas');
    canvas.width = 1;
    canvas.height = 1;
    const ctx = canvas.getContext('2d');
    if (!ctx) return '#3b82f6';
    ctx.fillStyle = cssValue;
    ctx.fillRect(0, 0, 1, 1);
    const [r, g, b, a] = ctx.getImageData(0, 0, 1, 1).data;
    if (a === 0) return '#3b82f6';
    return `rgb(${r},${g},${b})`;
  } catch {
    return '#3b82f6';
  }
}

/* 기본 차트 색상 팔레트 — oklch → rgb 변환 후 Recharts SVG에서 사용 가능
 * 모듈 로드 시점에는 CSS 변수가 아직 적용 안 될 수 있으므로 첫 사용 시 초기화 */
let _colorPalette: string[] | null = null;

function getColorPalette(): string[] {
  if (!_colorPalette) {
    _colorPalette = Array.from({ length: 17 }, (_, i) =>
      resolveColorToRgb(getCSSColor(`--chart-${i + 1}`))
    );
  }
  return _colorPalette;
}

interface AliasColor {
  legendName: string;
  color: string;
}

interface AliasColors {
  defaultColor?: string;
  colors: AliasColor[];
}

const GetColor = (
  index: number,
  legendName?: string,
  aliasColors?: AliasColors | undefined,
  useStatusPalette?: string[] | boolean,
) => {
  const palette = Array.isArray(useStatusPalette)
    ? useStatusPalette
    : getColorPalette();

  const color = aliasColors?.colors.find(
    (aliasColor) => aliasColor.legendName === legendName,
  )?.color;
  if (color) {
    return color;
  } else if (aliasColors?.defaultColor) {
    return aliasColors.defaultColor;
  }

  return palette[index % palette.length];
};

export {GetColor};
export type {AliasColor, AliasColors};
