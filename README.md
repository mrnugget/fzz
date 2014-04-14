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

### Use it in Vim!

Use it as interactive project search in vim

```
:set grepprg=fzz\ ag\ \{\{\}\}
```

## Installation

```
go get github.com/mrnugget/fzz
```


## TODO

* Get rid of the TODOs in the code
* Buffer the stdin of fzz and pass it to the specified command
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


