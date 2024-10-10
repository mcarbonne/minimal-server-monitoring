## unreleased
### Breaking Changes
- `threshold_percent` replaced by `threshold` in `filesystemusage` provider ([details](README.md#filesystemusage)).
### New Features
- `filesystemusage` introduced a new metric to detect rapid changes in available disk spaces (both increase and decrease)
### Fixes
- `mountpoint_whitelist` in `filesystemusage` provider is now working as expected

## 2.0.0
### Breaking Changes
- `config.json` is now  `config.yml`
- `scrape_interval` (for scrapers) is now a string with unit. Before, it was an integer (seconds). Example: `scrape_interval: 120` is now `scrape_interval: 120s` (or  even `scrape_interval: 2m`).
### New Features
- `filesystemusage` provider has been reworked to allow automatic mountpoints detection. See [here](#filesystemusage) for details.
