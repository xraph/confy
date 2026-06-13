# Changelog

All notable changes to this project will be documented in this file.

## [v0.5.2] - 2026-06-13

### Initial Release

- 53bcc5f (HEAD -> main, origin/main) fix(deps): bump golang.org/x/net to v0.55.0 (govulncheck)
- 7c23b5f fix: fixed go sec bug
- ca37aa1 fix: fixed case fold bug
- 726f281 feat: add documentation for variadic options and update environment variable handling
- c918b91 test: annotate expected DSNs with #nosec to indicate test credentials
- 2c16bab refactor: enhance config discovery to support multiple base and local config paths
- 616bd23 refactor: transition to NewFromConfig for instance creation
- 7e8bc26 feat: add exports.go for type aliasing and update go-utils dependency
- 7b22310 docs: add variable resolution section to README
- 58054b6 test: skip certain tests on Windows platform
- 5df0243 refactor: improve resource management in file operations
- e67b437 chore: update Go version in auto-release workflow to 1.25
- 8c00323 chore: update Go version in CI and release workflows to 1.25
- 4eee9de refactor: improve code clarity in GetWithOptions methods
- 1e7aef4 refactor: enhance file access security and improve error handling
- 4e68196 refactor: enhance type conversion and parsing logic
- 2e4d117 refactor: enhance test error handling and improve code clarity
- a4add44 refactor: improve error handling and code clarity in tests and sources
- 36be7d0 chore: update go-utils dependency to v0.0.5
- 71bb06d refactor: remove unnecessary lifecycle comments from confy.go
- 93e99fb refactor: remove deprecated compatibility aliases and update tests
- 4d27f01 refactor: rename ConfigManager to Confy and update related methods
- 9f87a6b feat: implement configuration watcher with support for change detection and event handling

