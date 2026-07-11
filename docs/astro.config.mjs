// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import config from './config.mjs';
import sidebar from './sidebar.mjs';

// https://astro.build/config
export default defineConfig({
	site: config.url,
	base: config.basePath,
	integrations: [
		starlight({
			title: config.title,
			social: [
				{ icon: 'github', label: 'GitHub', href: config.github },
			],
			customCss: [
				'./src/styles/custom.css',
			],
			sidebar: sidebar,
		}),
	],
});
