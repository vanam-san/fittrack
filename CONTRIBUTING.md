# Contributing to FitTrack

Please follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## Report a bug or propose a feature

Search [GitHub Issues](https://github.com/vanam-san/fittrack/issues) for related work. For bugs, include the FitTrack version, operating system, install method, reproduction steps, expected result, and actual result. Use sample data and remove personal information from logs and screenshots. For features, describe the workflow you want to improve.

## Set up and verify a change

1. Fork the repository, clone your fork, and create a branch from `master`.
2. Follow the [development setup](README.md#development) and run `wails3 dev` to exercise changes in the desktop app.
3. Run the relevant [Go and frontend checks](README.md#checks). Changes to storage, imports, or installers should also be tested against existing data or packages using the [release checks](docs/release-checks.md).
4. Open a pull request against `master`. Include what changed, why, and how you verified it. Include screenshots for visible UI changes and call out changes to storage or package compatibility.

FitTrack is offline-first. Exercise data, journal entries, and body measurements should remain usable without network access. See the [architecture guide](docs/architecture.md) for where to make changes. Contributions are accepted under the [MIT License](LICENSE).
