# Importing

Rundown documents can be merged together via importing.

For example, it's you might have a `RUNDOWN.md` file in the root of a project, and then have specialised documents in another location, such as `docs/administration.md`.

To merge these together, the `RUNDOWN.md` file can import the other files into it, optionally prefixing that document's commands.

To import, use the `<r import>` element, wrapping a hyperlink to the document. For example, the following imports all commands in the administration document under the prefix of `admin`.

~~~ markdown

Our <r import="admin">[Administration Guide](./docs/administration.md)</r> contains common administration tasks.

~~~

Then, rundown help may look like this:

```
$ rundown --help
Rundown turns Markdown files into console scripts.

Usage:
  rundown [command] [flags]...
  rundown [command]

Available Commands:
  help             Help about any command
  main             A task within the main RUNDOWN.md file
  admin:bake       A task within the admin document.

Flags:
      --debug          Write debugging info to rundown.log
  -f, --file string    File to run (defaults to RUNDOWN.md then README.md)
  -h, --help           help for rundown
      --serve string   Set the port to serve a HTML interface for Rundown
  -v, --version        version for rundown

Use "rundown [command] --help" for more information about a command.
```

It's important to note that each command will run with it's working directory being the directory containing the command's file.

So in the above example, `main` will have the root as it's working directory, while `admin:bake` will have the `docs/` working directory.