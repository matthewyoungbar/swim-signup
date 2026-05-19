import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/practices': 'http://localhost:8080',
      '/signups':   'http://localhost:8080',
      '/my-signups':'http://localhost:8080',
    }
  }
})
