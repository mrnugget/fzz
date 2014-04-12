# fzz

**Do one thing, do it well — multiple times!**

![fzz-gif-cast](http://recordit.co/FCnvkoyAKV.gif)

# TODO

* If the user enters a new line, the selected line will be printed and the
  program terminates
* There will be special key combinations and keys that will not get added to the
  input but trigger special behaviour. Those will be:
    * Ctrl-C: quit and do nothing
    * Ctrl-D: stop accepting input and print the currently selected line
    * Return/New-line: stop accepting input and print the currently selected
      line
    * Ctrl-J: move the selection down
    * Ctrl-K: move the selection up
    * Backspace: delete the last character from the input and rerun the command
    * Ctrl-W: delete the last word from the input and rerun the command


# Tricks

Use it as interactive project search in vim

```
:set grepprg=fzz\ ag\ \{\{\}\}
```
