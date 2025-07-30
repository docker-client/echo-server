# Why Enter is Required After Ctrl+D

The requirement to press Enter after Ctrl+D is due to how terminal input is handled at the operating system level, not in our Go code.

## Terminal Input Modes

In Unix-like systems (including Alpine Linux), terminals operate in "canonical mode" (cooked mode) by default where:
- Input is line-buffered by the terminal driver at the OS level
- Input isn't sent to the application until a line delimiter (Enter/CR/LF) is received
- Special characters like Ctrl+D are interpreted by the terminal driver

## The Actual Issue

When you press Ctrl+D in a terminal:
1. The terminal driver in canonical mode doesn't immediately send this to our application
2. It holds the input in its buffer until Enter is pressed
3. Only after Enter is pressed does our application receive the input, including the Ctrl+D character
4. Our ctrlDReader then detects the Ctrl+D and signals EOF

## Why bufio.Reader Doesn't Help

The bufio.Reader in our code provides buffering for data that has already been delivered to our application by the OS. It doesn't affect how or when the OS delivers that data to our application in the first place.

## Potential Solution

To make Ctrl+D work without requiring Enter, you would need to put the terminal in raw mode using terminal control libraries like `golang.org/x/term`. This would bypass the OS-level line buffering so every keystroke would be immediately sent to your application.

However, raw mode has other implications - it disables all terminal processing, including echo, so you'd need to handle that yourself if needed.
