# Changelog

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
