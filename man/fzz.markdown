FZZ 1
=======================================

NAME
----

fzz -- run commands interactively

SYNOPSIS
--------

`fzz` [command with `{{}}` placeholder]

DESCRIPTION
-----------

`foo` makes it possible to run command line utilities interactively, by
allowing to change the input passed to commands in an interactive interface.
The interface shows the current output of the specified command with the
current input. As soon as the input changes, the command is re-run and the
output updated.

The specified command MUST include the default placeholder `{{}}`

OPTIONS
-------

`-v`
  Print version and exit

EXAMPLES
--------

Run `grep` interactively:

    fzz grep {{}} *.txt

Run `find` interactively:

    fzz find . -iname "*{{}}*"

Run `grep` interactively and read from STDIN:

    cat hello_world.txt | fzz grep {{}}

Run `grep` interactively with a custom placeholder:

    FZZ_PLACEHOLDER=%% fzz grep %% *.txt

ENVIRONMET
-----------

`FZZ_PLACEHOLDER`
  Setting this variable changes the placeholder fzz uses from `{{}}` to the one
  specified.

AUTHOR
------

Thorsten Ball <mrnugget@gmail.com>
