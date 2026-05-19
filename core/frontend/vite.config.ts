import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'
import fs from 'fs'

// ─── Load generated extension metadata ────────────────────────────────────────

interface ViteExtensionEntry {
  name: string;
  path: string;
}

interface ViteExtensions {
  extensions: ViteExtensionEntry[];
}

const viteExtensionsPath = path.resolve(__dirname, '../../meta/generated/frontend/vite-extensions.json');
const viteExtensions: ViteExtensions = (() => {
  if (!fs.existsSync(viteExtensionsPath)) {
    return { extensions: [] };
  }
  try {
    const parsed = JSON.parse(fs.readFileSync(viteExtensionsPath, 'utf-8'));
    if (parsed && Array.isArray(parsed.extensions)) {
      return { extensions: parsed.extensions };
    }
    return { extensions: [] };
  } catch {
    return { extensions: [] };
  }
})();

const extensionAliases = Object.fromEntries(
  viteExtensions.extensions.map(ext => [ext.name, path.resolve(__dirname, ext.path)])
);

const extensionExcludes = viteExtensions.extensions.map(ext => ext.name);

// ─── Config ───────────────────────────────────────────────────────────────────

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
  ],
  base: '/ui/',
  build: {
    outDir: 'out',
    emptyOutDir: true,
    sourcemap: false,
    chunkSizeWarningLimit: 1000,
    rolldownOptions: {
      output: {
        manualChunks(id) {
          if (
            id.includes('ace-builds/src-noconflict/mode-sql') ||
            id.includes('ace-builds/src-noconflict/theme-github') ||
            id.includes('ace-builds/src-noconflict/theme-tomorrow_night') ||
            id.includes('ace-builds/src-noconflict/ext-language_tools')
          ) {
            return 'ace-editor';
          }
        }
      }
    }
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@config': path.resolve(__dirname, './src/config'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@components': path.resolve(__dirname, './src/components'),
      '@features': path.resolve(__dirname, './src/features'),
      '@providers': path.resolve(__dirname, './src/providers'),
      '@styles': path.resolve(__dirname, './src/styles'),
      '@lib': path.resolve(__dirname, './src/lib'),
      '@utils': path.resolve(__dirname, './src/utils'),
      '@routes': path.resolve(__dirname, './src/routes'),
      '@shared/frontend': path.resolve(__dirname, '../../shared/frontend/src'),
      '@pharos/shared/extension-registry': path.resolve(__dirname, '../../shared/frontend/src/extension-registry.ts'),
      '@pharos/shared/components': path.resolve(__dirname, '../../shared/frontend/src/components'),
      '@pharos/shared/components/ui': path.resolve(__dirname, '../../shared/frontend/src/components/ui'),
      '@pharos/shared/features/auth': path.resolve(__dirname, '../../shared/frontend/src/features/auth'),
      '@pharos/shared/hooks/use-auth': path.resolve(__dirname, '../../shared/frontend/src/features/auth'),
      '@pharos/shared/hooks': path.resolve(__dirname, '../../shared/frontend/src/hooks'),
      '@pharos/shared/lib': path.resolve(__dirname, '../../shared/frontend/src/lib'),
      '@pharos/shared/types': path.resolve(__dirname, '../../shared/frontend/src/types'),
      '@pharos/shared/styles': path.resolve(__dirname, '../../shared/frontend/src/styles'),
      '@pharos/shared': path.resolve(__dirname, '../../shared/frontend/src'),
      '@pharos/core/datasource-editor-registry': path.resolve(__dirname, './src/features/dashboard/datasources'),
      '@pharos/core/panel-registry': path.resolve(__dirname, './src/features/dashboard/panels/registry'),
      // Meta generated files
      '@pharos/meta/extension-loader': path.resolve(__dirname, '../../meta/generated/frontend/extension-loader.ts'),
      '@pharos/meta/userJSONSchema': path.resolve(__dirname, '../../meta/generated/frontend/userJSONSchema.ts'),
      '@pharos/meta/menu-order': path.resolve(__dirname, '../../meta/generated/frontend/menu-order.ts'),
      // Extension aliases (from meta/generated/frontend/vite-extensions.json)
      ...extensionAliases,
    },
  },
  // workspace 패키지는 path alias로 소스 직접 참조 → pre-bundle 대상에서 제외
  // (제외하지 않으면 소스 변경 시 chunk 해시 불일치로 캐시 오류 발생)
  optimizeDeps: {
    include: [
      'react-ace',
    ],
    exclude: [
      '@pharos/shared',
      // Extension excludes (from meta/generated/frontend/vite-extensions.json)
      ...extensionExcludes,
    ],
  },
  server: {
    port: 3000,
    proxy: {
      '/sql': {
        target: 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/auth': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/master': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/ui-config': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/dashboard': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/folders': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/proxy/clickhouse': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/proxy/postgresql': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/notification': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/role': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/user': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/me': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/plugins': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/alert': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/badges': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/catv': {
        target: process.env.API_SERVER_URL || 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/alarm': {
        target: process.env.ALARM_SERVER_URL || process.env.API_SERVER_URL || 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/api': {
        target: 'http://192.168.15.102:31000',
        changeOrigin: true,
        secure: false,
      },
      '/ws': {
        target: 'ws://192.168.15.102:31000',
        ws: true,
      },
    },
  },
})
