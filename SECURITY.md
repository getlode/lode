# Security Policy

## Reporting a vulnerability

Please report security issues **privately**, not via public issues.

- Preferred: open a [GitHub private vulnerability report](https://github.com/getlode/lode/security/advisories/new).
- Or email **j.s.torchia@gmail.com** with details and steps to reproduce.

You can expect an initial acknowledgement within a few business days. Please
allow reasonable time for a fix before public disclosure.

## Supported versions

lode is pre-1.0. Security fixes are applied to the latest released version.
Until 1.0, only the most recent tag is supported.

## Scope notes

- lode handles your data and credentials only locally and against the remote you
  configure; it does not phone home.
- Remote credentials are read from the environment / your DVC config / the AWS
  credential chain — lode does not store them.
