# Synesthesia

Synesthesia updates your iTerm2 tab colours depending on the go module name in your directory ancestry.

# Installation
### 1. Install from source using golang 1.16

```
go install github.com/roryq/synesthesia@vlatest
```

### 2. Then configure a hook for your shell.
### fish
Add the following line to your `~/.config/fish/config.fish`:
```fish
synesthesia hook fish | source
```

### zsh
Add the following line to your `~/.zshrc`
```zsh
eval "$(synesthesia hook zsh)"
```

# Usage
Navigate between your directories as usual. When you have multiple tabs open for different go projects, 
a consistent random colour will be chosen for any tabs with the same go module name.

# License
[MIT](LICENSE)
