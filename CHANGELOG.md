# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased]


### Added

- Initialize go module with dependencies
- Add config package with TOML loading and multi-server profiles
- Add service container for truenas-go interfaces
- Add TabBar widget with wrap-around navigation
- Add PoolsView with list display and data loading
- Add DatasetsView with list display and data loading
- Add SnapshotsView with list display, selection, and hold indicator
- Add root App widget with tab navigation and view switching
- Add main entrypoint with config loading and client setup
- Make SSH config optional with smart defaults and path expansion
- Make SSH optional and auto-detect host key fingerprint
- Add ViewLoaded event type for async view load notifications
- Add shared loading state renderer for views
- Add staleness tracking and loading state to all views
- Load all views in parallel with stale-aware tab switching and 'r' refresh key
- Add async connection lifecycle with connecting/failed states to App
- Add Table, BarGauge, and Sparkline widgets
- Extend Services to include System, Reporting, Interfaces, and Apps
- Add DashboardView with realtime streaming and subscription support
- Integrate Dashboard as first tab with streaming updates

### Build

- Update truenas-go to v0.2.3
- Upgrade truenas-go to v0.2.4

### Documentation

- Add installation, configuration, and usage to README
- Update SSH config docs to reflect optional status and renamed field
- Simplify configuration example and update keybindings in README
- Add dashboard screenshot and update keybindings reference

### Fixed

- Reload active view on tab switch and simplify key handling

### Testing

- Fill coverage gaps to meet 80%+ target
- Add Loaded, Stale, and loading draw state tests for all views
- Add LoadAll, ViewLoaded handling, stale refetch, and refresh key tests

