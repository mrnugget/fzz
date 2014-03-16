# NOT READY YET

# TODO

* Take commands like this:

    fzz "ag {{}} /my/dir"

  Where "{{}}" is a placeholder that should not be escaped by the shell
  evaluating the command. I am not sure about which placeholder to use yet. We
  have to see what's practical. First thought: `find` uses `{}` so that is no free
  to use, but would have been perfect.

* It then shows an empty screen with a input line at the top. 
* As soon as the user enters a character the specified command will be run and
  the output displayed in all the rows from 2-N. When running the command the
  placeholder {{}} will be replaced by the entered character.
* If the users enters more than one character, those will get added to the input
  and the command will be run again with the updated input.
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
