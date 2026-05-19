// Mock for parse-duration library to avoid ESM issues in Jest

const parse = (duration: string): number | null => {
  if (!duration) {
    return null;
  }

  const match = duration.match(/^(\d+(?:\.\d+)?)\s*(\w+)$/);
  if (!match) {
    return null;
  }

  const value = parseFloat(match[1]);
  const unit = match[2].toLowerCase();

  const units: Record<string, number> = {
    ms: 1,
    millisecond: 1,
    milliseconds: 1,
    s: 1000,
    sec: 1000,
    second: 1000,
    seconds: 1000,
    m: 60 * 1000,
    min: 60 * 1000,
    minute: 60 * 1000,
    minutes: 60 * 1000,
    h: 60 * 60 * 1000,
    hour: 60 * 60 * 1000,
    hours: 60 * 60 * 1000,
    d: 24 * 60 * 60 * 1000,
    day: 24 * 60 * 60 * 1000,
    days: 24 * 60 * 60 * 1000,
    w: 7 * 24 * 60 * 60 * 1000,
    week: 7 * 24 * 60 * 60 * 1000,
    weeks: 7 * 24 * 60 * 60 * 1000,
    month: 30 * 24 * 60 * 60 * 1000,
    months: 30 * 24 * 60 * 60 * 1000,
    y: 365 * 24 * 60 * 60 * 1000,
    year: 365 * 24 * 60 * 60 * 1000,
    years: 365 * 24 * 60 * 60 * 1000,
  };

  const multiplier = units[unit];
  return multiplier !== undefined ? value * multiplier : null;
};

export default parse;
