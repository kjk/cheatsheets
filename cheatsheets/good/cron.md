---
title: cron
category: CLI
layout: 2017/sheet
---

## Main

Crontab is a Unix tool to schedule running a command at a given interval.<br>
https://crontab.guru/ helps to understand cron syntax.

Common tasks:
* `echo "@reboot echo hi" | crontab` : add task that will run `echo hi` command on every reboot:
* `crontab -e` : open cron config in an editor:
* `crontab -l [-u user]` : list tasks

### Format

Cron task consists of:
* a command to execute
* a pattern defining when the command will execute

```
Min  Hour Day  Mon  Weekday
```
{: .-setup}

```
*    *    *    *    *  command to be executed
```

```
┬    ┬    ┬    ┬    ┬
│    │    │    │    └─  Weekday  (0=Sun .. 6=Sat)
│    │    │    └──────  Month    (1..12)
│    │    └───────────  Day      (1..31)
│    └────────────────  Hour     (0..23)
└─────────────────────  Minute   (0..59)
```
{: .-setup.-box-chars}

### Operators

| Operator | Description                |
| ---      | ---                        |
| `*`      | all values                 |
| `,`      | separate individual values |
| `-`      | a range of values          |
| `/`      | divide a value into steps  |

### Examples

| Example        | Description                 |
| ---            | ---                         |
| `0 * * * *`    | every hour                  |
| `0 8 * * *`    | every day at 8 am           |
| `*/15 * * * *` | every 15 mins               |
| `0 */2 * * *`  | every 2 hours               |
| `0 18 * * 0-6` | every week Mon-Sat at 6pm   |
| `10 2 * * 6,7` | every Sat and Thu on 2:10am |
| `0 0 * * 0`    | every Sunday midnight       |
| ---            | ---                         |
| `@reboot`      | every reboot                |
