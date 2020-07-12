# Basic CI

This is a refactored version of github.com/sidletsky/tfilecaps/

## Usage
```bash
kek -repo=<owner/repo> -token=<your token here>
```

## Description
- runs locally
- runs using docker

## Flow
1. Verify `token`
2. Get config file `./kek/config.yml` from the specified repo
3. Pull the `runner_image`
4. Create the container with specified in config `runner_image`
5. Start container
6. Upload repo archive from github using `token` to the container
7. Run commands one by one in the directory of an archive
8. Output to the console container's output
    - if exit code is non-zero, cleanup and exit with provided exit code
9. Cleanup: stop and remove containers
