<p align="center">
    <img src="https://raw.githubusercontent.com/1pkg/gomer/master/gopher.png?nc=1" alt="gomer"/>
</p>

# Gomer: cli tool for go module index search

[![lint](https://github.com/1pkg/gomer/workflows/lint/badge.svg)](https://github.com/1pkg/gomer/actions?query=workflow%3Alint+branch%3Amaster+)
[![build](https://github.com/1pkg/gomer/workflows/build/badge.svg)](https://github.com/1pkg/gomer/actions?query=workflow%3Abuild+branch%3Amaster+)
[![report](https://goreportcard.com/badge/github.com/1pkg/gomer?nocache)](https://goreportcard.com/report/github.com/1pkg/gomer)
[![version](https://img.shields.io/github/go-mod/go-version/1pkg/gomer?nocache)](https://github.com/1pkg/gomer/blob/master/go.mod)
[![license](https://img.shields.io/github/license/1pkg/gomer?nocache)](LICENSE)

`go get -u github.com/1pkg/gomer`

## Introduction

Gomer is a cli tool that can quicly search and cache go module [index](https://index.golang.org/) database to provide the abilitty to match modules based on regular expression for mobule path or [semver constraint](github.com/Masterminds/semver/v3) for module version.

## Usage

```
gomer -h
Usage of gomer: [-f val] [-f val] <module_path_regexp>
  -constraint string
        cli semver constraint pattern; used only if non empty valid constraint specified
  -format string
        cli printf format for printing a module entry; \n is auto appended (default "%s %s %s")
  -index string
        cli go module index database url; golang.org index is used by default (default "https://index.golang.org/index")
  -nocache
        cli modules no caching flag; caching is enabled by default (default false)
  -timeout int
        cli timeout in seconds; only considered when value bigger than 0 (default 0)
  -verbose
        cli verbosity logging flag; verbosity is disabled by default (default false)
```

e.g.

```
gomer -constraint="<1.2" "gopium$"
github.com/1pkg/gopium v1.1.2 2020-06-18T19:45:44.710921Z
github.com/1pkg/gopium v1.1.1 2020-06-18T19:45:30.350059Z
github.com/1pkg/gopium v1.1.0 2020-06-17T20:51:55.723138Z
github.com/1pkg/gopium v1.0.1 2020-06-17T20:28:53.968057Z
github.com/1pkg/gopium v1.0.0 2020-06-13T13:44:46.041278Z
```

## Licence

Gomer is licensed under the MIT License.  
See [LICENSE](LICENSE) for the full license text.
