// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Copies docs/ into the Starlight content collection, so docs/ stays the single source
// and keeps rendering on GitHub. Each page's H1 becomes its title, its first paragraph
// its description, and its links are rewritten: links between docs pages become site
// routes, links out of docs/ point at the file on GitHub. Run by `task site:build`.

import { cpSync, existsSync, mkdirSync, readdirSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { dirname, join, posix, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

/** The repository on GitHub, for links that leave docs/ and for "Edit page". */
export const REPO = 'https://github.com/jimmykurian/ci-parity';

/** The path the site is served under on GitHub Pages. */
export const BASE = '/ci-parity';

/**
 * Returns where a docs/ page lands in the content collection: README.md becomes its
 * directory's index, and the top-level README becomes overview, since / is the home page.
 * @param {string} rel - path relative to docs/, with forward slashes
 */
export function outputPath(rel) {
  if (rel === 'README.md') return 'overview.md';
  if (posix.basename(rel) === 'README.md') return posix.join(posix.dirname(rel), 'index.md');
  return rel;
}

/**
 * Returns the site route for a docs/ page, under BASE and ending in a slash. Astro
 * lowercases each segment, so STYLE_GUIDE.md is served at style_guide/.
 * @param {string} rel - path relative to docs/, with forward slashes
 */
export function route(rel) {
  const out = outputPath(rel).replace(/\.md$/, '').toLowerCase();
  const path = out === 'index' ? '' : out.replace(/(^|\/)index$/, '');
  return `${BASE}/${path ? `${path}/` : ''}`;
}

/**
 * Rewrites one link target found in the page at rel.
 * @param {string} target - the link as written, such as `rules.md#p001` or `../../LICENSE`
 * @param {string} rel - path of the page relative to docs/
 * @param {(repoPath: string) => boolean} isDir - whether a repository path is a directory
 */
export function rewriteLink(target, rel, isDir) {
  if (/^[a-z][a-z0-9+.-]*:/i.test(target) || target.startsWith('#')) return target;
  const [path, anchor] = target.split('#', 2);
  const suffix = anchor === undefined ? '' : `#${anchor}`;
  const repoPath = posix.normalize(posix.join('docs', posix.dirname(rel), path)).replace(/\/$/, '');
  if (repoPath.startsWith('docs/assets/')) return `${BASE}/${posix.basename(repoPath)}${suffix}`;
  if (repoPath.startsWith('docs/') && repoPath.endsWith('.md')) {
    return route(repoPath.slice('docs/'.length)) + suffix;
  }
  const kind = isDir(repoPath) ? 'tree' : 'blob';
  return `${REPO}/${kind}/main/${repoPath}${suffix}`;
}

/**
 * Converts one docs/ page into a Starlight page: frontmatter from its H1 and first
 * paragraph, the license header and H1 removed, and links rewritten. Fenced code is
 * left alone.
 * @param {string} source - the Markdown as it is in docs/
 * @param {string} rel - path relative to docs/
 * @param {(repoPath: string) => boolean} isDir - whether a repository path is a directory
 */
export function convert(source, rel, isDir) {
  const lines = source.replace(/\r\n/g, '\n').split('\n');
  let i = 0;
  if (lines[i]?.startsWith('<!--') && lines[i].includes('SPDX-License-Identifier')) i++;
  while (lines[i] === '') i++;
  if (!lines[i]?.startsWith('# ')) throw new Error(`${rel}: no H1 after the license header`);
  const title = lines[i].slice(2).trim();
  i++;
  while (lines[i] === '') i++;
  const paragraph = [];
  for (let j = i; j < lines.length && lines[j] !== '' && !/^[#|`>-]/.test(lines[j]); j++) {
    paragraph.push(lines[j].trim());
  }
  const description = paragraph
    .join(' ')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .trim();

  let fenced = false;
  const body = lines.slice(i).map((line) => {
    if (/^\s*```/.test(line)) fenced = !fenced;
    if (fenced || /^\s*```/.test(line)) return line;
    return line.replace(/\]\(([^)\s]+)\)/g, (_, target) => `](${rewriteLink(target, rel, isDir)})`);
  });

  const frontmatter = ['---', `title: ${JSON.stringify(title)}`];
  if (description) frontmatter.push(`description: ${JSON.stringify(description)}`);
  frontmatter.push(`editUrl: ${JSON.stringify(`${REPO}/edit/main/docs/${rel}`)}`);
  // A section's README is its overview, so it comes first in the sidebar group.
  if (posix.basename(rel) === 'README.md') frontmatter.push('sidebar:', '  order: 0');
  frontmatter.push('---', '');
  return frontmatter.join('\n') + body.join('\n');
}

/**
 * Converts CHANGELOG.md, which Release Please writes, into the changelog page: its H1
 * becomes the title, and the page has no "Edit page" link, since edits would be
 * overwritten. Its links are already absolute.
 * @param {string} source - CHANGELOG.md as Release Please writes it
 */
export function changelog(source) {
  const body = source.replace(/\r\n/g, '\n').replace(/^# [^\n]*\n+/, '');
  return [
    '---',
    'title: "Changelog"',
    'description: "Every ci-parity release, written by Release Please from the commits."',
    'editUrl: false',
    '---',
    '',
    body,
  ].join('\n');
}

/** Lists every Markdown file under dir, as forward-slash paths relative to it. */
function markdownFiles(dir, prefix = '') {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const rel = prefix ? `${prefix}/${entry.name}` : entry.name;
    if (entry.isDirectory()) return markdownFiles(join(dir, entry.name), rel);
    return entry.name.endsWith('.md') ? [rel] : [];
  });
}

/** Syncs docs/ into the site, replacing whatever an earlier sync wrote. */
function main() {
  const site = resolve(dirname(fileURLToPath(import.meta.url)), '..');
  const repo = resolve(site, '..');
  const docs = join(repo, 'docs');
  const content = join(site, 'src', 'content', 'docs');
  const isDir = (repoPath) => existsSync(join(repo, repoPath)) && statSync(join(repo, repoPath)).isDirectory();

  for (const entry of readdirSync(content)) {
    if (entry !== 'index.mdx') rmSync(join(content, entry), { recursive: true, force: true });
  }
  const files = markdownFiles(docs);
  for (const rel of files) {
    const target = join(content, outputPath(rel));
    mkdirSync(dirname(target), { recursive: true });
    writeFileSync(target, convert(readFileSync(join(docs, rel), 'utf8'), rel, isDir));
  }

  // The logo and images: the logo for Starlight's header, every asset for links to it.
  for (const dir of [join(site, 'src', 'assets'), join(site, 'public')]) {
    rmSync(dir, { recursive: true, force: true });
    cpSync(join(docs, 'assets'), dir, { recursive: true });
  }
  writeFileSync(join(content, 'changelog.md'), changelog(readFileSync(join(repo, 'CHANGELOG.md'), 'utf8')));
  console.log(`synced ${files.length} pages from ${relative(site, docs)}/, and the changelog`);
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) main();
