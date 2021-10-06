---
title: Pass
layout: 2017/sheet
category: CLI
---

# Topics

## Intro
[pass](https://passwordstore.org) is a command-line unix password maanger.

Passwords are stored in ~/.password-store

Summary:
* `pass init "password for storage"` : set up
* `pass` : list all passwords
* `pass insert twitter` : add password 

## Create

```sh
$ pass init [-p] <gpg-id>
$ pass git init
$ pass git remote add origin <your.git:repository>
$ pass git push -u --all
```

## Store

```sh
$ pass insert [-m] twitter.com/rsc
$ pass generate [-n] twitter.com/rsc length
```

## Retrieve

```sh
$ pass ls twitter.com/
$ pass show twitter.com/rsc
$ pass -c twitter.com/rsc
```

## Search

```sh
$ pass find twitter.com
```

## Management

```sh
$ pass mv twitter.com twitter.com/rsc
$ pass rm [-rf] twitter.com
$ pass cp twitter.com/rsc twitter.com/ricosc
```

```sh
$ pass edit twitter.com/rsc
```

## Synchronize

```sh
$ pass git push
$ pass git pull
```
