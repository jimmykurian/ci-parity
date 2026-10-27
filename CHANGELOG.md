# Changelog

## 0.1.0 (2026-10-27)


### Features

* **action:** add the GitHub Action, releases and complete README (Phase 4) ([#5](https://github.com/jimmykurian/ci-parity/issues/5)) ([b3dab72](https://github.com/jimmykurian/ci-parity/commit/b3dab72cb64a1cb02fb3a712396747a27a0e69c3))
* **baseline:** read, write and apply the inline-tool ratchet ([db0a183](https://github.com/jimmykurian/ci-parity/commit/db0a183aae8cac67d45135a080c0dfc926677625))
* **cli:** add subcommand dispatch and version command ([d5e9b58](https://github.com/jimmykurian/ci-parity/commit/d5e9b58e4ad842c4d579e1775898f4226d341bd4))
* **cli:** add the baseline subcommand and check --strict, --baseline and formats ([d58c51a](https://github.com/jimmykurian/ci-parity/commit/d58c51af186b505b55c64f5be0be9ecd887aab7b))
* **cli:** load the Taskfile for check ([f3442b0](https://github.com/jimmykurian/ci-parity/commit/f3442b0e8ee8dfda9085001f3fb8a97a2659237a))
* **cli:** wire check end to end in text format ([1342cde](https://github.com/jimmykurian/ci-parity/commit/1342cde038b11a08e14ee880a7bd5b06e0f160f0))
* **config:** load and validate .ci-parity.yml ([2be8ea2](https://github.com/jimmykurian/ci-parity/commit/2be8ea2b1784d3ed9872f0108dd0453b5cc4bf6a))
* **report:** add github and json formats and the text ratchet block ([9e2eac9](https://github.com/jimmykurian/ci-parity/commit/9e2eac95b05c306912a08a61cd95399b16efec4e))
* **report:** add text reporter ([6c268f6](https://github.com/jimmykurian/ci-parity/commit/6c268f6d362715ea1fdcd282dd3b4ea09d4968e7))
* **rules:** report P001 inline tools and P005 unsupported steps ([2647c35](https://github.com/jimmykurian/ci-parity/commit/2647c35427521e8c5f5aa9ba9e15270290e2bc0c))
* **rules:** report P002 unknown and P003 internal tasks with suggestions ([0ca22e7](https://github.com/jimmykurian/ci-parity/commit/0ca22e7a02f00e10ad1c05a06aae7c2f218a3aa4))
* **rules:** treat pnx as a built-in inline tool ([c932967](https://github.com/jimmykurian/ci-parity/commit/c93296701d49630687b17f9173c9053c8fd1c283))
* **runner:** add Runner interface and static Taskfile reader ([a0bd431](https://github.com/jimmykurian/ci-parity/commit/a0bd43105ff471305902ebe4422e68254d5186f1))
* **shell:** extract tool and invoker calls from run scripts ([1785c9e](https://github.com/jimmykurian/ci-parity/commit/1785c9e9659e027dc0898688d3734c34dc050201))
* **shell:** keep argument and action input positions ([1395a56](https://github.com/jimmykurian/ci-parity/commit/1395a56e5708be8c1a43767923f0afc2e42d4cb2))
* **workflow:** enumerate steps with source positions ([7e0c1a6](https://github.com/jimmykurian/ci-parity/commit/7e0c1a61247525f160f7c16a10b9fd062d5e3318))


### Bug Fixes

* **ci:** install actionlint with mise-action ([921bf72](https://github.com/jimmykurian/ci-parity/commit/921bf72116d26a64daa9413aa6f9686d4b8cea08))
* **ci:** pin actionlint to a version install-action supports ([8195bae](https://github.com/jimmykurian/ci-parity/commit/8195bae2aa27964665726a5b37b8138bfa91ca2c))
* **cli:** report a missing -C directory and pluralize baseline output ([21a341d](https://github.com/jimmykurian/ci-parity/commit/21a341d66b9dacadbc189094ab0859576d059f64))
* **fixtures:** check only listed formats and require every golden ([339eda3](https://github.com/jimmykurian/ci-parity/commit/339eda3addb5a7cb2be2ffdccbc1e06b7a5847a8))
* **release:** start Release Please at v0.1.0 ([#7](https://github.com/jimmykurian/ci-parity/issues/7)) ([174219f](https://github.com/jimmykurian/ci-parity/commit/174219f0dd21996feb045a6998a56d03f31dbc1b))
* **rules:** give P002 suggestions the task &lt;name&gt; form ([607b465](https://github.com/jimmykurian/ci-parity/commit/607b4650882eb330f478008c36a33255cd08cb36))
* **rules:** report parse errors at the parser's position ([3b1e5ca](https://github.com/jimmykurian/ci-parity/commit/3b1e5cabd289a2eeb6bdbdb066702233dff13a2f))
* **shell:** skip wrapper flag values and map positions through expressions ([65a5cf9](https://github.com/jimmykurian/ci-parity/commit/65a5cf9c7a286cd82f00e155194a912ca91c2bcb))
* **tooling:** mark the devcontainer workspace safe for git and lock the mise feature ([339162d](https://github.com/jimmykurian/ci-parity/commit/339162dd0a88f711b5945d264cdb332f7c64888f))
* **tooling:** skip the generated CHANGELOG in markdownlint and typos ([#2](https://github.com/jimmykurian/ci-parity/issues/2)) ([4c2580b](https://github.com/jimmykurian/ci-parity/commit/4c2580be0c311d2985e545d9ceee753ce4e07b0f))
* **workflow:** report exact lines for folded and multi-line plain values ([00e24ad](https://github.com/jimmykurian/ci-parity/commit/00e24ad6058447d729480752c048831f6625b92c))


### Documentation

* add a status table to the README ([#4](https://github.com/jimmykurian/ci-parity/issues/4)) ([32013f1](https://github.com/jimmykurian/ci-parity/commit/32013f16ecb42350ae35ae5fd7529d728e12ed7b))
* add ACCESSIBILITY.md with output and documentation rules ([6193b2a](https://github.com/jimmykurian/ci-parity/commit/6193b2ad30cac2b1b2d7a62b4a2885cb5c2182d6))
* add issue and pull request templates and community files ([caf26cf](https://github.com/jimmykurian/ci-parity/commit/caf26cfe80d16f404c7e8e83596f9870b0228b4c))
* add README and decisions log, tick Phase 0 ([77529c6](https://github.com/jimmykurian/ci-parity/commit/77529c6f78a449ca0961da0f8d677cd2b8e732cb))
* add v0.1 spec, build plan and agent instructions ([8dbda98](https://github.com/jimmykurian/ci-parity/commit/8dbda98944e5b907c41d2a1c2bbd8ad6acd1ffb3))
* adopt a clean-architecture package layout ([1ef5cca](https://github.com/jimmykurian/ci-parity/commit/1ef5cca6353b4503bf00de43275e6a00b4473ec7))
* document the baseline and JSON schema, tick Phase 3 ([a721173](https://github.com/jimmykurian/ci-parity/commit/a7211734d03100aaa77f69d0ea07ef7505712305))
* drop the CLAUDE.md rule from AGENTS.md ([38a9ad2](https://github.com/jimmykurian/ci-parity/commit/38a9ad20fdf25c883c4c7d6ed59c78c7ef6c3b1f))
* make AGENTS.md the only agent instruction file ([b35a4c3](https://github.com/jimmykurian/ci-parity/commit/b35a4c32c3f7de428e2d5c5a74419c3572b50f7e))
* specify Phase 2 invoker, include and suggestion rules ([562875c](https://github.com/jimmykurian/ci-parity/commit/562875c864cf886900281a2f7b57dd69ecdd7b6f))
* specify wrapper value flags and Phase 1 clarifications ([ca28490](https://github.com/jimmykurian/ci-parity/commit/ca284909bf9f66b4c3ec53861db39c3d4eeb7b45))
* tick Phase 1 ([24fc93e](https://github.com/jimmykurian/ci-parity/commit/24fc93e3fb9da5a917f1491afba8ab55746c5f67))
* tick Phase 2 ([91266e6](https://github.com/jimmykurian/ci-parity/commit/91266e65662dceed94ee41429825b7b71aaba202))


### Code Refactoring

* add app use cases and slim cmd to a composition root ([b4f35a2](https://github.com/jimmykurian/ci-parity/commit/b4f35a2f840b0b584a8c6c385a61b61fed9cdda8))
* **config:** keep YAML key names out of domain.Config ([aaf69cc](https://github.com/jimmykurian/ci-parity/commit/aaf69cc56ff6ded4cbdd3d6dbf0e12f55a687e5f))
* extract types and pure logic into internal/domain ([a48d388](https://github.com/jimmykurian/ci-parity/commit/a48d388343ec4f67901200c8f9303f62ef04b6b4))
* move parsers and I/O behind ports into internal/adapters ([ccab24a](https://github.com/jimmykurian/ci-parity/commit/ccab24ac9c9c7ecb3a0fe961a7828b63a17682e6))
