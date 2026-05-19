// Mock for query-string (ESM-only package) — CJS-compatible for ts-jest v29 + Jest v30.
// ts-jest v30 (unreleased) with native ESM support will allow removing this mock.
const parse = (str) => {
  if (!str) return {};
  const params = new URLSearchParams(str);
  const result = {};
  for (const [key, value] of params) {
    result[key] = value;
  }
  return result;
};

const stringify = (obj) => {
  if (!obj) return '';
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(obj)) {
    if (value !== undefined && value !== null) {
      params.append(key, String(value));
    }
  }
  return params.toString();
};

const parseUrl = (url) => {
  if (!url) return { url: '', query: {} };
  const [baseUrl, queryString] = url.split('?');
  return { url: baseUrl || '', query: parse(queryString || '') };
};

const stringifyUrl = (obj) => {
  if (!obj || !obj.url) return '';
  const query = stringify(obj.query || {});
  return query ? `${obj.url}?${query}` : obj.url;
};

module.exports = { parse, stringify, parseUrl, stringifyUrl };
module.exports.default = module.exports;