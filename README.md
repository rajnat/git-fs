# git-fs

git-fs is a tool for maintaining a Git-backed encrypted cloud storage of your files. It watches a local directory, encrypts the files found there, and commits them to a Git repository, allowing you to securely version and back up your data.

## Features
* Encryption: Uses AES-GCM to encrypt files before committing to a Git repository.
* Version Control: Automatically commits changes so you can roll back to previous versions.
* Daemon Mode: Runs in the background, continuously watching for changes.
* Configurable: Uses Viper for flexible configuration from files, environment variables, and CLI flags.
* User-Friendly CLI: Uses Cobra to provide a clean command-line interface with multiple subcommands.

## How It Works

* #### Initialization:
    Set up the repository and generate a salt. A derived key from a provided password encrypts all files.

* #### Watching and Encrypting:
    The daemon monitors a specified directory for file changes. When a file changes, it’s encrypted and committed to the .encrypted directory within your repo.

* #### Decryption:
    On another machine, clone the repository, run git-fs decrypt, and provide the same password to decrypt all files.

## Requirements

    Go 1.18+ (or a recent stable release)
    Git installed on your system
    A working Git repository (local or remote)

## Installation

Clone the repository:

```
git clone https://github.com/rajnat/git-fs.git
cd git-fs
```
Initialize Go module and dependencies:
```
go mod tidy
```
Build the binary:
```
go build -o git-fs ./cmd
```
This produces a git-fs binary in the current directory.

(Optional) Move the binary to a location in your PATH:
```
    mv git-fs /usr/local/bin/
```
Configuration

git-fs uses Viper, allowing configuration via:    config.yaml file:
    Place a config.yaml in the same directory you run git-fs. Example:

password: "mysecretpassword"
repo_path: "./myrepo"
watch_path: "./watched_directory"
remote_url: "<path to git repo to store encrypted files>"

Environment Variables:
Prefix environment variables with GITFS_. For example:

    export GITFS_PASSWORD="mypassword"
    export GITFS_REPO_PATH="/path/to/myrepo"
    export GITFS_WATCH_PATH="/path/to/watch"

    Command-line Flags:
    For supported commands, you can use flags like --config config.yaml.

Note: The password should not be stored in plain text in a public repository. Use environment variables or a secure prompt method.
Usage

git-fs [command]

Commands

    git-fs init
    Initializes the repository by generating a salt file (.salt) and verifying the encryption key is derivable.

git-fs init

git-fs daemon
Starts the background watcher. This will:

    Watch the watch_path directory.
    On changes, encrypt files into .encrypted.
    Run git add and git commit automatically. Optionally push changes if remote_url is set.

git-fs daemon

git-fs decrypt
Decrypts all files from the .encrypted directory into their original plaintext form, using the provided password.

git-fs decrypt

git-fs version
Shows the current version of git-fs.

    git-fs version

Examples

Initialize a Repository:

export GITFS_PASSWORD="mysupersecret"
export GITFS_REPO_PATH="./my-encrypted-repo"
export GITFS_WATCH_PATH="./my-data"

git-fs init

Start the Daemon:

git-fs daemon

This will monitor ./my-data and commit encrypted changes to ./my-encrypted-repo.

Decrypt on Another Machine:

    Clone the repo:

git clone git@github.com:username/my-encrypted-repo.git
cd my-encrypted-repo

Run decrypt:

    export GITFS_PASSWORD="mysupersecret"
    git-fs decrypt

    Now the .encrypted files are decrypted back into their original structure.

### Security Considerations

    Password Management:
    Don’t store your password in version control. Use environment variables, secure prompts, or a password manager.

    Key Rotation:
    Implement a process to rotate keys if needed. Currently, the project relies on a stable password-derived key.

### Contributing

Contributions are welcome! Please open issues or pull requests on GitHub.