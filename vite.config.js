import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

// https://vite.dev/config/
export default defineConfig({
	plugins: [svelte(), tailwindcss()],
	base: '/chad/',
	resolve: {
		conditions: ['browser', 'svelte']
	},
	test: {
		environment: 'jsdom',
		globals: true
	}
});
