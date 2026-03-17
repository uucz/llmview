import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		proxy: {
			'/api': 'http://localhost:4700',
			'/ws': { target: 'ws://localhost:4700', ws: true },
			'/proxy': 'http://localhost:4700'
		}
	}
});
