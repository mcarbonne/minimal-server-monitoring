### unreleased
- `threshold_percent` replaced by `threshold` in `filesystemusage` provider ([details](README.md#filesystemusage)).

### 2.0.0
- `config.json` is now  `config.yml`
- `scrape_interval` (for scrapers) is now a string with unit. Before, it was an integer (seconds). Example: `scrape_interval: 120` is now `scrape_interval: 120s` (or  even `scrape_interval: 2m`).
- `filesystemusage` provider has been reworked to allow automatic mountpoints detection. See [here](#filesystemusage) for details.