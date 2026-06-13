# Versioning and Releases

`ycontext` versions use this format:

```text
YY.MM.DD.NN
```

Approved releases add `-Release`:

```text
26.06.11.01
26.06.11.01-Release
```

## Development Versions

A version without `-Release` is a development version. Development versions may be used for testing, but they should not be treated as stable releases.

The final numeric segment, `NN`, is the daily sequence number. It allows multiple development versions on the same date.

## Release Versions

A release version is a tested and approved development version with `-Release` appended.

Releases are expected to follow this flow:

1. Continue active development under `YY.MM.DD.NN`.
2. Pause feature work at a selected version.
3. Test that version.
4. Apply focused fixes as needed.
5. Approve the version for release.
6. Publish the same version with `-Release`.
7. Resume wider development.

## Compatibility

Before v1, public APIs, storage schemas, and protocol methods may change. Release notes should call out any migration or compatibility impact once persistent storage is implemented.
