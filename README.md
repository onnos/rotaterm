This should work on most recent Go versions:
`go get github.com/onnos/rotaterm`

Create an image object and render it to a terminal window using braille characters. Basically a thin wrapper around [gg](https://github.com/fogleman/gg), [tcell](https://github.com/gdamore/tcell) and [dotmatrix](https://github.com/kevin-cantwell/dotmatrix).

Everything in one file for convenience, this is just a base for further experimentation. The basic idea is to use image generation algos to feed to a braille printer, animating a sequence based on simple time-based incremental counters.

Key modifiers are displayed on-screen, additionally:
- CTRL-L: restarts with sane defaults
- ESC/Enter: exits
- Arrow keys: move X/Y coordinates

The name stands for "Rotate Terminal" or something. It doesn't `rm` anything. That didn't really occur to us, dude.

