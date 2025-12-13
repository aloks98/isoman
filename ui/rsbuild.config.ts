import {defineConfig} from '@rsbuild/core';
import {pluginReact} from '@rsbuild/plugin-react';
import * as path from 'node:path';

// Docs: https://rsbuild.rs/config/
export default defineConfig({
    plugins: [
        pluginReact(),
    ],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
        },
    },
    server: {
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
            '/ws': {
                target: 'http://localhost:8080',
                ws: true,
                changeOrigin: true,
            },
            '/images': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
        },
    },
});
