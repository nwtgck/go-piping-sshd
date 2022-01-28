# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)

## [Unreleased]

## [0.7.0] - 2022-01-28
### Added
* Support Windows partially (Port-forwarding and SFTP supported, pty not supported)

### Changed
* (breaking) Rename --allows-empty-password to --allow-empty-password because `--allow-` is more commonly used

## [0.6.0] - 2022-01-27
### Changed
* Allow no Content-Type
* Use `--allows-empty-password` to allow the empty password

## [0.5.0] - 2022-01-25
### Added
* Support "exec" request type
* Support SFTP

### Changed
* Use an embed private key, not random one
* Use 4096 as default buffer sizes in HTTP for better speed
* Update dependencies
* Add `,reuseaddr` to the socat command	hint

## [0.4.0] - 2021-03-27
### Changed
* Make the hint command shorter
* No auth when password is empty

## [0.3.1] - 2021-03-25
### Changed
* (internal) Use github.com/creack/pty instead of github.com/kr/pty because moved

## [0.3.0] - 2021-03-25
### Added
* Support "direct-tcpip", which is for port forwarding with -L or tunneling with -D

## [0.2.0] - 2021-03-25
### Added
* Add --shell option to specify shell

## 0.1.0 - 2021-03-25
### Added
* Initial release

[Unreleased]: https://github.com/nwtgck/go-piping-sshd/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/nwtgck/go-piping-sshd/compare/v0.6.0...0.7.0
[0.6.0]: https://github.com/nwtgck/go-piping-sshd/compare/v0.5.0...0.6.0
[0.5.0]: https://github.com/nwtgck/go-piping-sshd/compare/v0.4.0...0.5.0
[0.4.0]: https://github.com/nwtgck/go-piping-sshd/compare/v0.3.1...0.4.0
[0.3.1]: https://github.com/nwtgck/go-piping-sshd/compare/v0.3.0...0.3.1
[0.3.0]: https://github.com/nwtgck/go-piping-sshd/compare/v0.2.0...0.3.0
[0.2.0]: https://github.com/nwtgck/go-piping-sshd/compare/v0.1.0...0.2.0
