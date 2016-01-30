# Command Parsing

This package has all the sauce for command parsing.  Some guidelines follow.

- `init()` methods for subcommands can error out if something is wrong, and the command will immediately exit.
- The `Run()` function of a command should not `log.Fatal()`, `os.Exit()`, or `panic()`, as it will bypass cleanup.