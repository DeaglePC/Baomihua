# Output setup
Output demo.gif

# Visual Settings
Require bmh
Set FontSize 18
Set FontFamily "Cascadia Code"
Set Width 1200
Set Height 600
Set TypingSpeed 100ms
Set Theme "Dracula"
Set Padding 30
Set WindowBar Colorful
Set MarginFill "#6B50FF"
Set Margin 20

# Start Recording
Sleep 1s

# Type out the prompt naturally
Type "?? 查看公网IP"
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

# 找出占用内存最多的5个进程
Type "bmh 找出占用内存最多的5个进程"
Sleep 500ms
Enter
Sleep 4s
Enter
Sleep 3s

Type "clear"
Enter
Sleep 1s

# 找出CPU占用最多的5个进程
Type "bmh 找出CPU占用最多的5个进程"
Sleep 500ms
Enter
Sleep 4s
Enter
Sleep 3s

Type "clear"
Enter
Sleep 1s

# 查找*.log中包含cat的行
Type "bmh 查找*.log中包含cat的行"
Sleep 500ms
Enter
Sleep 4s
Enter
Sleep 3s
