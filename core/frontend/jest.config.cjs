module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'jsdom',
  roots: ['<rootDir>/src'],
  testMatch: ['**/*.test.ts', '**/*.test.tsx'],
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@components/(.*)$': '<rootDir>/src/components/$1',
    '^@features/(.*)$': '<rootDir>/src/features/$1',
    '^@pages/(.*)$': '<rootDir>/pages/$1',
    '^@hooks/(.*)$': '<rootDir>/src/hooks/$1',
    '^@lib/(.*)$': '<rootDir>/src/lib/$1',
    '^@providers/(.*)$': '<rootDir>/src/providers/$1',
    '^@config/(.*)$': '<rootDir>/src/config/$1',
    '^@utils/(.*)$': '<rootDir>/src/utils/$1',
    '^@types$': '<rootDir>/src/types/index',
    '^@shared/(.*)$': '<rootDir>/../../shared/$1',
    '^@pharos/shared/schema$': '<rootDir>/../../shared/schema/src/index.ts',
    '^@pharos/shared/types/(.*)$': '<rootDir>/../../shared/frontend/src/types/$1/index.ts',
    '^@pharos/shared/hooks/(.*)$': '<rootDir>/../../shared/frontend/src/hooks/$1',
    'query-string': '<rootDir>/__mocks__/query-string.js',
    'parse-duration': '<rootDir>/__mocks__/parse-duration.ts',
    '^uuid$': '<rootDir>/__mocks__/uuid.ts',
  },
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', {
      tsconfig: {
        jsx: 'react-jsx',
        esModuleInterop: true,
        allowSyntheticDefaultImports: true,
        rootDir: '.',
        ignoreDeprecations: '6.0',
      },
      useESM: false,
    }],
  },
  transformIgnorePatterns: [
    'node_modules/(?!(query-string|decode-uri-component|split-on-first|filter-obj|lucide-react|parse-duration))'
  ],
};
