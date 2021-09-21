---
title: Tig
category: Git
layout: 2017/sheet
updated: 2018-11-16
weight: -3
---

## Main

### Intro

[Tig](https://jonas.github.io/tig/) is a text-mode interface for Git.

To install:

```bash
# MacOS + Homebrew
$ brew install tig --HEAD
```

```bash
# Ubuntu
$ sudo apt install tig
```

Not available for windows.

**Invocation**:

| Command             | Description                                  |
| ------------------- | -------------------------------------------- |
| `tig`               |                                              |
| ---                 | ---                                          |
| `tig status`        | Status                                       |
| `tig blame FILE`    | Blame                                        |
| ---                 | ---                                          |
| `tig master`        | Show a branch                                |
| `tig test..master`  | Show difference between two branches         |
| `tig FILE`          | Show history of file                         |
| `tig v0.0.3:README` | Show contents of file in a specific revision |
| `tig -C /repo/path` | Run in dir `/repo/path` (like `git -C`)      |

You can substitute `git log` → `tig`.

## Shortcut keys
{: .-three-column}

### Switching views

| Key | Command |
| --- | ------- |
| `m` | Main view |
| `s` | Status |
| `t` | Tree (files) |
| `y` | Stash view |
| `g` | Grep |
| `h` | Help |
{: .-shortcuts}

### All views

| Shortcut | Description |
| --- | --- |
| `j`  `k` | Up/down                 |
| `J`  `K` | Next/previous           |
| ---      | ---                     |
| `<`      | Back                    |
| `R`      | Refresh                 |
| `q`      | Close                   |
| `Q`      | Close all               |
| ---      | ---                     |
| `^N`     | Next on parent view     |
| `^P`     | Previous on parent view |
{: .-shortcuts}

### `m` - Main view

| Shortcut | Description |
| --- | --- |
| `D`      | Toggle date display modes   |
| `A`      | Toggle author display modes |
| `X`      | Toggle commit sha           |
| `C`      | Cherry pick a commit        |
{: .-shortcuts}

### `s` - Status view

| Shortcut | Description |
| --- | --- |
| `u`     | Stage/unstage file or chunk        |
| `!`     | Revert file or chunk               |
| `C`     | Commit                             |
| `M`     | Merge                              |
| `1`     | Stage line                         |
| `[` `]` | Increase/decrease the diff context |
{: .-shortcuts}

### `h` - Branch view

| Shortcut | Description |
| --- | --- |
| `i` | Change sort header |
{: .-shortcuts}

### `h` - Blame view

| Shortcut | Description |
| --- | --- |
| `,` | Parent commit |
{: .-shortcuts}
