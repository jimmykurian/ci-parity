// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// The docs site, served from GitHub Pages at https://jimmykurian.github.io/ci-parity/.
// Its pages come from docs/ through scripts/sync-docs.mjs; only the home page lives here.

import starlight from '@astrojs/starlight';
import { defineConfig } from 'astro/config';
import mermaid from 'astro-mermaid';
import starlightLinksValidator from 'starlight-links-validator';

export default defineConfig({
  site: 'https://jimmykurian.github.io',
  base: '/ci-parity',
  integrations: [
    // Renders the ```mermaid diagrams in the browser, following the light or dark theme.
    // It must come before Starlight.
    mermaid({ theme: 'default', autoTheme: true }),
    starlight({
      title: 'ci-parity',
      description:
        'Lint GitHub Actions workflows so CI runs project commands through the task runner, with a baseline that only shrinks.',
      // The logo includes the wordmark, so it stands in for the title text.
      logo: { light: './src/assets/logo-light.svg', dark: './src/assets/logo-dark.svg', replacesTitle: true },
      social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/jimmykurian/ci-parity' }],
      // Each page ends with its own "Last reviewed" date.
      lastUpdated: false,
      sidebar: [
        { label: 'Overview', link: '/overview/' },
        {
          label: 'Guides',
          items: [
            'guides',
            'guides/getting-started',
            'guides/configuring',
            'guides/baseline',
            'guides/fixing',
            'guides/ci',
            'guides/recipes',
            'guides/why',
            'guides/faq',
          ],
        },
        { label: 'Rules', items: [{ autogenerate: { directory: 'rules' } }] },
        { label: 'Reference', items: [{ autogenerate: { directory: 'reference' } }] },
        { label: 'Architecture', items: [{ autogenerate: { directory: 'architecture' } }] },
        { label: 'Releases and CI', items: [{ autogenerate: { directory: 'deployment' } }] },
        { label: 'Testing', items: [{ autogenerate: { directory: 'testing' } }] },
        { label: 'Security', items: [{ autogenerate: { directory: 'security' } }] },
        { label: 'Style guide', link: '/style_guide/' },
        { label: 'Changelog', link: '/changelog/' },
      ],
      // Fails the build on a broken link or anchor between pages.
      plugins: [starlightLinksValidator()],
    }),
  ],
});
