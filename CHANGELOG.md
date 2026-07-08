# Changelog

## [0.8.0](https://github.com/aaronflorey/pupdate/compare/v0.7.0...v0.8.0) (2026-07-08)


### Features

* add reset command ([ba21f1b](https://github.com/aaronflorey/pupdate/commit/ba21f1bbc05f361de15cec7676a192c6b03c36b5))
* add run dry-run mode ([77b5809](https://github.com/aaronflorey/pupdate/commit/77b58093747ccdc9fed7b8c1d617096a07e2f948))
* cli error visibility ([2f3da4d](https://github.com/aaronflorey/pupdate/commit/2f3da4dd204c9f2c501c444ae71be52b44f75721))
* **cli:** add project status diagnostics ([07243bb](https://github.com/aaronflorey/pupdate/commit/07243bb0e53ec39544ffd52f431440b64a735db6))
* **cli:** Quiet no-op runs and drop JSON output ([ff80520](https://github.com/aaronflorey/pupdate/commit/ff8052078e453d04533baff54c44593769124e9b))
* **config:** Add root directory scoping and config inspect command ([10e3d52](https://github.com/aaronflorey/pupdate/commit/10e3d5259bf8935f4cae0f7fb34f9e8f91ce61ae))
* **config:** add run default controls ([cff2653](https://github.com/aaronflorey/pupdate/commit/cff26539ad02cd6b870bbfc8511e2373ef29bd1b))
* **config:** Support root directory lists and config bootstrap ([24edab1](https://github.com/aaronflorey/pupdate/commit/24edab1637056c23cc7b6e2fe127ca100414a535))
* **deps:** Integrate git-pkgs package tooling ([a5e412e](https://github.com/aaronflorey/pupdate/commit/a5e412e070d02f34ac9a41334818f843066bc680))
* **detection:** Add depth-1 subfolder support ([32e869f](https://github.com/aaronflorey/pupdate/commit/32e869f524105104bf06beb5518e9090be63be29))
* **detection:** Add Kasetto support ([841bead](https://github.com/aaronflorey/pupdate/commit/841bead7fd6b824be75dba53264e5af33a23852b))
* **detection:** Scan packages children and honor gitignore ([28dd28f](https://github.com/aaronflorey/pupdate/commit/28dd28f077f88431316038ba952945be0f22d414))
* **hook:** add opt-in async shell mode ([237ce3b](https://github.com/aaronflorey/pupdate/commit/237ce3b6717565a81d3095e964983828914342fb))
* initial commit ([6c5dc75](https://github.com/aaronflorey/pupdate/commit/6c5dc7545227fbe2b24b1c15986e3b41feacb4eb))
* **P1-R1:** review hook and platform contract updates ([6439d9a](https://github.com/aaronflorey/pupdate/commit/6439d9aad9dee06ea95e9293902e42c1f484e8b5))
* **P1-T1:** make init async by default ([274f566](https://github.com/aaronflorey/pupdate/commit/274f566fc82ac65f0b98c6e114a8fe0f113f87f1))
* **P1-T2:** align shipped platform support docs ([282d1af](https://github.com/aaronflorey/pupdate/commit/282d1af0f98633753e31995a99e7352755a9c6c9))
* **P2-R1:** review shared preflight and status guidance ([4010a1e](https://github.com/aaronflorey/pupdate/commit/4010a1e8002b97e6181b0ce9c8303426345066ff))
* **P2-T1:** extract shared preflight collection ([8125281](https://github.com/aaronflorey/pupdate/commit/8125281a80144213aae2e2b61b6a7b952d2b27ba))
* **P2-T2:** add status remediation guidance ([69e6be1](https://github.com/aaronflorey/pupdate/commit/69e6be17f12c9eb18498a5e0df6317e717d95869))
* **P3-R1:** review monorepo scan expansion ([c8038d8](https://github.com/aaronflorey/pupdate/commit/c8038d8ec5ae378a994faddfa9395308a7ed6ff9))
* **P3-T1:** define workspace_globs config support ([9639f17](https://github.com/aaronflorey/pupdate/commit/9639f17ce47ae8552da59756fe5113a935eb027d))
* **P3-T2:** apply workspace_globs during detection ([c37f900](https://github.com/aaronflorey/pupdate/commit/c37f900327e636704a71d9972e47116141ca207a))
* **P3-T3:** document workspace_globs behavior ([7392d0d](https://github.com/aaronflorey/pupdate/commit/7392d0d2190e7f666e2b891131aad1254bc84fbc))
* **P3-T4:** add folder blacklist config support ([e6029f7](https://github.com/aaronflorey/pupdate/commit/e6029f7062fe4baa3b66c8235cc1042fca1e420d))
* **P3-T5:** apply folder blacklist across detection ([96f75e5](https://github.com/aaronflorey/pupdate/commit/96f75e5fd487871217798445a09cc804c55007fc))
* **P3-T6:** document folder blacklist behavior ([e36f744](https://github.com/aaronflorey/pupdate/commit/e36f74429b29d5a302b3edb26e312e3f2fa4c8a9))
* **P3:** add folder blacklist planning ([dff8c3d](https://github.com/aaronflorey/pupdate/commit/dff8c3d870a0540e38b70caa453f7290ddf9f305))
* support PUPDATE_CONFIG override ([4d5b9f3](https://github.com/aaronflorey/pupdate/commit/4d5b9f3773ff2737c6192b9cad76a24b9fa104cd))


### Bug Fixes

* change syscall to platform specific ([569ef48](https://github.com/aaronflorey/pupdate/commit/569ef48f579f50db2d372d7f7189c08fa785a5b4))
* **ci:** Remove Windows and normalize macOS paths ([5b0a560](https://github.com/aaronflorey/pupdate/commit/5b0a560b4e29c154eb340f3f8f19d58bc4c39f27))
* cli silence ([82e8465](https://github.com/aaronflorey/pupdate/commit/82e8465fb7e607b3e35c3aa942ee7c4693767d91))
* **config:** Match root directories case-insensitively ([cfd1e2c](https://github.com/aaronflorey/pupdate/commit/cfd1e2c352ea54e63ec436d308faa0e222670b79))
* **config:** respect root directory casing ([7df2bd5](https://github.com/aaronflorey/pupdate/commit/7df2bd5a262136542408d887dc2120fdef332f23))
* **config:** stop auto-creating user config ([15953af](https://github.com/aaronflorey/pupdate/commit/15953afd54bec857afaa1406105885a76819c253))
* **config:** Use ~/.config path on macOS ([15c3747](https://github.com/aaronflorey/pupdate/commit/15c374785f85aaeaa4379aaaf3bbe8d7e884ca7f))
* **detection:** preserve canonical Cargo.lock casing ([a729662](https://github.com/aaronflorey/pupdate/commit/a729662b13321834a88e144e8181277e02ab3057))
* **detection:** preserve matched lockfile paths ([5510774](https://github.com/aaronflorey/pupdate/commit/55107745d430be041671e2861b34b59ee251a516))
* **freshness:** always hash lockfiles ([ccdda5d](https://github.com/aaronflorey/pupdate/commit/ccdda5d2c627c312398fd5720ad97305d8dcf6c7))
* **freshness:** bound git submodule status checks ([e8827fa](https://github.com/aaronflorey/pupdate/commit/e8827faf55707024d33f8bc7d799491f662ae70e))
* **freshness:** safely reuse unchanged lockfile hashes ([19f0def](https://github.com/aaronflorey/pupdate/commit/19f0def5c24b3e901b5e217c137b9b1e42bf6dad))
* **hook:** keep async overlap locks live ([7941bbb](https://github.com/aaronflorey/pupdate/commit/7941bbbf0c1fc21b5ed0fdc8ab423db89e6af5f6))
* **kasetto:** correct subproject config path and close audit drift ([a0d944e](https://github.com/aaronflorey/pupdate/commit/a0d944ef99576e9752532d7b39d840eca238c663))
* **kasetto:** require local config for sync ([dcd3c10](https://github.com/aaronflorey/pupdate/commit/dcd3c10ec10d0b30a53cfb54df3ec5aba9c9fcf5))
* **kasetto:** scope sync to detected project ([f603877](https://github.com/aaronflorey/pupdate/commit/f6038771267a512502e336195f4180ed66098750))
* macos testing ([c4ed3d9](https://github.com/aaronflorey/pupdate/commit/c4ed3d9ac9d3f64df5e3bf81f48af9f68680ece7))
* **P4-R1:** review installation tooling and cli confidence phase ([2417aa8](https://github.com/aaronflorey/pupdate/commit/2417aa81e629011e339580c8c3dd38cb3e4b694a))
* **P4-T2:** pin mise go version ([77fae62](https://github.com/aaronflorey/pupdate/commit/77fae62e8ab183b70b6948b9431f6ac1258b9657))
* persist post-install lockfile state ([5769ef7](https://github.com/aaronflorey/pupdate/commit/5769ef712d41eb1cb5331084fb8c7b7dbc0387e7))
* python command  error wording ([7be5e08](https://github.com/aaronflorey/pupdate/commit/7be5e089a6a925aa5f34ef5665e57bc13bbb355a))
* release policy drift ([40a30c0](https://github.com/aaronflorey/pupdate/commit/40a30c0f0460e3d99ff60955509491b3e44d8ade))
* **release:** keep one release workflow path ([99f62be](https://github.com/aaronflorey/pupdate/commit/99f62beccaa2990a7a56bc05cd894260ce7a21c0))
* replace go-git ignore matcher ([3fea4c4](https://github.com/aaronflorey/pupdate/commit/3fea4c46a440f107ad67d8405cf12cbd361d81be))
* **run:** Skip all phases when .pupignore exists ([f1ae2c9](https://github.com/aaronflorey/pupdate/commit/f1ae2c97c5ebd3676a477cdbaf997f983c414fb3))
* skip unchanged composer lockfiles ([43dfcaa](https://github.com/aaronflorey/pupdate/commit/43dfcaadc36333c19d852dd2122bb56bb192e5f0))
* **state:** fsync parent directory on save ([1a67c6d](https://github.com/aaronflorey/pupdate/commit/1a67c6db8e6baed8e486292f86a80e30f5e333e0))
* **state:** prune stale run entries ([96623aa](https://github.com/aaronflorey/pupdate/commit/96623aa845059a9fc4080f45b0fdcbef6ca8bbd5))
* **test:** Align config path assertion with resolver ([c227daf](https://github.com/aaronflorey/pupdate/commit/c227daf912d208120427e41906572cc1c0fcc4a2))
* **test:** Normalize resolved config path assertion ([b2e0298](https://github.com/aaronflorey/pupdate/commit/b2e02982b152a65d2cacac33633f3f6435268d3c))
* **test:** Stabilize root directory resolution on macOS ([95d3dd0](https://github.com/aaronflorey/pupdate/commit/95d3dd037bca79668c3845b161c796cb864aa715))
* use absolute hook executable path ([4b51ecc](https://github.com/aaronflorey/pupdate/commit/4b51ecc9172ec08cb042c819c74a64e89bcd46a8))


### Performance Improvements

* **freshness:** cache lockfile metadata ([85771e3](https://github.com/aaronflorey/pupdate/commit/85771e3fededbb71bbcb48fd0cef067044855fd3))

## [0.7.0](https://github.com/aaronflorey/pupdate/compare/v0.6.2...v0.7.0) (2026-05-28)


### Features

* **cli:** add project status diagnostics ([121f3dd](https://github.com/aaronflorey/pupdate/commit/121f3dde2d23bd6547cff45e91113f850e6cc3ed))
* **cli:** Quiet no-op runs and drop JSON output ([3e5a207](https://github.com/aaronflorey/pupdate/commit/3e5a2077056e999ddf4d842ebdc656643d5840b9))
* **config:** Add root directory scoping and config inspect command ([abb273b](https://github.com/aaronflorey/pupdate/commit/abb273b2a8751b1d330aa59458353b3c90638fad))
* **config:** add run default controls ([d28efac](https://github.com/aaronflorey/pupdate/commit/d28efac79294a6c3ca37e7289db2c49e478ed173))
* **config:** Support root directory lists and config bootstrap ([949840d](https://github.com/aaronflorey/pupdate/commit/949840d078f1bc464136d9439604b9a16a524b6e))
* **deps:** Integrate git-pkgs package tooling ([d4d4b24](https://github.com/aaronflorey/pupdate/commit/d4d4b2460b3a7561247eb678a2f8cdecdfabb435))
* **detection:** Add depth-1 subfolder support ([3d50a3d](https://github.com/aaronflorey/pupdate/commit/3d50a3de045bbd6f5288b7d3a8d41a0479a6869c))
* **detection:** Add Kasetto support ([c4372f8](https://github.com/aaronflorey/pupdate/commit/c4372f8b4f46767b1fe83316d977a60450e848af))
* **detection:** Scan packages children and honor gitignore ([84dc635](https://github.com/aaronflorey/pupdate/commit/84dc635f9e085763bb421194019efb48d0c9a951))
* **hook:** add opt-in async shell mode ([1da12ad](https://github.com/aaronflorey/pupdate/commit/1da12ad5becaa6954df0849b58b5e7b577491347))
* initial commit ([6c5dc75](https://github.com/aaronflorey/pupdate/commit/6c5dc7545227fbe2b24b1c15986e3b41feacb4eb))


### Bug Fixes

* change syscall to platform specific ([1abb368](https://github.com/aaronflorey/pupdate/commit/1abb3685739b7a13962b8f38a45b0cdde43d9113))
* **ci:** Remove Windows and normalize macOS paths ([b9b8ea4](https://github.com/aaronflorey/pupdate/commit/b9b8ea4e782cf3e8ea5350788e4a02b21adc2e5e))
* **config:** Match root directories case-insensitively ([452e4ea](https://github.com/aaronflorey/pupdate/commit/452e4ea4dbdf4dd0841f5168fe68a6ae53ea8a4a))
* **config:** respect root directory casing ([ea52f04](https://github.com/aaronflorey/pupdate/commit/ea52f0497451104022d01a4a5c0008e74e3fd62d))
* **config:** stop auto-creating user config ([ff4b2e3](https://github.com/aaronflorey/pupdate/commit/ff4b2e3134f07f9f6eaed791da4171b8a3598bcf))
* **config:** Use ~/.config path on macOS ([2e73f6c](https://github.com/aaronflorey/pupdate/commit/2e73f6c44156efe0a20e4d283de709d4183bb316))
* **detection:** preserve canonical Cargo.lock casing ([ce9035b](https://github.com/aaronflorey/pupdate/commit/ce9035bde25759075a731af5dac33047b9ff2b08))
* **detection:** preserve matched lockfile paths ([74b8786](https://github.com/aaronflorey/pupdate/commit/74b878651fd4b4152c39170cb58ce772b528669c))
* **freshness:** always hash lockfiles ([4004abd](https://github.com/aaronflorey/pupdate/commit/4004abdcac8fc0b36541cc1bce37102b13c7a681))
* **freshness:** bound git submodule status checks ([52a25b4](https://github.com/aaronflorey/pupdate/commit/52a25b48cf0a36466635954e7b06fa9be792652e))
* **freshness:** safely reuse unchanged lockfile hashes ([2d8fb08](https://github.com/aaronflorey/pupdate/commit/2d8fb081482de922a74ed2e68fc5f3a1bd5f5e01))
* **hook:** keep async overlap locks live ([6e6ee66](https://github.com/aaronflorey/pupdate/commit/6e6ee668636ecf21c8200d979680be76d55c5452))
* **kasetto:** correct subproject config path and close audit drift ([295bafe](https://github.com/aaronflorey/pupdate/commit/295bafe975b8e74451e0b5946e6f697697c181e2))
* **kasetto:** require local config for sync ([09155f3](https://github.com/aaronflorey/pupdate/commit/09155f31daa05d81314942c9a8175487d808957c))
* **kasetto:** scope sync to detected project ([a14c7d0](https://github.com/aaronflorey/pupdate/commit/a14c7d08b5bee5863f739048a5bcf2dc31da609e))
* persist post-install lockfile state ([b43dbe5](https://github.com/aaronflorey/pupdate/commit/b43dbe53dc918ff11138835efd990a731262f9e4))
* **release:** keep one release workflow path ([6910b25](https://github.com/aaronflorey/pupdate/commit/6910b25e4e692f0927be573a15702166cf032fe7))
* **run:** Skip all phases when .pupignore exists ([6a1de36](https://github.com/aaronflorey/pupdate/commit/6a1de36b5bce19e82c82ffbc0e118c488be36ed0))
* skip unchanged composer lockfiles ([366b8cc](https://github.com/aaronflorey/pupdate/commit/366b8cc02557b99ca54cb254b3950b698fb6b279))
* **state:** fsync parent directory on save ([c6c3144](https://github.com/aaronflorey/pupdate/commit/c6c3144a55d611c808f08c8f120961544cc26c48))
* **state:** prune stale run entries ([e72cd24](https://github.com/aaronflorey/pupdate/commit/e72cd24dd56962143a22a4011a2bf042b6766aa2))
* **test:** Align config path assertion with resolver ([f5e832a](https://github.com/aaronflorey/pupdate/commit/f5e832a54da83525b16053527a7a3635bb76e8be))
* **test:** Normalize resolved config path assertion ([ce9583c](https://github.com/aaronflorey/pupdate/commit/ce9583c93e49dda47b3e4294a7723ea6b838c4bf))
* **test:** Stabilize root directory resolution on macOS ([4e89b84](https://github.com/aaronflorey/pupdate/commit/4e89b842a01e896346ce21f1b24a48ae2925173d))


### Performance Improvements

* **freshness:** cache lockfile metadata ([269fd80](https://github.com/aaronflorey/pupdate/commit/269fd80485f2f3efd9f4e69861b391ea7e8c9471))

## [0.6.2](https://github.com/aaronflorey/pupdate/compare/v0.6.1...v0.6.2) (2026-05-10)


### Bug Fixes

* persist post-install lockfile state ([e03dfe0](https://github.com/aaronflorey/pupdate/commit/e03dfe0771dc52abfa31c07c0cf2114294a285e1))

## [0.6.1](https://github.com/aaronflorey/pupdate/compare/v0.6.0...v0.6.1) (2026-05-04)


### Bug Fixes

* change syscall to platform specific ([08a9ec3](https://github.com/aaronflorey/pupdate/commit/08a9ec38aac8cfa1708607e49c0a195ac642e782))

## [0.6.0](https://github.com/aaronflorey/pupdate/compare/v0.5.0...v0.6.0) (2026-05-04)


### Features

* **cli:** add project status diagnostics ([1a70825](https://github.com/aaronflorey/pupdate/commit/1a70825fe46884ffca07ab70a56978834b74828a))
* **config:** add run default controls ([a82d79d](https://github.com/aaronflorey/pupdate/commit/a82d79dd5d429968a041f640c7e325f3a0c7881f))
* **detection:** Add Kasetto support ([67ed5e7](https://github.com/aaronflorey/pupdate/commit/67ed5e70e8100c9906b152987eff8904adca67b7))
* **hook:** add opt-in async shell mode ([3a93903](https://github.com/aaronflorey/pupdate/commit/3a93903757d329206f31399ec5699d7d8fbda3c2))


### Bug Fixes

* **config:** Match root directories case-insensitively ([34bbd36](https://github.com/aaronflorey/pupdate/commit/34bbd36a846339eac1a0e86f7c8e643d5e868174))
* **config:** respect root directory casing ([eea69d6](https://github.com/aaronflorey/pupdate/commit/eea69d6ee55bf47fc0e61a71f4109872256b9245))
* **config:** stop auto-creating user config ([12406a9](https://github.com/aaronflorey/pupdate/commit/12406a950b5a54c845e8bebeb6749a20989fc93b))
* **detection:** preserve canonical Cargo.lock casing ([e419124](https://github.com/aaronflorey/pupdate/commit/e4191243dccaf333ad6fe7e60cff44f90b8c511d))
* **detection:** preserve matched lockfile paths ([a0f0f59](https://github.com/aaronflorey/pupdate/commit/a0f0f59b071f83436b8364c63dc673d4adf04e47))
* **freshness:** always hash lockfiles ([c1c7049](https://github.com/aaronflorey/pupdate/commit/c1c70493b12248ec05e2b47b30ecba4ade967a8c))
* **freshness:** bound git submodule status checks ([44a0775](https://github.com/aaronflorey/pupdate/commit/44a077596edcc6ceed548c73d467edccec7435af))
* **freshness:** safely reuse unchanged lockfile hashes ([b7e893a](https://github.com/aaronflorey/pupdate/commit/b7e893ac443f3d450152e879d84216176d792837))
* **hook:** keep async overlap locks live ([38f7187](https://github.com/aaronflorey/pupdate/commit/38f7187eefe5b12528e7ce12dfe190b5e7cea3ae))
* **kasetto:** correct subproject config path and close audit drift ([57c0364](https://github.com/aaronflorey/pupdate/commit/57c03648c889cf3f78fdad97b2dfe1edf7d8d3fb))
* **kasetto:** require local config for sync ([e19f928](https://github.com/aaronflorey/pupdate/commit/e19f9280f7d53043fc667eba86714da5f0fcbe95))
* **kasetto:** scope sync to detected project ([702f006](https://github.com/aaronflorey/pupdate/commit/702f006f84cb3936d8a5a9ba7d2578de04eec6f9))
* **release:** keep one release workflow path ([ede7c28](https://github.com/aaronflorey/pupdate/commit/ede7c2889348a798a24c2b2180e3be5dc14d7ddd))
* **state:** fsync parent directory on save ([97be85e](https://github.com/aaronflorey/pupdate/commit/97be85eee9758954f11222edcac8509d9d791b7c))
* **state:** prune stale run entries ([97373c7](https://github.com/aaronflorey/pupdate/commit/97373c776f7bdf479c8df30888e088084f45d5dc))


### Performance Improvements

* **freshness:** cache lockfile metadata ([c2753d9](https://github.com/aaronflorey/pupdate/commit/c2753d948c1a49dc05a4e27044b512e4db06c74e))

## [0.5.0](https://github.com/aaronflorey/pupdate/compare/v0.4.0...v0.5.0) (2026-04-22)


### Features

* **config:** Support root directory lists and config bootstrap ([5a74999](https://github.com/aaronflorey/pupdate/commit/5a749990da7e7be58d03f69db80e6c13ec2dba69))


### Bug Fixes

* **test:** Stabilize root directory resolution on macOS ([91d77db](https://github.com/aaronflorey/pupdate/commit/91d77dbe41bf9cb06f72d96543e224cf6913127f))

## [0.4.0](https://github.com/aaronflorey/pupdate/compare/v0.3.1...v0.4.0) (2026-04-22)


### Features

* **cli:** Quiet no-op runs and drop JSON output ([e0eb310](https://github.com/aaronflorey/pupdate/commit/e0eb310c64ba4bf8b001c02e3e9a5a71467e70b9))
* **config:** Add root directory scoping and config inspect command ([799aaf7](https://github.com/aaronflorey/pupdate/commit/799aaf74245120a35888203dfd871ed02252b934))
* **deps:** Integrate git-pkgs package tooling ([22223ef](https://github.com/aaronflorey/pupdate/commit/22223ef1690932ff58b5a9514538d80378ea41a9))


### Bug Fixes

* **ci:** Remove Windows and normalize macOS paths ([b68dcac](https://github.com/aaronflorey/pupdate/commit/b68dcac508d9cfa7c626c38ab44bf601499a4e15))
* **config:** Use ~/.config path on macOS ([ce87058](https://github.com/aaronflorey/pupdate/commit/ce870584272c3bd4979d554783b30038a7e072ca))
* **run:** Skip all phases when .pupignore exists ([f0efeca](https://github.com/aaronflorey/pupdate/commit/f0efecaae64193178eafed9287c4eee82585fae7))
* **test:** Align config path assertion with resolver ([420f63b](https://github.com/aaronflorey/pupdate/commit/420f63bf7ac14324b3e58a64d5ec92fa3f2c93cd))
* **test:** Normalize resolved config path assertion ([0b45c58](https://github.com/aaronflorey/pupdate/commit/0b45c58e40ca3bb828df4380153b30696b1c69a7))

## [0.3.1](https://github.com/aaronflorey/pupdate/compare/v0.3.0...v0.3.1) (2026-04-14)


### Bug Fixes

* bump ([bb41d16](https://github.com/aaronflorey/pupdate/commit/bb41d16f0782be1b15f3e1bb690f567fea1de137))

## [0.3.0](https://github.com/aaronflorey/pupdate/compare/v0.2.0...v0.3.0) (2026-04-13)


### Features

* **detection:** Add depth-1 subfolder support ([770666a](https://github.com/aaronflorey/pupdate/commit/770666aa252fe5844ef06dad829539cf65249b40))
* **detection:** Scan packages children and honor gitignore ([6a11317](https://github.com/aaronflorey/pupdate/commit/6a11317001a11f66d1aff97345527ffc44ebec3c))

## [0.2.0](https://github.com/aaronflorey/pupdate/compare/v0.1.0...v0.2.0) (2026-04-12)


### Features

* initial commit ([0b5d7c2](https://github.com/aaronflorey/pupdate/commit/0b5d7c2a8ee0160ac3077e4a3381d73e16ba76e0))
