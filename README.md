# fzz

**Do one thing, do it well — multiple times!**

![fzz-gif-cast](http://recordit.co/FCnvkoyAKV.gif)

# TODO

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


# Use it in Vim!

Use it as interactive project search in vim

```
:set grepprg=fzz\ ag\ \{\{\}\}
```
