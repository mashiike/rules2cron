# rules2cron

[![Documentation](https://godoc.org/github.com/mashiike/rules2cron?status.svg)](https://godoc.org/github.com/mashiike/rules2cron)
![Latest GitHub release](https://img.shields.io/github/release/mashiike/rules2cron.svg)
![Github Actions test](https://github.com/mashiike/rules2cron/workflows/Test/badge.svg?branch=main)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/mashiike/rules2cron/blob/main/LICENSE)

cron-like notation converter for ScheduleExpression in EventBridge's Rule

## Usage 

For example, use with [github.com/takumakanari/cronv](https://github.com/takumakanari/cronv)

```shell
$ rules2cron -tz JST -show-disabled | cronv -o ./my_event_bridge_schedule_rules.html -d 24h
```
### Install 
#### Homebrew (macOS and Linux)

```console
$ brew install mashiike/tap/rules2cron
```
#### Binary packages

[Releases](https://github.com/mashiike/rules2cron/releases)
## LICENSE

MIT License

Copyright (c) 2022 IKEDA Masashi
