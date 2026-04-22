# Changelog

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
