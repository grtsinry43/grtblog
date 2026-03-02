import { fileURLToPath, URL } from 'node:url'

import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import { defineConfig, loadEnv } from 'vite'
import vueDevTools from 'vite-plugin-vue-devtools'
import pkg from './package.json'

// https://vite.dev/config/
export default defineConfig((env) => {
  const envVars = loadEnv(env.mode, process.cwd(), '')
  const apiProxyTarget = envVars.VITE_API_PROXY_TARGET || 'http://localhost:8080'
  const appBase = envVars.VITE_APP_BASE || '/'

  return {
    base: appBase,
    define: {
      __APP_VERSION__: JSON.stringify(pkg.version),
    },
    plugins: [vue(), vueJsx(), tailwindcss(), vueDevTools()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        '@shared': fileURLToPath(new URL('../shared', import.meta.url)),
      },
    },
    server: {
      port: 5799,
      host: true,
      fs: {
        allow: [fileURLToPath(new URL('..', import.meta.url))],
      },
      proxy: {
        '/api': {
          target: apiProxyTarget,
          changeOrigin: true,
          ws: true,
        },
        '/uploads': {
          target: apiProxyTarget,
          changeOrigin: true,
        },
      },
    },
    build: {
      rollupOptions: {
        output: {
          advancedChunks: {
            groups: [
              {
                name: 'echarts',
                test: /\/echarts/,
              },
              {
                name: 'chroma-js',
                test: /\/chroma-js/,
              },
              {
                name: 'lodash-es',
                test: /\/lodash-es/,
              },
              {
                name: 'naive-ui',
                test: /\/naive-ui/,
              },
              {
                name: 'vue-draggable-plus',
                test: /\/vue-draggable-plus/,
              },
              {
                name: 'vueuse',
                test: /\/vueuse/,
              },
              {
                name: 'vue',
                test: /\/vue/,
              },
              {
                name: 'vue-router',
                test: /\/vue-router/,
              },
              {
                name: 'pinia',
                test: /\/pinia/,
              },
            ],
          },

          assetFileNames: (info) => {
            const notHash = ['topography.svg', 'texture.png', 'noise.png']
            if (notHash.includes(info.names[0])) {
              return 'assets/[name][extname]'
            }
            return 'assets/[name]-[hash][extname]'
          },
        },
      },
    },
    esbuild: {
      drop: env.mode === 'production' ? ['console', 'debugger'] : [],
    },
  }
})
