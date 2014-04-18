# fzz

**Do one thing, do it well — multiple times!**

**fzz** allows you to change the input of a single command interactively. Have a
look and pay close attention to the bee's knees here:

![fzz-gif-cast](http://recordit.co/FCnvkoyAKV.gif)

## Usage

The general usage pattern is this:

```bash
fzz [command with {{}} placeholder]
```

Examples:

```
fzz grep {{}} *.go
fzz ag {{}}
fzz find . -iname "*{{}}*"
```

Run this, change the input and **fzz** runs the command again, this time with the
new input and gives you what the command wrote to STDOUT.

Press **return** and **fzz** will print the last output to its own STDOUT.
That means it works with pipes too:

```
fzz grep {{}} *.go | head -n 1 | awk -F":" '{print $1}'
```

### Use it in Vim!

Use it as interactive project search in vim

```
:set grepprg=fzz\ ag\ \{\{\}\}
```

Then use `:grep` in Vim to start it. **fzz** will then fill the quickfix window
with its results.

### Interactively search through files with ag and open them in your favorite editor

Put this in your shell config and configure it to use your favorite editor:

```bash
vimfzz() {
  vim $(fzz ag {{}} | awk -F":" '{print $1}' | uniq)
}
```

## Installation

```
go get github.com/mrnugget/fzz
```


## TODO

* Man page
* Get rid of the TODOs in the code
* Add the ability to specify a first input value:
  * `{{}}` doesn't specify any input and waits for the user
  * `{{foobar}}` tells fzz to run the specified command with the initial input
    of `foobar`
* Change how fzz is printing its output to the TTY: instead of clearing the
  screen, jump to the first line, right after the input and rewrite the visible
  lines (padded so that all the columns are used and the previous output is not
  visible anymore)
* Maybe add some keybindings:
  * Ctrl-W: delete the last word from the input and rerun the command
  * Maybe selection of a line with Ctrl-J and Ctrl-K and the only print the
    selected line.
* Do not run anything when no {{}} is given in command.


