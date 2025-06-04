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
git_branch() {
    # -- Finds and outputs the current branch name by parsing the list of
    #    all branches
    # -- Current branch is identified by an asterisk at the beginning
    # -- If not in a Git repository, error message goes to /dev/null and
    #    no output is produced
    git branch --no-color 2>/dev/null | sed -e '/^[^*]/d' -e 's/* \(.*\)/\1/'
}

git_status() {
    # Outputs a series of indicators based on the status of the
    # working directory:
    # + changes are staged and ready to commit
    # ! unstaged changes are present
    # ? untracked files are present
    # S changes have been stashed
    # P local commits need to be pushed to the remote
    local status="$(git status --porcelain 2>/dev/null)"
    local output=''
    [[ -n $(egrep '^[MADRC]' <<<"$status") ]] && output="$output+"
    [[ -n $(egrep '^.[MD]' <<<"$status") ]] && output="$output!"
    [[ -n $(egrep '^\?\?' <<<"$status") ]] && output="$output?"
    [[ -n $(git stash list) ]] && output="${output}S"
    [[ -n $(git log --branches --not --remotes) ]] && output="${output}P"
    [[ -n $output ]] && output="$output"  # separate from branch name
    echo $output
}

git_color() {
    # Receives output of git_status as argument; produces appropriate color
    # code based on status of working directory:
    # - White if everything is clean
    # - Green if all changes are staged
    # - Red if there are uncommitted changes with nothing staged
    # - Yellow if there are both staged and unstaged changes
    # - Blue if there are unpushed commits
    local staged=$([[ $1 =~ \+ ]] && echo yes)
    local dirty=$([[ $1 =~ [!\?] ]] && echo yes)
    local needs_push=$([[ $1 =~ P ]] && echo yes)
    if [[ -n $staged ]] && [[ -n $dirty ]]; then
        echo -e '\033[1;33m'  # bold yellow
    elif [[ -n $staged ]]; then
        echo -e '\033[1;35m'  # bold green
    elif [[ -n $dirty ]]; then
        echo -e '\033[1;31m'  # bold red
    elif [[ -n $needs_push ]]; then
        echo -e '\033[1;34m' # bold blue
    else
        echo -e '\033[1;37m'  # bold white
    fi
}

git_prompt() {
    # First, get the branch name...
    local branch=$(git_branch)
    # Empty output? Then we're not in a Git repository, so bypass the rest
    # of the function, producing no output
    if [[ -n $branch ]]; then
        local state=$(git_status)
        local color=$(git_color $state)
        # Now output the actual code to insert the branch and status
        echo -e "\x01$color\x02($branch$state)\x01\033[00m\x02"  # last bit resets color
    fi
}

# Sample prompt declaration. Tweak as you see fit, or just stick
# "$(git_prompt)" into your favorite prompt.
PS1='\[\e[0;35m\]${GITHUB_USER}\[\e[0m\]@\[\e[0;36m\]\u\[\e[0m\]: \w $(git_prompt)\[\033[00m\] \$ '

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
alias gcmsg='git commit -S -m'
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
if [ -f ".env" ]; then
  set -a
  . ".env"
  set +a
fi


### ─── Paths for Go, Turso, GCloud, etc. ───────────────────────────────────
export PATH=$PATH:/usr/local/go/bin
export PATH="$PATH:$HOME/.turso"

export DISPLAY=:1

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
      builtin cd "$@" && ls -a
}
export GPG_TTY=$(tty)
