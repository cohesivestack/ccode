import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightVersions from 'starlight-versions';
import tailwindcss from '@tailwindcss/vite';
import { docsSidebar } from './src/data/sidebar.mjs';

export default defineConfig({
  redirects: {
    '/0.1': '/0.1/getting-started/',
  },
  integrations: [
    starlight({
      title: 'Cohesive Code',
      description: 'AI-enabled code generation built around TypeScript processes, templates, and accelerators.',
      logo: {
        light: './public/cohesive-code-logo.svg',
        dark: './public/cohesive-code-logo-dark.svg',
        alt: 'Cohesive Code',
        replacesTitle: true,
      },
      favicon: '/favicon.png',
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/cohesivestack/ccode',
        },
      ],
      customCss: [
        "./src/styles/global.css",
      ],
      plugins: [
        starlightVersions({
          versions: [{ slug: '0.1', label: 'v0.1' }],
          current: { label: 'Latest' },
        }),
      ],
      sidebar: docsSidebar,
    }),
  ],

  vite: {
    plugins: [tailwindcss()],
  },
});
