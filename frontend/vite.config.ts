import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import svgr from 'vite-plugin-svgr';

// The Go master embeds the build from endpoints/web/data and serves /static/*
// plus the files from public/ at the root, so keep that layout.
// BUILD_PATH=../endpoints/web/data/ yarn build writes it there directly.
const backend = process.env.HUB_BACKEND ?? 'http://localhost:8002';

export default defineConfig({
    plugins: [react(), svgr()],
    server: {
        port: 3000,
        proxy: {
            '/api': { target: backend, changeOrigin: true },
            '/auth': { target: backend, changeOrigin: true },
            '/upload': { target: backend, changeOrigin: true },
            '/data': { target: backend, changeOrigin: true },
        },
    },
    build: {
        outDir: process.env.BUILD_PATH ?? 'build',
        emptyOutDir: true,
        assetsDir: 'static',
        chunkSizeWarningLimit: 4000,
    },
});
