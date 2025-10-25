---
name: 🐛 Bug Report
about: Report a bug to help us improve BubblyUI
title: '[BUG] '
labels: bug, needs-triage
assignees: ''
---

## 🐛 Bug Description

<!-- Provide a clear and concise description of the bug -->
<!-- What did you expect to happen vs. what actually happened? -->

## 🔍 To Reproduce

<!-- Provide detailed steps to reproduce the issue -->

### Steps to reproduce:
1. **Setup**: Describe your environment and setup
   ```bash
   # Example setup steps
   go mod init my-app
   go get github.com/yourusername/bubblyui
   ```

2. **Code**: Provide minimal, reproducible code
   ```go
   // Minimal code that reproduces the issue
   package main

   import "github.com/yourusername/bubblyui/pkg/bubbly"

   func main() {
       // Code that causes the issue
   }
   ```

3. **Commands**: Show the exact commands run
   ```bash
   # Commands that trigger the bug
   go run main.go
   ```

4. **Error**: What happens vs. what you expected
   ```
   Expected: [What you expected to happen]
   Actual: [What actually happened]
   ```

## 🎯 Expected Behavior

<!-- Describe what you expected to happen -->

## ❌ Actual Behavior

<!-- Describe what actually happened -->
<!-- Include error messages, stack traces, or unexpected output -->

## 🔧 Environment Details

<!-- Provide complete environment information -->

- **OS**: [e.g., Linux Ubuntu 22.04, macOS 14.1, Windows 11]
- **Architecture**: [e.g., amd64, arm64]
- **Go Version**: [e.g., `go version` output]
- **BubblyUI Version**: [e.g., v0.1.0, commit hash, or `main` branch]
- **Terminal/Shell**: [e.g., bash, zsh, Windows Terminal]
- **Additional Tools**: [e.g., any relevant tools or dependencies]

## 📊 Error Output

<!-- Paste complete error output, logs, or stack traces -->

```
[Paste error output here - include full stack traces if available]
```

## 🔬 Debugging Information

<!-- Any additional debugging information that might help -->

### Related Components
<!-- Which BubblyUI components or features are involved? -->
- [ ] Reactivity system (Ref[T], Computed[T])
- [ ] Component model (tea.Model implementation)
- [ ] Lifecycle hooks (onMounted, etc.)
- [ ] Directives (If, ForEach, Bind, On)
- [ ] Built-in components (Button, Input, etc.)
- [ ] Testing framework integration

### Attempted Solutions
<!-- What have you tried to fix this issue? -->

## 📋 Additional Context

<!-- Any other context about the problem -->

### Workaround
<!-- Is there a temporary workaround? -->

### Impact
<!-- How does this affect your usage? -->
- [ ] Blocking development completely
- [ ] Major functionality broken
- [ ] Minor inconvenience
- [ ] Cosmetic issue only

### Related Issues
<!-- Link to any related issues or discussions -->

---

<!-- Thank you for helping improve BubblyUI! 🛠️ -->
