# Quickstart

1. Install `est`
```
go get -u github.com/ryanberckmans/est
```

2. Integrate `est` into your prompt. We've designed this to be minimally distracting. Let us know how did!
```
est prompt
```

3. Enable `est` bash completion
```
est bash
```

4. Add your first task
```
est help add
```

5. Consider moving your `~/.estfile.toml` to a location with automatic backup, such as Dropbox or Google Drive. Set this location in `~/.estconfig.toml`.

# About `est`

`est` is a command-line tool to track time spent on tasks and predict the delivery date of future tasks.

In our experience, it's the case that for most time trackers "the juice is not worth the squeeze". `est` tries to improve on this by increasing "the juice" with auto predicted delivery dates, and decreasing "the squeeze" with auto time tracking and nice shell and prompt integration.

Auto predicted delivery dates is based on Joel Spolsky's [evidence-based scheduling](https://www.joelonsoftware.com/2007/10/26/evidence-based-scheduling/).
