# ***Archival Notice***
This repository has been archived.

As a result all of its historical issues and PRs have been closed.

Please *do not clone* this repo without understanding the risk in doing so:
- It may have unaddressed security vulnerabilities
- It may have unaddressed bugs

<details>
   <summary>Click for historical readme</summary>

# `lotta_jaffles`

Welcome to the `lotta_jaffles` repo! This is available publicly...because it's a fork of `jaffle_shop`. It's not really intended for public consumption, though you're free to use it!

**Important**: work in progress.

## Why?

We lack a consistent set or standards around large dbt projects for testing purposes.

I've been finding Go fun to use in place of Python, which is normally how I'd do something like this. So this was primarily a personal project to use Go for something I'd normally do with Python.

However, I figured creating a large dbt project would be a decent use that we could build on.

## What?

The `v1` directory is `jaffle_shop` with a `main.go` script that copies all of the models N times. This implies all the models are exactly the same, with the same `ref`s throughout.

The `v2` directory moves the source code into the `main.go` script itself, no longer copying models. This allows us to edit the strings, replacing `ref`s and providing more complicated logic.

## Setup

TODO: basically install Go, install taskfile if desired, `go run .` -- this creates the `models/` directory. Edit the script as needed.

TODOs for this to be a real project:

- (P0) gather requirements for desired projects to test
- (P0) fix up the ergonomics (this `README.md`, Python setup, Go setup)
- (P1) add other dbt stuff (tests, YAMLs, etc.)
- (P1) a lot more...


