# ~/.bashrc: executed by bash(1) for non-login shells.

### ─── If not running interactively, don’t do anything ───────────────────────
case $- in
  *i*) ;;
    *) return;;
esac

### ─── History Settings ─────────────────────────────────────────────────────
HISTCONTROL=ignoreboth:erasedups
HISTIGNORE="ls:cd:cd -:pwd:exit:date:* --help"
PROMPT_COMMAND="history -a; $PROMPT_COMMAND"
shopt -s histappend
HISTTIMEFORMAT="%F %T "    # optional: prepend timestamps
HISTSIZE=1000
HISTFILESIZE=2000
shopt -s checkwinsize      # update LINES/COLUMNS after each command

### ─── Prompt (colors + Git status) ─────────────────────────────────────────
git_status() {
  local porcelain out
  porcelain="$(git status --porcelain 2>/dev/null)" || return
  out=""
  [[ -n $(egrep '^[MADRC]' <<<"$porcelain") ]] && out="${out}+"
  [[ -n $(egrep '^.[MD]'   <<<"$porcelain") ]] && out="${out}!"
  [[ -n $(egrep '^\?\?'   <<<"$porcelain") ]] && out="${out}?"
  [[ -n $(git stash list)            ]] && out="${out}S"
  [[ -n $(git log --branches --not --remotes 2>/dev/null) ]] && out="${out}P"

  if [[ $out == *P* ]]; then
    out="${out//\?/}"
  fi

  [[ -n $out ]] && echo "$out"
}

# 1) Pick a 24‑bit ANSI color based on the status flags string (e.g. "+", "+!", "S", or "").
#    Returns a bare "\033[38;2;R;G;Bm" sequence.
# ──────────────────────────────────────────────────────────────────────────
git_color() {
  local st="$1" has_staged has_dirty has_push

  [[ $st =~ \+     ]] && has_staged=yes    # “+” means staged
  [[ $st =~ [!\?]  ]] && has_dirty=yes     # “!” or “?” means dirty/untracked
  [[ $st =~ P      ]] && has_push=yes      # “P” means push pending

  if [[ -n $has_staged && -n $has_dirty ]]; then
    echo "\033[38;2;255;255;0m"   # yellow
  elif [[ -n $has_staged ]]; then
    echo "\033[38;2;0;255;0m"     # green
  elif [[ -n $has_dirty ]]; then
    echo "\033[38;2;255;0;0m"     # red
  elif [[ -n $has_push ]]; then
    echo "\033[38;2;255;0;255m"   # magenta
  else
    echo "\033[38;2;255;255;255m" # white
  fi
}

# ──────────────────────────────────────────────────────────────────────────
# 2) Grab the current branch name; empty if not in a repo
# ──────────────────────────────────────────────────────────────────────────
git_branch() {
  git rev-parse --abbrev-ref HEAD 2>/dev/null
}

# ──────────────────────────────────────────────────────────────────────────
# 3) Build “(branch<status>)” in one shot, with ALL raw \033[…] escapes wrapped
#    in \[ \] so Bash treats them as zero‐width.  No extra “()” ever.
#
#    Even when clean (status=""), we still show “(branch)” in white.
# ──────────────────────────────────────────────────────────────────────────
git_prompt() {
  if ! git rev-parse --is-inside-work-tree &>/dev/null; then
    return
  fi

  local b st col
  b=$(git_branch) || return
  st=$(git_status)

  if [[ -z $st ]]; then
    col="\033[38;2;255;255;255m"  # white when clean
  else
    col=$(git_color "$st")
  fi

  # Emit: \[<col>\](<branch><status>)\[\033[0m\]
  echo -e "\[${col}\](${b}${st})\[\033[0m\]"
}

GIT_USER_COLOR='\033[38;2;0;200;0m'   # a mild green for username
GIT_HOST_COLOR='\033[38;2;0;255;255m' # cyan for hostname
RESET_COLOR='\033[0m'

PS1="\[$GIT_USER_COLOR\]\${GITHUB_USER}\[$RESET_COLOR\]@\
\[$GIT_HOST_COLOR\]\h\[$RESET_COLOR\]: \w $(git_prompt) $ "

export PROMPT_DIRTRIM=4

### ─── Aliases & Safety ────────────────────────────────────────────────────
alias ll='ls -alF'
alias la='ls -A'
alias l='ls -CF'

# safer file operations
alias rm='rm -i'
alias mv='mv -i'
alias cp='cp -i'

# colorized grep/diff if available
if [ -x /usr/bin/dircolors ]; then
  eval "$(dircolors -b)"
  alias ls='ls --color=auto'
  alias diff='diff --color=auto'
  alias grep='grep --color=auto'
  alias fgrep='fgrep --color=auto'
  alias egrep='egrep --color=auto'
fi

# Git shortcuts
alias gs='git status'
alias ga='git add'
alias gcmsg='git commit -m'
alias gp='git push'
alias gl='git pull --rebase'
alias gd='git diff'
alias gco='git checkout'
alias gcb='git checkout -b'
alias glog='git log --oneline'

# quick notification for long‑running jobs (if notify‑send is installed)
if command -v notify-send &>/dev/null; then
  alias alert='notify-send --urgency=low -i \
    "$([ $? = 0 ] && echo terminal || echo error)" \
    "$(history|tail -n1|sed -e '\''s/^\s*[0-9]\+\s*//;s/[;&|]\s*alert$//'\'')"'
fi

### ─── vi-mode (optional) ───────────────────────────────────────────────────
# Uncomment to use vi‑style keybindings at the prompt:
# set -o vi

### ─── Environment (.env) ──────────────────────────────────────────────────
if [ -f "$HOME/.env" ]; then
  set -a
  . "$HOME/.env"
  set +a
fi

### ─── Paths for Go, Turso, GCloud, etc. ───────────────────────────────────
export PATH=$PATH:/usr/local/go/bin
export PATH="$PATH:$HOME/.turso"

export DISPLAY=:1
export GPG_TTY=$(tty)

# Google Cloud SDK
if [ -f "/usr/lib/google-cloud-sdk/path.bash.inc" ]; then
  . /usr/lib/google-cloud-sdk/path.bash.inc
fi
if [ -f "/usr/lib/google-cloud-sdk/completion.bash.inc" ]; then
  . /usr/lib/google-cloud-sdk/completion.bash.inc
fi

# Bash completion (loads all “/usr/share/bash-completion/completions/*” etc.)
if ! shopt -oq posix; then
  if [ -f /usr/share/bash-completion/bash_completion ]; then
    . /usr/share/bash-completion/bash_completion
  elif [ -f /etc/bash_completion ]; then
    . /etc/bash_completion
  fi
fi

### ─── Handy functions ──────────────────────────────────────────────────────
# Auto‑listing on cd:
cd() {
  if [ "$1" = ".." ]; then
    prev="$PWD"
    builtin cd .. && ls "$prev"
  else
    builtin cd "$@" && ls
  fi
}
