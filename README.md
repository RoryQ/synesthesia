# Synesthesia
### Sense your projects as iTerm2 tab colours

Synesthesia updates your iTerm2 tab colours depending on the go module name in your directory ancestry.

Supports tinting the background with `--background-tint` and worktrees by default

# Installation
### 1. Install from source using golang 1.25

```
go install github.com/roryq/synesthesia@latest
```

### 2. Then configure a hook for your shell.
### fish
Add the following line to your `~/.config/fish/config.fish`:
```fish
synesthesia hook fish --background-tint | source
```

### zsh
Add the following line to your `~/.zshrc`
```zsh
eval "$(synesthesia hook zsh --background-tint)"
```

# Usage
Navigate between your directories as usual. When you have multiple tabs open for different go projects, 
a consistent random colour will be chosen for any tabs with the same go module name.

If you use git worktrees (or jj workspaces), then each copy of the same project will have a different but consistent colour.

![](demo.gif)

# License
[MIT](LICENSE)
