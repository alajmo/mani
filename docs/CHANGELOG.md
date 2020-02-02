# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2017-06-20

### Added

- Add support for running from nested sub-directories
- Add info sub-command that shows which configuration file is being used
- Add flag to point to config file
- Accept different config names (.mani, .mani.yaml, .mani.yml, mani.yaml, mani.yml)
- Add new command exec to run arbitrary command
- Add config flag
- Add first argument to init should be path, if empty, current dir
- Add completion for all commands bash
- Update auto-discovery to equal true by default

### Fixed

- Fix crashing on not found config file
- Check possible, non-handled nil/err values
- Don't add project to gitignore if doesn't have a url
- Remove path if path is same as name
