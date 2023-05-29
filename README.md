# pfind

Parallel find.

Usage: `pfind [-d ROOT] [-n] GLOB`

`ROOT` defaults to the current directory (`.`).

The `-n` flag appends `:1:` at the end of each output. It makes the outputs look
like those from `grep`.

Example:

```
pfind -d ~ 'emacs*'
```
