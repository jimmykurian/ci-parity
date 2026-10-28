// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

import assert from 'node:assert/strict';
import { test } from 'node:test';

import { changelog, convert, outputPath, rewriteLink, route } from './sync-docs.mjs';

const noDirs = () => false;

test('outputPath maps READMEs to indexes and the top-level README to overview', () => {
  // Arrange
  const cases = [
    ['README.md', 'overview.md'],
    ['reference/README.md', 'reference/index.md'],
    ['reference/rules.md', 'reference/rules.md'],
  ];

  for (const [rel, want] of cases) {
    // Act
    const got = outputPath(rel);

    // Assert
    assert.equal(got, want, rel);
  }
});

test('route serves each page under the base with a trailing slash', () => {
  // Arrange
  const cases = [
    ['README.md', '/ci-parity/overview/'],
    ['reference/README.md', '/ci-parity/reference/'],
    ['testing/testing-strategy.md', '/ci-parity/testing/testing-strategy/'],
    ['STYLE_GUIDE.md', '/ci-parity/style_guide/'],
  ];

  for (const [rel, want] of cases) {
    // Act
    const got = route(rel);

    // Assert
    assert.equal(got, want, rel);
  }
});

test('rewriteLink turns links between pages into routes and keeps anchors', () => {
  // Arrange
  const rel = 'reference/rules.md';

  // Act
  const got = [
    rewriteLink('analysis.md#invokers', rel, noDirs),
    rewriteLink('../architecture/README.md', rel, noDirs),
    rewriteLink('#suggestions', rel, noDirs),
  ];

  // Assert
  assert.deepEqual(got, ['/ci-parity/reference/analysis/#invokers', '/ci-parity/architecture/', '#suggestions']);
});

test('rewriteLink points links out of docs/ at GitHub, as a tree for directories', () => {
  // Arrange
  const isDir = (repoPath) => repoPath === '.github';

  // Act
  const got = [
    rewriteLink('../../CONTRIBUTING.md', 'deployment/releases-and-ci.md', isDir),
    rewriteLink('../.github/', 'README.md', isDir),
    rewriteLink('https://example.com/x.md', 'README.md', isDir),
  ];

  // Assert
  assert.deepEqual(got, [
    'https://github.com/jimmykurian/ci-parity/blob/main/CONTRIBUTING.md',
    'https://github.com/jimmykurian/ci-parity/tree/main/.github',
    'https://example.com/x.md',
  ]);
});

test('rewriteLink serves docs/assets files from the site root', () => {
  // Act
  const got = rewriteLink('../assets/logo.svg', 'reference/README.md', noDirs);

  // Assert
  assert.equal(got, '/ci-parity/logo.svg');
});

test('convert writes frontmatter from the H1 and first paragraph and drops the header', () => {
  // Arrange
  const source = [
    '<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->',
    '',
    '# Rules',
    '',
    'Rules P001 to P005, with [their messages](cli.md).',
    '',
    '## P001',
    '',
    'See [analysis](analysis.md#commands).',
  ].join('\n');

  // Act
  const got = convert(source, 'reference/rules.md', noDirs);

  // Assert
  assert.equal(
    got,
    [
      '---',
      'title: "Rules"',
      'description: "Rules P001 to P005, with their messages."',
      'editUrl: "https://github.com/jimmykurian/ci-parity/edit/main/docs/reference/rules.md"',
      '---',
      'Rules P001 to P005, with [their messages](/ci-parity/reference/cli/).',
      '',
      '## P001',
      '',
      'See [analysis](/ci-parity/reference/analysis/#commands).',
    ].join('\n'),
  );
});

test('convert puts a section README first in its sidebar group', () => {
  // Arrange
  const source = '# Reference\n\nThe contract.\n';

  // Act
  const got = convert(source, 'reference/README.md', noDirs);

  // Assert
  assert.match(got, /\nsidebar:\n {2}order: 0\n---\n/);
});

test('convert leaves fenced code untouched', () => {
  // Arrange
  const source = ['# Title', '', 'Intro.', '', '```markdown', '- [Doc](../x.md)', '```'].join('\n');

  // Act
  const got = convert(source, 'README.md', noDirs);

  // Assert
  assert.ok(got.endsWith('```markdown\n- [Doc](../x.md)\n```'), got);
});

test('convert rejects a page with no H1', () => {
  // Act and Assert
  assert.throws(() => convert('Intro only.\n', 'README.md', noDirs), /no H1/);
});

test('changelog drops the H1 and turns off the edit link', () => {
  // Arrange
  const source = '# Changelog\n\n## [0.6.0](https://example.com) (2026-10-28)\n\n* a change\n';

  // Act
  const got = changelog(source);

  // Assert
  assert.ok(got.startsWith('---\ntitle: "Changelog"\n'), got);
  assert.match(got, /\neditUrl: false\n---\n/);
  assert.ok(got.endsWith('## [0.6.0](https://example.com) (2026-10-28)\n\n* a change\n'), got);
  assert.ok(!got.includes('# Changelog'), got);
});
