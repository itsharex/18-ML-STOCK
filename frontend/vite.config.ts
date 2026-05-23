import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

// 唯一版本号来源：../wails.json -> info.productVersion
// 通过 Vite define 在编译期注入为 __APP_VERSION__，避免 Settings.tsx 与 wails.json 双写
const wailsJsonPath = resolve(__dirname, '../wails.json')
const wailsJson = JSON.parse(readFileSync(wailsJsonPath, 'utf-8')) as {
  info?: { productVersion?: string }
}
const appVersion = wailsJson.info?.productVersion
if (!appVersion) {
  throw new Error(`无法从 ${wailsJsonPath} 读取 info.productVersion`)
}

export default defineConfig({
  plugins: [react()],
  define: {
    __APP_VERSION__: JSON.stringify(appVersion),
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
