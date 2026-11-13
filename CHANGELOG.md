# Changelog

All notable changes to this project will be documented in this file. (after 0.0.1)

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/en/2.0.0/).

## [0.9.0](https://github.com/flushwhy/fe/compare/v0.8.0...v0.9.0) (2026-11-13)


### Features

* add Blender add-on validation and packaging ([#108](https://github.com/flushwhy/fe/issues/108)) ([a18c6cf](https://github.com/flushwhy/fe/commit/a18c6cf14de85d7e27c680081449c86f0141abcf))
* add conservative build suggestions to project inspection ([#83](https://github.com/flushwhy/fe/issues/83)) ([699585f](https://github.com/flushwhy/fe/commit/699585f86191a2f0147178a6ef5813d983e46b7d))
* add named build target selection ([#106](https://github.com/flushwhy/fe/issues/106)) ([fbeca9a](https://github.com/flushwhy/fe/commit/fbeca9a62bbc994acfbbae2091a6b8992b09649b))
* add O3DE build configuration and target selection ([#111](https://github.com/flushwhy/fe/issues/111)) ([cb11bda](https://github.com/flushwhy/fe/commit/cb11bda6c156dabe968a7afe8d866177f7249b6a))
* add package-manager-aware Fe updater ([#92](https://github.com/flushwhy/fe/issues/92)) ([ff422ce](https://github.com/flushwhy/fe/commit/ff422ceb81b846bc117668a746f45d42ffb3ee5c))
* add stable Gumroad CLI releases ([#201](https://github.com/flushwhy/fe/issues/201)) ([ae2e809](https://github.com/flushwhy/fe/commit/ae2e809d47ff099a1fe132b76487d1459e82912c))
* **clean:** confirm destructive cleanup interactively ([#185](https://github.com/flushwhy/fe/issues/185)) ([7f1e832](https://github.com/flushwhy/fe/commit/7f1e8321189502c45e97930998bec40e447813ae))
* **cli:** add interactive project and tool setup ([#183](https://github.com/flushwhy/fe/issues/183)) ([47d3e31](https://github.com/flushwhy/fe/commit/47d3e31161e41851343faca9beb4d01fdccfa9ea))
* **cli:** polish errors and interactive entry point ([#182](https://github.com/flushwhy/fe/issues/182)) ([e186c15](https://github.com/flushwhy/fe/commit/e186c15275b7c3915dc039f4356157bef5343bb9))
* extend release orchestration ([#104](https://github.com/flushwhy/fe/issues/104)) ([16e9607](https://github.com/flushwhy/fe/commit/16e960752203c8acc807b1bfd26bf86c1c63b1f5))
* generate changelogs from Git history ([#102](https://github.com/flushwhy/fe/issues/102)) ([01341a1](https://github.com/flushwhy/fe/commit/01341a122893750d2b4f33d0d17df64f12b32f57))
* **godot:** discover Linux dedicated-server export presets ([#149](https://github.com/flushwhy/fe/issues/149)) ([c211fb6](https://github.com/flushwhy/fe/commit/c211fb6d8734a8693b541817362ab62aad78bb24))
* improve template metadata and offline discovery ([#109](https://github.com/flushwhy/fe/issues/109)) ([e1b61d4](https://github.com/flushwhy/fe/commit/e1b61d49dcb083b310f8e359e3de56747a525a38))
* package build artifacts as archives ([#103](https://github.com/flushwhy/fe/issues/103)) ([391eb46](https://github.com/flushwhy/fe/commit/391eb4658be247f613d29405b4f970d861f567f1))
* **project:** sync Godot toolchains and reversible cleanup ([#154](https://github.com/flushwhy/fe/issues/154)) ([1894a5d](https://github.com/flushwhy/fe/commit/1894a5da8ff8fd5eff2afbfafbd1a4f779fe5487))
* **release:** add Git safety guardrails ([#145](https://github.com/flushwhy/fe/issues/145)) ([9fb338c](https://github.com/flushwhy/fe/commit/9fb338c6ee0cd6b6a1817f7f721d44e35adb71ce))
* repair CMake compilation databases for clangd ([#167](https://github.com/flushwhy/fe/issues/167)) ([f7bb41e](https://github.com/flushwhy/fe/commit/f7bb41e6ef9eb81f3d6eebdb1f9617ee530dac42))
* **setup:** make toolchain results clearer ([#190](https://github.com/flushwhy/fe/issues/190)) ([6495ad8](https://github.com/flushwhy/fe/commit/6495ad8cf0911f9e65caa0a01495fc04556f36ce))
* **site:** add expandable command help explorer ([#129](https://github.com/flushwhy/fe/issues/129)) ([793895b](https://github.com/flushwhy/fe/commit/793895bb8764bd331e1b126acfdb41156cb7bb08))
* **site:** add GitHub release download buttons ([#147](https://github.com/flushwhy/fe/issues/147)) ([d1e4186](https://github.com/flushwhy/fe/commit/d1e4186c381d4d47b57fe458c1a5511dba9912a4))
* **site:** redesign Fe website with clay UI and terminal intro ([#128](https://github.com/flushwhy/fe/issues/128)) ([682900a](https://github.com/flushwhy/fe/commit/682900a3580472cf2abc1cd72d5df2fec52b2c04))
* **transcode:** add game-friendly video defaults ([#127](https://github.com/flushwhy/fe/issues/127)) ([12be118](https://github.com/flushwhy/fe/commit/12be1186d1a586dfe70e591c9c470d18f745043b))
* validate build target configuration ([#107](https://github.com/flushwhy/fe/issues/107)) ([9973d63](https://github.com/flushwhy/fe/commit/9973d634142226a58e4f7d73607f4f70950660ce))


### Bug Fixes

* avoid duplicate RPM checksum entries ([#97](https://github.com/flushwhy/fe/issues/97)) ([667b915](https://github.com/flushwhy/fe/commit/667b91508f620183ca45df87b03fa8fffb00fdcc))
* **blender:** accept extension manifest type arrays ([#141](https://github.com/flushwhy/fe/issues/141)) ([ab77bd2](https://github.com/flushwhy/fe/commit/ab77bd2d5df1c59eceadebe895ad3295240bf41c))
* **build:** validate target selection before preflight ([#144](https://github.com/flushwhy/fe/issues/144)) ([e2ca61a](https://github.com/flushwhy/fe/commit/e2ca61afc11a952dec99120241935babf07fc2ee))
* **ci:** allow generated staging files in itch release ([#178](https://github.com/flushwhy/fe/issues/178)) ([cacfd7f](https://github.com/flushwhy/fe/commit/cacfd7f8dc39d1382afb4b35c4b2ed346284c8e1))
* **ci:** avoid duplicate release checksums ([#189](https://github.com/flushwhy/fe/issues/189)) ([bdd3923](https://github.com/flushwhy/fe/commit/bdd3923d9b611e4cf1ff4f35c6a77c4366fb8524))
* **ci:** install local Debian packages with apt ([#170](https://github.com/flushwhy/fe/issues/170)) ([1e55c96](https://github.com/flushwhy/fe/commit/1e55c96874fc52a465fd77c51a32d2bcfd236c23))
* **godot:** avoid false native extension prerequisite warning ([#140](https://github.com/flushwhy/fe/issues/140)) ([d4e2d65](https://github.com/flushwhy/fe/commit/d4e2d651f85e325818a31b8daa29527e964185e4))
* **godot:** resolve relative export paths from project root ([#143](https://github.com/flushwhy/fe/issues/143)) ([9622825](https://github.com/flushwhy/fe/commit/962282550414ecc15961ef26aab06b4f2a913535))
* **godot:** validate export presets before running ([#148](https://github.com/flushwhy/fe/issues/148)) ([745e6af](https://github.com/flushwhy/fe/commit/745e6af86c7f1b36faec410b9188594137380827))
* **o3de:** run Windows CLI batch scripts correctly ([#151](https://github.com/flushwhy/fe/issues/151)) ([a2bd289](https://github.com/flushwhy/fe/commit/a2bd2894f02207746cb142fdff1cad7ee873d149))
* **o3de:** validate project and engine paths ([#152](https://github.com/flushwhy/fe/issues/152)) ([0b547d5](https://github.com/flushwhy/fe/commit/0b547d542c163ca94e26bf6e6fc5075963328bf1))
* **packaging:** declare correct RPM license ([#139](https://github.com/flushwhy/fe/issues/139)) ([978eab0](https://github.com/flushwhy/fe/commit/978eab091e182dbac7fc675beb7f1aa17dd69784))
* **release:** detect tag release branch by ancestry ([#191](https://github.com/flushwhy/fe/issues/191)) ([e2aeaa6](https://github.com/flushwhy/fe/commit/e2aeaa6b3ef94a3b33c50c2e018b086b87408c83))
* **release:** isolate maintenance release publishing ([#134](https://github.com/flushwhy/fe/issues/134)) ([3056f22](https://github.com/flushwhy/fe/commit/3056f22788637bb60e5d6c3f45b4177d3c15b338))
* **release:** keep Butler download outside checkout ([#195](https://github.com/flushwhy/fe/issues/195)) ([a03c81e](https://github.com/flushwhy/fe/commit/a03c81e3bdfb9da0743e5c7e74ea9146b7b67579))
* **release:** show files blocking dirty-tree preflight ([#179](https://github.com/flushwhy/fe/issues/179)) ([abfe553](https://github.com/flushwhy/fe/commit/abfe5533780e6e0123dad2cbab37f10957287411))
* **security:** bound archive and image processing ([#113](https://github.com/flushwhy/fe/issues/113)) ([2b717e4](https://github.com/flushwhy/fe/commit/2b717e4f167029748ca1ee149de29b81b1dac8d4))
* **security:** protect filesystem writes and rolling assets ([#125](https://github.com/flushwhy/fe/issues/125)) ([6912caf](https://github.com/flushwhy/fe/commit/6912cafd83a612c71a641895cf704d0447ae501c))
* **security:** update Go to 1.26.11 ([#196](https://github.com/flushwhy/fe/issues/196)) ([57408ff](https://github.com/flushwhy/fe/commit/57408ff43ee0a5b24935761f20b2a457adb5fa84))
* **security:** validate release tags before publishing ([#112](https://github.com/flushwhy/fe/issues/112)) ([3020903](https://github.com/flushwhy/fe/commit/302090330f537e8f5b488ea79fb7eb0e57487f3e))
* select Godot engines by project compatibility ([#110](https://github.com/flushwhy/fe/issues/110)) ([5160924](https://github.com/flushwhy/fe/commit/5160924634d4b224e546f31b01b9fa4d41e5acf1))
* **template:** validate list command arguments ([#153](https://github.com/flushwhy/fe/issues/153)) ([95d322b](https://github.com/flushwhy/fe/commit/95d322b1e3c68e81ac68d2445ca15f4c88b1de8d))
* unblock Debian and stable release publishing ([#93](https://github.com/flushwhy/fe/issues/93)) ([6ed4bb2](https://github.com/flushwhy/fe/commit/6ed4bb2aedfb71ed3d3b8e65dc4d0dcd3be3e499))

## [0.8.0](https://github.com/flushwhy/fe/compare/v0.7.1...v0.8.0) (2026-10-31)


### Features

* accept multiple watch ignore dirs in one flag ([#65](https://github.com/flushwhy/fe/issues/65)) ([4a5bddb](https://github.com/flushwhy/fe/commit/4a5bddb1e0bf0afd81e60539a56aca8edf7d77bc))
* add Chocolatey package metadata ([365850f](https://github.com/flushwhy/fe/commit/365850fdd3b35c8bb52c54b538319078d4712fbd))
* add Chocolatey package metadata ([80cfb11](https://github.com/flushwhy/fe/commit/80cfb1113fce5c12e162172bd7560b46e2a8bacb))
* add Chocolatey package metadata ([f8e6f45](https://github.com/flushwhy/fe/commit/f8e6f45c552c45fec9e244aaf04826d367b9cec6))
* add configurable build preflight checks ([#75](https://github.com/flushwhy/fe/issues/75)) ([87db860](https://github.com/flushwhy/fe/commit/87db860ec0f6d40b9357d0d9a09b6d8e4f196ace))
* add disposable Docker and Podman builds ([#50](https://github.com/flushwhy/fe/issues/50)) ([b71be0a](https://github.com/flushwhy/fe/commit/b71be0ac5fd506ddb36e7029b8cdcc36cea5ecab))
* add existing funding links to site support section ([#62](https://github.com/flushwhy/fe/issues/62)) ([bcc8b4f](https://github.com/flushwhy/fe/commit/bcc8b4fba754e7273632fbdf384f2ceec7f1b3eb))
* add Godot tooling and native extension workflow ([#43](https://github.com/flushwhy/fe/issues/43)) ([37dcc06](https://github.com/flushwhy/fe/commit/37dcc067b1966396c89ca1d2ae341696d7922c93))
* add language-aware static checker ([#67](https://github.com/flushwhy/fe/issues/67)) ([3cbf9a1](https://github.com/flushwhy/fe/commit/3cbf9a10248e2c9f4e390dfbd33bebfdb0a6c182))
* add O3DE discovery and doctor ([#23](https://github.com/flushwhy/fe/issues/23)) ([02d4c2e](https://github.com/flushwhy/fe/commit/02d4c2e883a5099364b8c4be94d11acc24ad87f6))
* add project setup guidance ([#79](https://github.com/flushwhy/fe/issues/79)) ([9f15757](https://github.com/flushwhy/fe/commit/9f1575757aa0e2f0a5eb3c1c5851f98cedd05fb7))
* add read-only project inspection ([#78](https://github.com/flushwhy/fe/issues/78)) ([6235351](https://github.com/flushwhy/fe/commit/62353512f908526c38aec72eabec13ba87c24174))
* add safe Odin repository installer ([103c12d](https://github.com/flushwhy/fe/commit/103c12d66ecf785488a714eb1a3f161b30ced62b))
* add workspace-aware fe watch command ([#39](https://github.com/flushwhy/fe/issues/39)) ([bff470f](https://github.com/flushwhy/fe/commit/bff470f74515e724bff730fdc306ebf85f2f200d))
* automate Chocolatey packaging on releases ([f42c748](https://github.com/flushwhy/fe/commit/f42c7484b8026d925fb9ed947e721bc90e8e69a0))
* download and cache named project templates ([#22](https://github.com/flushwhy/fe/issues/22)) ([77346c2](https://github.com/flushwhy/fe/commit/77346c2a996b6d7e37c87973433525982b571b9b))
* expose install and get commands ([61a805a](https://github.com/flushwhy/fe/commit/61a805a00c34f8ef8ca2c40b6f23a4c6bf668dce))
* **godot:** add headless project export command ([#47](https://github.com/flushwhy/fe/issues/47)) ([1983754](https://github.com/flushwhy/fe/commit/19837540b299140394c052873c788fec142131d3))
* improve release previews, diagnostics, and dev workflows ([#64](https://github.com/flushwhy/fe/issues/64)) ([d7f32e5](https://github.com/flushwhy/fe/commit/d7f32e510d2655655cb4ce3b1965a2cae81b094c))
* **o3de:** add Gem management commands ([#37](https://github.com/flushwhy/fe/issues/37)) ([0ae3712](https://github.com/flushwhy/fe/commit/0ae37124beb339a6cd52fe33849f4e8031d5872f))
* **o3de:** expand project and engine workflows ([#38](https://github.com/flushwhy/fe/issues/38)) ([e5ec464](https://github.com/flushwhy/fe/commit/e5ec4642967c8712a373cf3ad7c29adb94a32984))
* publish Debian packages for Linux ([#26](https://github.com/flushwhy/fe/issues/26)) ([e09f4e8](https://github.com/flushwhy/fe/commit/e09f4e844e20eacfd099a71390de60756814f49e))
* publish Linux RISC-V binaries in releases ([#66](https://github.com/flushwhy/fe/issues/66)) ([47b5907](https://github.com/flushwhy/fe/commit/47b59071e5eec76a48cb7e3f19bb2d2bd373a0e4))
* refresh site with docs-first design and support section ([#61](https://github.com/flushwhy/fe/issues/61)) ([cc67186](https://github.com/flushwhy/fe/commit/cc671860c79ab54a980e60e140e1535b83e05fc1))
* support multiple named spritesheets ([#21](https://github.com/flushwhy/fe/issues/21)) ([a44bde1](https://github.com/flushwhy/fe/commit/a44bde15cda042988c7f546cb4c1f2200ddf7d5c))


### Bug Fixes

* align release Go version and smoke-test Debian install ([#27](https://github.com/flushwhy/fe/issues/27)) ([df0581d](https://github.com/flushwhy/fe/commit/df0581d19fe72ea092a624190022bb24f86d2e21))
* build ARM64 RPMs on a native runner ([#56](https://github.com/flushwhy/fe/issues/56)) ([f8ddabb](https://github.com/flushwhy/fe/commit/f8ddabb0acad66bd8d19c7409181c46f252ca0c6))
* build binaries and attach them to GitHub releases ([25c1b68](https://github.com/flushwhy/fe/commit/25c1b68fe95eccb94f971f240b99a79f90e0be6b))
* conditionally publish Chocolatey package using secret ([b1ed148](https://github.com/flushwhy/fe/commit/b1ed1484ada18dd5fab1011a4e537140d5207c0d))
* fail early when release build output is missing ([#32](https://github.com/flushwhy/fe/issues/32)) ([7e02d6f](https://github.com/flushwhy/fe/commit/7e02d6f7b3b76494b8def4e449405255de098099))
* **godot:** Fixed issues with escape characters ([#44](https://github.com/flushwhy/fe/issues/44)) ([92823a8](https://github.com/flushwhy/fe/commit/92823a823079b9433d7bd91191ab013c06ff5962))
* handle TUI executable lookup errors ([#19](https://github.com/flushwhy/fe/issues/19)) ([08c0da4](https://github.com/flushwhy/fe/commit/08c0da4cb38a9a5241a5e36f9fdaf9fe447c0224))
* harden paths and command execution ([#73](https://github.com/flushwhy/fe/issues/73)) ([64603ee](https://github.com/flushwhy/fe/commit/64603ee8ef6ab3bdf2092dbaf8b2a6a9a7ec6c62))
* limit extracted archive size and tighten installer permissions ([4df259f](https://github.com/flushwhy/fe/commit/4df259fefe5bde0f140deeca88b9d6caebc3da6c))
* limit total extracted archive size ([419f152](https://github.com/flushwhy/fe/commit/419f15248f44c43637f40847e3f85cbbe6d9f093))
* make generated Godot extensions loadable ([#45](https://github.com/flushwhy/fe/issues/45)) ([8a73b4c](https://github.com/flushwhy/fe/commit/8a73b4c62d5e9c5390d1b196d21b4d46e8a5b3a4))
* make TUI command interaction reliable ([#24](https://github.com/flushwhy/fe/issues/24)) ([1dcd640](https://github.com/flushwhy/fe/commit/1dcd640ea083cb0ad03e4ef9866307c3c8d6758e))
* remove duplicated rolling workflow jobs ([#58](https://github.com/flushwhy/fe/issues/58)) ([1aea03d](https://github.com/flushwhy/fe/commit/1aea03d6e1944a7856fdfe197030e05af6cfc5bf))
* remove stale Chocolatey packages from rolling release ([#49](https://github.com/flushwhy/fe/issues/49)) ([2fc4971](https://github.com/flushwhy/fe/commit/2fc49713a0d0b181589e349d7888fe97be2b85aa))
* remove stale rolling release assets ([#63](https://github.com/flushwhy/fe/issues/63)) ([f0a5e19](https://github.com/flushwhy/fe/commit/f0a5e1924b995d359ebcbb1b621ffce318af3156))
* remove stray duplicate workflow lines ([#57](https://github.com/flushwhy/fe/issues/57)) ([c193ba3](https://github.com/flushwhy/fe/commit/c193ba3f6aa9655c165ed3c5f27dfc052bcb1c71))
* remove unused context import from check command ([#69](https://github.com/flushwhy/fe/issues/69)) ([d4379a2](https://github.com/flushwhy/fe/commit/d4379a292bdb92169d01bda2fa1ef04c9dd17b77))
* repair rolling Chocolatey release workflow ([#52](https://github.com/flushwhy/fe/issues/52)) ([63dfd5c](https://github.com/flushwhy/fe/commit/63dfd5c218f76f3f449f02a73cabe2fe38033f93))
* repair rolling release workflow YAML ([#42](https://github.com/flushwhy/fe/issues/42)) ([bff53c3](https://github.com/flushwhy/fe/commit/bff53c324ce2aec5f032174dec8e437d3eac54c7))
* restore rolling releases with Linux packages ([#54](https://github.com/flushwhy/fe/issues/54)) ([afd35af](https://github.com/flushwhy/fe/commit/afd35afcd3d827ad8faca109d08bbadfc2302899))
* retry transient Chocolatey rolling push failures ([#48](https://github.com/flushwhy/fe/issues/48)) ([c07c205](https://github.com/flushwhy/fe/commit/c07c20501179309e2400046e34f7694c92f3e03a))
* stabilize patch releases and Linux RPM packaging ([#81](https://github.com/flushwhy/fe/issues/81)) ([d4c5f66](https://github.com/flushwhy/fe/commit/d4c5f66e579716d1ac3f6653fbb0c5053b3e7516))
* support release-please tag names in Chocolatey ([5d40001](https://github.com/flushwhy/fe/commit/5d4000111221c8610cebe483c9a9cd06128fcda8))
* tighten remote downloads and file permissions ([#74](https://github.com/flushwhy/fe/issues/74)) ([3e935f8](https://github.com/flushwhy/fe/commit/3e935f81c9b0c5dc0ffdea7e1d7da7f5a965d3a6))
* update itch.io butler download URL ([3e155ba](https://github.com/flushwhy/fe/commit/3e155ba5fbc58f16b2a549e175542c29c1d8c66d))
* use active itch.io butler download domain ([055e3f4](https://github.com/flushwhy/fe/commit/055e3f40c454091b422533ab780e40599e5c5c7f))
* use standard RPM target triplets ([#55](https://github.com/flushwhy/fe/issues/55)) ([4c4704b](https://github.com/flushwhy/fe/commit/4c4704b8bc2ccee25fc88b7ecb695870890b4154))
* validate archive extraction paths ([1977a57](https://github.com/flushwhy/fe/commit/1977a57860cc04d0b8c3b0b203e86d961a170d5f))
* validate only required project paths ([#16](https://github.com/flushwhy/fe/issues/16)) ([d1c5f67](https://github.com/flushwhy/fe/commit/d1c5f675d13a3b1140f1c4028d1b9638a1b1455a))

## [0.7.1](https://github.com/flushwhy/fe/compare/v0.7.0...v0.7.1) (2026-10-30)


### Bug Fixes

* stabilize patch releases and Linux RPM packaging ([#81](https://github.com/flushwhy/fe/issues/81)) ([d4c5f66](https://github.com/flushwhy/fe/commit/d4c5f66e579716d1ac3f6653fbb0c5053b3e7516))

## [0.7.0](https://github.com/flushwhy/fe/compare/v0.6.0...v0.7.0) (2026-10-30)


### Features

* accept multiple watch ignore dirs in one flag ([#65](https://github.com/flushwhy/fe/issues/65)) ([4a5bddb](https://github.com/flushwhy/fe/commit/4a5bddb1e0bf0afd81e60539a56aca8edf7d77bc))
* add configurable build preflight checks ([#75](https://github.com/flushwhy/fe/issues/75)) ([87db860](https://github.com/flushwhy/fe/commit/87db860ec0f6d40b9357d0d9a09b6d8e4f196ace))
* add disposable Docker and Podman builds ([#50](https://github.com/flushwhy/fe/issues/50)) ([b71be0a](https://github.com/flushwhy/fe/commit/b71be0ac5fd506ddb36e7029b8cdcc36cea5ecab))
* add existing funding links to site support section ([#62](https://github.com/flushwhy/fe/issues/62)) ([bcc8b4f](https://github.com/flushwhy/fe/commit/bcc8b4fba754e7273632fbdf384f2ceec7f1b3eb))
* add language-aware static checker ([#67](https://github.com/flushwhy/fe/issues/67)) ([3cbf9a1](https://github.com/flushwhy/fe/commit/3cbf9a10248e2c9f4e390dfbd33bebfdb0a6c182))
* add project setup guidance ([#79](https://github.com/flushwhy/fe/issues/79)) ([9f15757](https://github.com/flushwhy/fe/commit/9f1575757aa0e2f0a5eb3c1c5851f98cedd05fb7))
* add read-only project inspection ([#78](https://github.com/flushwhy/fe/issues/78)) ([6235351](https://github.com/flushwhy/fe/commit/62353512f908526c38aec72eabec13ba87c24174))
* **godot:** add headless project export command ([#47](https://github.com/flushwhy/fe/issues/47)) ([1983754](https://github.com/flushwhy/fe/commit/19837540b299140394c052873c788fec142131d3))
* improve release previews, diagnostics, and dev workflows ([#64](https://github.com/flushwhy/fe/issues/64)) ([d7f32e5](https://github.com/flushwhy/fe/commit/d7f32e510d2655655cb4ce3b1965a2cae81b094c))
* publish Linux RISC-V binaries in releases ([#66](https://github.com/flushwhy/fe/issues/66)) ([47b5907](https://github.com/flushwhy/fe/commit/47b59071e5eec76a48cb7e3f19bb2d2bd373a0e4))
* refresh site with docs-first design and support section ([#61](https://github.com/flushwhy/fe/issues/61)) ([cc67186](https://github.com/flushwhy/fe/commit/cc671860c79ab54a980e60e140e1535b83e05fc1))


### Bug Fixes

* build ARM64 RPMs on a native runner ([#56](https://github.com/flushwhy/fe/issues/56)) ([f8ddabb](https://github.com/flushwhy/fe/commit/f8ddabb0acad66bd8d19c7409181c46f252ca0c6))
* harden paths and command execution ([#73](https://github.com/flushwhy/fe/issues/73)) ([64603ee](https://github.com/flushwhy/fe/commit/64603ee8ef6ab3bdf2092dbaf8b2a6a9a7ec6c62))
* make generated Godot extensions loadable ([#45](https://github.com/flushwhy/fe/issues/45)) ([8a73b4c](https://github.com/flushwhy/fe/commit/8a73b4c62d5e9c5390d1b196d21b4d46e8a5b3a4))
* remove duplicated rolling workflow jobs ([#58](https://github.com/flushwhy/fe/issues/58)) ([1aea03d](https://github.com/flushwhy/fe/commit/1aea03d6e1944a7856fdfe197030e05af6cfc5bf))
* remove stale Chocolatey packages from rolling release ([#49](https://github.com/flushwhy/fe/issues/49)) ([2fc4971](https://github.com/flushwhy/fe/commit/2fc49713a0d0b181589e349d7888fe97be2b85aa))
* remove stale rolling release assets ([#63](https://github.com/flushwhy/fe/issues/63)) ([f0a5e19](https://github.com/flushwhy/fe/commit/f0a5e1924b995d359ebcbb1b621ffce318af3156))
* remove stray duplicate workflow lines ([#57](https://github.com/flushwhy/fe/issues/57)) ([c193ba3](https://github.com/flushwhy/fe/commit/c193ba3f6aa9655c165ed3c5f27dfc052bcb1c71))
* remove unused context import from check command ([#69](https://github.com/flushwhy/fe/issues/69)) ([d4379a2](https://github.com/flushwhy/fe/commit/d4379a292bdb92169d01bda2fa1ef04c9dd17b77))
* repair rolling Chocolatey release workflow ([#52](https://github.com/flushwhy/fe/issues/52)) ([63dfd5c](https://github.com/flushwhy/fe/commit/63dfd5c218f76f3f449f02a73cabe2fe38033f93))
* restore rolling releases with Linux packages ([#54](https://github.com/flushwhy/fe/issues/54)) ([afd35af](https://github.com/flushwhy/fe/commit/afd35afcd3d827ad8faca109d08bbadfc2302899))
* retry transient Chocolatey rolling push failures ([#48](https://github.com/flushwhy/fe/issues/48)) ([c07c205](https://github.com/flushwhy/fe/commit/c07c20501179309e2400046e34f7694c92f3e03a))
* tighten remote downloads and file permissions ([#74](https://github.com/flushwhy/fe/issues/74)) ([3e935f8](https://github.com/flushwhy/fe/commit/3e935f81c9b0c5dc0ffdea7e1d7da7f5a965d3a6))
* use standard RPM target triplets ([#55](https://github.com/flushwhy/fe/issues/55)) ([4c4704b](https://github.com/flushwhy/fe/commit/4c4704b8bc2ccee25fc88b7ecb695870890b4154))

## [0.6.0](https://github.com/flushwhy/fe/compare/v0.5.0...v0.6.0) (2026-10-19)


### Features

* add Godot tooling and native extension workflow ([#43](https://github.com/flushwhy/fe/issues/43)) ([37dcc06](https://github.com/flushwhy/fe/commit/37dcc067b1966396c89ca1d2ae341696d7922c93))
* add workspace-aware fe watch command ([#39](https://github.com/flushwhy/fe/issues/39)) ([bff470f](https://github.com/flushwhy/fe/commit/bff470f74515e724bff730fdc306ebf85f2f200d))


### Bug Fixes

* **godot:** Fixed issues with escape characters ([#44](https://github.com/flushwhy/fe/issues/44)) ([92823a8](https://github.com/flushwhy/fe/commit/92823a823079b9433d7bd91191ab013c06ff5962))
* repair rolling release workflow YAML ([#42](https://github.com/flushwhy/fe/issues/42)) ([bff53c3](https://github.com/flushwhy/fe/commit/bff53c324ce2aec5f032174dec8e437d3eac54c7))

## [0.5.0](https://github.com/flushwhy/fe/compare/v0.4.0...v0.5.0) (2026-10-14)


### Features

* add Chocolatey package metadata ([365850f](https://github.com/flushwhy/fe/commit/365850fdd3b35c8bb52c54b538319078d4712fbd))
* add Chocolatey package metadata ([80cfb11](https://github.com/flushwhy/fe/commit/80cfb1113fce5c12e162172bd7560b46e2a8bacb))
* add Chocolatey package metadata ([f8e6f45](https://github.com/flushwhy/fe/commit/f8e6f45c552c45fec9e244aaf04826d367b9cec6))
* add O3DE discovery and doctor ([#23](https://github.com/flushwhy/fe/issues/23)) ([02d4c2e](https://github.com/flushwhy/fe/commit/02d4c2e883a5099364b8c4be94d11acc24ad87f6))
* add safe Odin repository installer ([103c12d](https://github.com/flushwhy/fe/commit/103c12d66ecf785488a714eb1a3f161b30ced62b))
* automate Chocolatey packaging on releases ([f42c748](https://github.com/flushwhy/fe/commit/f42c7484b8026d925fb9ed947e721bc90e8e69a0))
* download and cache named project templates ([#22](https://github.com/flushwhy/fe/issues/22)) ([77346c2](https://github.com/flushwhy/fe/commit/77346c2a996b6d7e37c87973433525982b571b9b))
* expose install and get commands ([61a805a](https://github.com/flushwhy/fe/commit/61a805a00c34f8ef8ca2c40b6f23a4c6bf668dce))
* **o3de:** add Gem management commands ([#37](https://github.com/flushwhy/fe/issues/37)) ([0ae3712](https://github.com/flushwhy/fe/commit/0ae37124beb339a6cd52fe33849f4e8031d5872f))
* **o3de:** expand project and engine workflows ([#38](https://github.com/flushwhy/fe/issues/38)) ([e5ec464](https://github.com/flushwhy/fe/commit/e5ec4642967c8712a373cf3ad7c29adb94a32984))
* publish Debian packages for Linux ([#26](https://github.com/flushwhy/fe/issues/26)) ([e09f4e8](https://github.com/flushwhy/fe/commit/e09f4e844e20eacfd099a71390de60756814f49e))
* support multiple named spritesheets ([#21](https://github.com/flushwhy/fe/issues/21)) ([a44bde1](https://github.com/flushwhy/fe/commit/a44bde15cda042988c7f546cb4c1f2200ddf7d5c))


### Bug Fixes

* align release Go version and smoke-test Debian install ([#27](https://github.com/flushwhy/fe/issues/27)) ([df0581d](https://github.com/flushwhy/fe/commit/df0581d19fe72ea092a624190022bb24f86d2e21))
* build binaries and attach them to GitHub releases ([25c1b68](https://github.com/flushwhy/fe/commit/25c1b68fe95eccb94f971f240b99a79f90e0be6b))
* conditionally publish Chocolatey package using secret ([b1ed148](https://github.com/flushwhy/fe/commit/b1ed1484ada18dd5fab1011a4e537140d5207c0d))
* fail early when release build output is missing ([#32](https://github.com/flushwhy/fe/issues/32)) ([7e02d6f](https://github.com/flushwhy/fe/commit/7e02d6f7b3b76494b8def4e449405255de098099))
* handle TUI executable lookup errors ([#19](https://github.com/flushwhy/fe/issues/19)) ([08c0da4](https://github.com/flushwhy/fe/commit/08c0da4cb38a9a5241a5e36f9fdaf9fe447c0224))
* limit extracted archive size and tighten installer permissions ([4df259f](https://github.com/flushwhy/fe/commit/4df259fefe5bde0f140deeca88b9d6caebc3da6c))
* limit total extracted archive size ([419f152](https://github.com/flushwhy/fe/commit/419f15248f44c43637f40847e3f85cbbe6d9f093))
* make TUI command interaction reliable ([#24](https://github.com/flushwhy/fe/issues/24)) ([1dcd640](https://github.com/flushwhy/fe/commit/1dcd640ea083cb0ad03e4ef9866307c3c8d6758e))
* support release-please tag names in Chocolatey ([5d40001](https://github.com/flushwhy/fe/commit/5d4000111221c8610cebe483c9a9cd06128fcda8))
* update itch.io butler download URL ([3e155ba](https://github.com/flushwhy/fe/commit/3e155ba5fbc58f16b2a549e175542c29c1d8c66d))
* use active itch.io butler download domain ([055e3f4](https://github.com/flushwhy/fe/commit/055e3f40c454091b422533ab780e40599e5c5c7f))
* validate archive extraction paths ([1977a57](https://github.com/flushwhy/fe/commit/1977a57860cc04d0b8c3b0b203e86d961a170d5f))
* validate only required project paths ([#16](https://github.com/flushwhy/fe/issues/16)) ([d1c5f67](https://github.com/flushwhy/fe/commit/d1c5f675d13a3b1140f1c4028d1b9638a1b1455a))

## [0.4.0](https://github.com/flushwhy/fe/compare/v0.3.2...v0.4.0) (2026-10-13)


### Features

* add O3DE discovery and doctor ([#23](https://github.com/flushwhy/fe/issues/23)) ([02d4c2e](https://github.com/flushwhy/fe/commit/02d4c2e883a5099364b8c4be94d11acc24ad87f6))
* download and cache named project templates ([#22](https://github.com/flushwhy/fe/issues/22)) ([77346c2](https://github.com/flushwhy/fe/commit/77346c2a996b6d7e37c87973433525982b571b9b))
* publish Debian packages for Linux ([#26](https://github.com/flushwhy/fe/issues/26)) ([e09f4e8](https://github.com/flushwhy/fe/commit/e09f4e844e20eacfd099a71390de60756814f49e))
* support multiple named spritesheets ([#21](https://github.com/flushwhy/fe/issues/21)) ([a44bde1](https://github.com/flushwhy/fe/commit/a44bde15cda042988c7f546cb4c1f2200ddf7d5c))


### Bug Fixes

* align release Go version and smoke-test Debian install ([#27](https://github.com/flushwhy/fe/issues/27)) ([df0581d](https://github.com/flushwhy/fe/commit/df0581d19fe72ea092a624190022bb24f86d2e21))
* handle TUI executable lookup errors ([#19](https://github.com/flushwhy/fe/issues/19)) ([08c0da4](https://github.com/flushwhy/fe/commit/08c0da4cb38a9a5241a5e36f9fdaf9fe447c0224))
* make TUI command interaction reliable ([#24](https://github.com/flushwhy/fe/issues/24)) ([1dcd640](https://github.com/flushwhy/fe/commit/1dcd640ea083cb0ad03e4ef9866307c3c8d6758e))
* validate only required project paths ([#16](https://github.com/flushwhy/fe/issues/16)) ([d1c5f67](https://github.com/flushwhy/fe/commit/d1c5f675d13a3b1140f1c4028d1b9638a1b1455a))

## [0.3.2](https://github.com/flushwhy/fe/compare/v0.3.1...v0.3.2) (2026-10-11)


### Bug Fixes

* update itch.io butler download URL ([3e155ba](https://github.com/flushwhy/fe/commit/3e155ba5fbc58f16b2a549e175542c29c1d8c66d))
* use active itch.io butler download domain ([055e3f4](https://github.com/flushwhy/fe/commit/055e3f40c454091b422533ab780e40599e5c5c7f))

## [0.3.1](https://github.com/flushwhy/fe/compare/v0.3.0...v0.3.1) (2026-10-11)


### Bug Fixes

* limit extracted archive size and tighten installer permissions ([4df259f](https://github.com/flushwhy/fe/commit/4df259fefe5bde0f140deeca88b9d6caebc3da6c))
* limit total extracted archive size ([419f152](https://github.com/flushwhy/fe/commit/419f15248f44c43637f40847e3f85cbbe6d9f093))
* validate archive extraction paths ([1977a57](https://github.com/flushwhy/fe/commit/1977a57860cc04d0b8c3b0b203e86d961a170d5f))

## [0.3.0](https://github.com/flushwhy/fe/compare/v0.2.0...v0.3.0) (2026-10-11)


### Features

* add Chocolatey package metadata ([365850f](https://github.com/flushwhy/fe/commit/365850fdd3b35c8bb52c54b538319078d4712fbd))
* add Chocolatey package metadata ([80cfb11](https://github.com/flushwhy/fe/commit/80cfb1113fce5c12e162172bd7560b46e2a8bacb))
* add Chocolatey package metadata ([f8e6f45](https://github.com/flushwhy/fe/commit/f8e6f45c552c45fec9e244aaf04826d367b9cec6))
* add safe Odin repository installer ([103c12d](https://github.com/flushwhy/fe/commit/103c12d66ecf785488a714eb1a3f161b30ced62b))
* automate Chocolatey packaging on releases ([f42c748](https://github.com/flushwhy/fe/commit/f42c7484b8026d925fb9ed947e721bc90e8e69a0))
* expose install and get commands ([61a805a](https://github.com/flushwhy/fe/commit/61a805a00c34f8ef8ca2c40b6f23a4c6bf668dce))


### Bug Fixes

* build binaries and attach them to GitHub releases ([25c1b68](https://github.com/flushwhy/fe/commit/25c1b68fe95eccb94f971f240b99a79f90e0be6b))
* conditionally publish Chocolatey package using secret ([b1ed148](https://github.com/flushwhy/fe/commit/b1ed1484ada18dd5fab1011a4e537140d5207c0d))
* support release-please tag names in Chocolatey ([5d40001](https://github.com/flushwhy/fe/commit/5d4000111221c8610cebe483c9a9cd06128fcda8))

## [0.2.0](https://github.com/flushwhy/fe/compare/v0.1.0...v0.2.0) (2026-10-09)


### Features

* add Chocolatey package metadata ([365850f](https://github.com/flushwhy/fe/commit/365850fdd3b35c8bb52c54b538319078d4712fbd))
* add Chocolatey package metadata ([80cfb11](https://github.com/flushwhy/fe/commit/80cfb1113fce5c12e162172bd7560b46e2a8bacb))
* add Chocolatey package metadata ([f8e6f45](https://github.com/flushwhy/fe/commit/f8e6f45c552c45fec9e244aaf04826d367b9cec6))
* add safe Odin repository installer ([103c12d](https://github.com/flushwhy/fe/commit/103c12d66ecf785488a714eb1a3f161b30ced62b))
* automate Chocolatey packaging on releases ([f42c748](https://github.com/flushwhy/fe/commit/f42c7484b8026d925fb9ed947e721bc90e8e69a0))
* expose install and get commands ([61a805a](https://github.com/flushwhy/fe/commit/61a805a00c34f8ef8ca2c40b6f23a4c6bf668dce))


### Bug Fixes

* build binaries and attach them to GitHub releases ([25c1b68](https://github.com/flushwhy/fe/commit/25c1b68fe95eccb94f971f240b99a79f90e0be6b))
* conditionally publish Chocolatey package using secret ([b1ed148](https://github.com/flushwhy/fe/commit/b1ed1484ada18dd5fab1011a4e537140d5207c0d))
* support release-please tag names in Chocolatey ([5d40001](https://github.com/flushwhy/fe/commit/5d4000111221c8610cebe483c9a9cd06128fcda8))

## [0.1.0](https://github.com/flushwhy/fe/compare/fe-v0.0.2...fe-v0.1.0) (2026-10-09)


### Features

* add Chocolatey package metadata ([365850f](https://github.com/flushwhy/fe/commit/365850fdd3b35c8bb52c54b538319078d4712fbd))
* add Chocolatey package metadata ([80cfb11](https://github.com/flushwhy/fe/commit/80cfb1113fce5c12e162172bd7560b46e2a8bacb))
* add Chocolatey package metadata ([f8e6f45](https://github.com/flushwhy/fe/commit/f8e6f45c552c45fec9e244aaf04826d367b9cec6))
* add safe Odin repository installer ([103c12d](https://github.com/flushwhy/fe/commit/103c12d66ecf785488a714eb1a3f161b30ced62b))
* automate Chocolatey packaging on releases ([f42c748](https://github.com/flushwhy/fe/commit/f42c7484b8026d925fb9ed947e721bc90e8e69a0))
* expose install and get commands ([61a805a](https://github.com/flushwhy/fe/commit/61a805a00c34f8ef8ca2c40b6f23a4c6bf668dce))


### Bug Fixes

* build binaries and attach them to GitHub releases ([25c1b68](https://github.com/flushwhy/fe/commit/25c1b68fe95eccb94f971f240b99a79f90e0be6b))
* conditionally publish Chocolatey package using secret ([b1ed148](https://github.com/flushwhy/fe/commit/b1ed1484ada18dd5fab1011a4e537140d5207c0d))

## [Unreleased]

---

## [0.0.4]

### Added
- Full architectural rewrite — all business logic moved to `internal/` packages (`butler`, `transcode`, `pack`, `scaffold`, `build`, `config`, `tty`, `ui`, `tooling`, `doctor`). Cobra commands are now thin wiring only.
- New `fe init <lang> --name <project>` command replacing the old `init`. Supports `odin`, `c`, `cpp`, `zig`, and `go` with per-language templates embedded directly in the binary via `go:embed`.
- New `fe add <tool>` command for wiring up LSP and editor tooling without manual config. Supports `clangd`, `ols`, `zls`, and `gopls`.
  - `fe add clangd` detects your build system (cmake vs bear/make), writes `.clangd` with the correct `CompilationDatabase` path, and creates a `compile_commands.json` stub so clangd starts without errors.
  - `fe add ols` writes or patches `ols.json` for the Odin Language Server.
  - `fe add zls` writes `zls.json` for the Zig Language Server.
  - `fe add gopls` writes `.gopls.yaml` workspace settings for Go.
- New `fe build` command — config-driven compiler wrapper. Define per-platform build commands in `.fe.yaml` under `build.targets`.
- New `fe release` command — orchestrates the full pipeline: `build → pack → bmp` in one shot, with `--skip-build`, `--skip-pack`, `--skip-bmp` flags.
- New `fe validate` command — pre-flight checker for config and paths, CI-safe (non-zero exit on failure).
- New `fe doctor` command — auto-detects project language and checks binaries, LSP configs, `compile_commands.json` validity, and cmake flags. Prints fix hints for every failed check.
- Shared `internal/ui` package providing spinners, step runners, styled output helpers (`OK`, `Fail`, `Warn`, `Dim`, `Title`, `Step`, `Hint`, `Elapsed`), and a `Confirm` prompt. Spinners degrade to plain log lines automatically in CI.
- Spinners on every command. All long-running operations show a Bubble Tea spinner in interactive terminals.
- Full Bubble Tea TUI (`fe tui`) with four tabs: Project, Ship, Doctor, Config. Supports keyboard navigation, inline project name prompts, per-command result views, and active config display. TTY-guarded — refuses to run in CI.
- Typed config struct in `internal/config` loaded via `viper.Unmarshal`. Config priority: env vars > `.fe.yaml` > conventions > defaults.
- CI-friendly env var overrides: `ITCHIO_USERNAME`, `ITCHIO_GAME`, `FE_USERVERSION`, `FE_BUILD_DIR`.
- Convention-based defaults — `fe` works with zero config if your project follows the standard layout (`builds/`, `assets/sprites/`, etc.).
- Remote template support — set `templates.source` in `.fe.yaml` to fetch `<lang>.tar.gz` from a custom URL, with automatic fallback to built-ins on failure.
- `fe-templates` submodule linked at [github.com/flushwhy/fe-templates](https://github.com/flushwhy/fe-templates).
- GitHub Actions workflows for building `fe` itself as a release binary across 5 platforms (linux-x64, linux-arm64, windows-x64, macos-x64, macos-arm64).
- Butler push now supports both flat convention layout (`builds/windows-x64/`) and legacy two-level layout (`builds/windows/x64/`), with `CombinedOutput()` so errors from butler are visible.

### Changed
- `fe init` now takes a language argument (`fe init odin --name mygame`) instead of just a project name.
- `fe bmp` now resolves platform channels from both flat and two-level directory structures.
- All commands use `RunE` (returning errors) instead of `Run` + `log.Fatal` for proper error propagation.
- `internal/transcode` ffmpeg-go arg keys corrected — no leading dash (e.g. `"c:v"` not `"-c:v"`).

### Fixed
- `pack.go` mutex-in-loop bug removed. File removal before write is now a clean single operation.
- `fs.Sub` embed path now uses `path.Join` (forward slashes) instead of `filepath.Join`, fixing template resolution on Windows.
- `compile_commands.json` stub written immediately on `fe init c/cpp` and `fe add clangd` so clangd never errors on startup before the first build.

---

## [0.0.3]

### Added
- Started work on a TUI.
- Added [fe-templates](https://github.com/flushwhy/fe-templates) submodule, to be integrated into the `init` command.

### Fixed
- Fixed file access permissions and added thread safety to the `pack` command.

---

## [0.0.2]

### Added
- Viper support for `.fe.yaml` config files.
- New `pack` command that packs multiple PNGs into a single PNG.
- New `init` command.

### Changed
- How `transcode` works; updated flags and options.

### Removed
- `resizetexture` and `PNGJoiner` commands.

### Fixed
- `transcode` and `bmp` commands.

---
