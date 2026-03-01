# Output setup
Output demo.gif

# Visual Settings
Require bmh
Set FontSize 15
Set FontFamily "Monaco, PingFang SC"
Set LetterSpacing 0
Set Width 1200
Set Height 600
Set TypingSpeed 100ms
Set Theme "Dracula"
Set Padding 30
Set WindowBar Colorful
Set MarginFill "#6B50FF"
Set Margin 20

Hide
Type 'eval "$(bmh --init $(basename $SHELL))"'
Enter
Type "clear"
Enter
Show
Sleep 1s

# Type out the prompt naturally
Type "?? Check my public IP"
Sleep 500ms
Enter

# Wait for the AI generation spinner to finish
Sleep 4s

# The UI shows the analysis and options, browse through the menu
Down
Sleep 1s
Up
Sleep 1s
Enter
Sleep 3s

Type "clear"
Enter
Sleep 1s

Type "?? show my local ip addr"
Sleep 500ms
Enter
Sleep 4s
Enter
Sleep 3s

Type "clear"
Enter
Sleep 1s

Type "?? find max file in this folder"
Sleep 500ms
Enter
Sleep 4s
Enter
Sleep 3s

Type "clear"
Enter
Sleep 1s
