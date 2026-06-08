import react from '@vitejs/plugin-react'
import path from 'path'
import { fileURLToPath } from 'url'
import { defineConfig } from 'vite'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const apiProxyTarget = process.env.VITE_API_PROXY_TARGET || process.env.VITE_API_BASE_URL || 'http://localhost:3001'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      // Deixamos o /api padrão que você já tinha
      '/api': {
        target: apiProxyTarget,
        changeOrigin: true,
        secure: false,
      },
      // Se o seu formulário faz fetch para '/auth/register', o Vite vai mandar direto
      // para o NestJS. Se o Nest usar o prefixo global '/api', o rewrite adiciona automaticamente.
      '/auth': {
        target: apiProxyTarget,
        changeOrigin: true,
        secure: false,
        configure: (proxy) => {
          proxy.on('error', (err) => {
            console.error('Erro no proxy do Vite para /auth:', err.message);
          });
        },
        // Se o seu NestJS usa setGlobalPrefix('api'), descomente a linha abaixo:
        // rewrite: (path) => path.replace(/^\/auth/, '/api/auth'),
      },
    },
  },
})
