// Mock for uuid to resolve ESM-only package issue in Jest CJS mode.
// uuid v9+ ships as ESM-only. Node.js 19+ crypto.randomUUID() is available in
// both Node and jsdom environments, so we delegate to it for correctness.
export const v4 = () => crypto.randomUUID();
export const v1 = v4;
export const v3 = v4;
export const v5 = v4;
