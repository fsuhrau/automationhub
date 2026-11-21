import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import svgr from 'vite-plugin-svgr';

// The Go master embeds the build from endpoints/web/data and serves /static/*
// plus the files from public/ at the root, so keep that layout.
// BUILD_PATH=../endpoints/web/data/ yarn build writes it there directly.
const backend = process.env.HUB_BACKEND ?? 'http://localhost:8002';

export default defineConfig({
    plugins: [react(), svgr({
        svgrOptions: {
            // the platform icons hard-code black; let them follow the text color
            replaceAttrValues: {'#000000': 'currentColor', '#000': 'currentColor'},
        },
    })],
    // pre-bundle the shared UI libraries so the first dev start doesn't
    // re-optimize mid-load (which briefly loads @emotion/react twice)
    optimizeDeps: {
        include: ['@emotion/react', '@emotion/styled', '@mui/material', '@mui/icons-material', '@mui/x-data-grid', '@mui/x-tree-view'],
    },
    resolve: {
        dedupe: ['react', 'react-dom', '@emotion/react', '@emotion/styled'],
    },
    server: {
        port: 3000,
        proxy: {
            // ws: the device video stream is a websocket under /api
            '/api': { target: backend, changeOrigin: true, ws: true },
            '/auth': { target: backend, changeOrigin: true },
            '/upload': { target: backend, changeOrigin: true },
            '/data': { target: backend, changeOrigin: true },
        },
    },
    test: {
        environment: 'jsdom',
        setupFiles: ['src/setupTests.ts'],
        coverage: {
            provider: 'v8',
            reporter: ['text', 'json-summary'],
            include: ['src/**/*.{ts,tsx}'],
            exclude: ['src/**/*.test.{ts,tsx}', 'src/testing/**', 'src/setupTests.ts', 'src/vite-env.d.ts'],
        },
    },
    build: {
        outDir: process.env.BUILD_PATH ?? 'build',
        emptyOutDir: true,
        assetsDir: 'static',
        chunkSizeWarningLimit: 4000,
    },
});
