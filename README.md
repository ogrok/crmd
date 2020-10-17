# crmd
crmd is a simple cli reminders tool that handles timestamps properly. You specify your reminder, a simple recurrence schedule, and a date with optional time, and crmd will store it nicely in JSON and surface it to you when called after the specified time.

I wrote this tool to append to new terminal sessions, i.e. add `crmd` to the end of `.bashrc` etc.

## installation
- compile it locally with a proper go toolchain: `go build`
- move resulting binary anywhere in PATH
- for best effect, add `crmd` to the end of `.bashrc` or equivalent

## features
- quickly add reminders to list kept in JSON:
    `crmd -d 2020-01-01 -t 07:00 -r yearly "it's the new year!"`
- call without arguments to check for due reminders:
    `crmd`
- get reminded repeatedly until you acknowledge / confirm:
    `crmd -c 1`
- delete reminder:
    `crmd -x 1`
- list all reminders:
    `crmd -a`
- doesn't result in an ever-growing archive you'll never look at
- properly schedules recurrence: corrects for daylight savings, uses local time zone, avoids due date drift in monthly, annual scenarios, etc. The only case of note is that reminders on the 29th through 31st of a month can drift to the 1st of the following month, and each 1st thereafter; this is fine for my use cases, so I didn't address it.

## limitations
- recurrence is limited to `daily`, `weekly`, `monthly`, `quarterly`, `yearly` and will never schedule a new instance of a recurring reminder in the past; always the next time relative to the current time
- no tags, contexts, projects, priorities etc. as this is not designed to replace Taskwarrior etc.

## contributing
Like most of what I create, I wrote this tool for my own personal use, and I have an interest in keeping the project dead-simple. I welcome issues and PRs that keep this in mind; otherwise, you could fork or, frankly, rewrite this project from scratch in very little time.
