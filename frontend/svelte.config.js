import adapterAuto from '@sveltejs/adapter-auto';
import adapterStatic from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

// Determine which adapter to use based on environment
const isElectronBuild = process.env.ELECTRON_BUILD === 'true';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),

	kit: {
		// Use adapter-static for Electron builds, adapter-auto for web deployment
		adapter: isElectronBuild
			? adapterStatic({
					pages: 'build',
					assets: 'build',
					fallback: 'index.html', // SPA fallback for client-side routing
					precompress: false,
					strict: false
				})
			: adapterAuto()
	}
};

export default config;
