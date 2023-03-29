# Slug

[![Go Reference](https://pkg.go.dev/badge/github.com/dotse/slug.svg)](https://pkg.go.dev/github.com/dotse/slug)
![GitHub](https://img.shields.io/github/license/dotse/slug?style=flat-square)

![](./_doc/img/slug.svg)

Slug is a [`slog.Handler`] that writes human-readable logs.

⚠ [`slog`] isn’t finalised yet. Slug is even more unstable.

![](./_doc/img/screenshot.png)

## Features

-   Colours!

-   Non-printable characters are escaped

## Non-Features

-   Performance

    Being fast or low on resources is not a goal

-   Predictable syntax

    The output is meant for human eyes only, not to be parsed

[`slog`]: https://pkg.go.dev/golang.org/x/exp/slog
[`slog.Handler`]: https://pkg.go.dev/golang.org/x/exp/slog#Handler
