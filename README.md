![Version](https://img.shields.io/badge/version-0.0.0-orange.svg)

# put.io cli client

If you are [put.io](https://put.io) user, you’ll like this mini command-line
tool.

You need a put.io token. Create your token [here](https://app.put.io/settings/account/oauth/apps)

---

## Installation

You can install from the source;

```bash
$ go get -u github.com/vigo/putio-cli
```

or, you can install from `brew`: (*not ready yet...*)

```bash
$ brew tap vigo/putio-cli
$ brew install putio-cli
```

---

## Usage

You can set your put.io token via;

1. Setting `PUTIO_TOKEN` environment variable
1. or, use `-t`, `-token` cli-flag

```bash

usage: putio-cli [-flags] [subcommand] [args]

  You can set your token via PUTIO_TOKEN environment variable! unless
  you need to pass token via -t or -token.

flags:

  -help, -h                           display help
  -version, -v                        display version (0.0.0)
  -color, -c                          enable/disable color (default: disabeld)
  -token, -t                          set put.io token

subcommands:

  list FOLDERID                       list files under given FOLDERID (default: 0 which is root folder)
  list -delete FILEID FOLDERID        first delete given FILEID then list files under given FOLDERID
  list -id FOLDERID                   list files under given FOLDERID as file id
  list -url FOLDERID                  list files under given FOLDERID as downloadable URL

  upload url URL URL URL...           tell put.io to download given URL(s)

  delete FILEID FILEID...             delete given FILEIDs
  move FILEID FILEID... FOLDERID      move given files (FILEIDs) to target folder (FOLDERID)

examples:

  $ putio-cli -t YOURTOKEN list
  $ putio-cli -t YOURTOKEN -c list

  # list files under given FOLDERID in color!
  $ putio-cli -t YOURTOKEN -c list FOLDERID

  # first delete given FILEID then list files for given FOLDERID in color!
  $ putio-cli -t YOURTOKEN -c list -delete FILEID FOLDERID

  # tell putio to upload given urls
  $ putio-cli -t YOURTOKEN upload url https://www.youtube.com/watch?v=eeiTP69qc9Q https://www.youtube.com/watch?v=SrBUaaNsZzg

  # tell putio to move given FILEIDs to given FOLDERID in color!
  $ putio-cli -t YOURTOKEN -c move FILEID FILEID FILEID... FOLDERID

  # delete files in color!
  $ putio-cli -t YOURTOKEN -c delete FILEID FILEID

  # move files in color!
  $ putio-cli -t YOURTOKEN -c move FILEID FILEID... FOLDERID

```

---

## Rake Tasks

```bash
$ rake -T

rake default               # show avaliable tasks (default task)
rake docker:build          # Build
rake docker:rmi            # Delete image
rake docker:run            # Run
rake release[revision]     # Release new version major,minor,patch, default: patch
rake test:run[verbose]     # run tests, generate coverage
rake test:show_coverage    # show coverage after running tests
rake test:update_coverage  # update coverage value in README
```

---

## Docker

build:

```bash
$ docker build . -t putio-cli
```

run:

```bash
$ docker run -i -t putio-cli:latest putio-cli -h
$ docker run -i -t putio-cli:latest putio-cli -t "${PUTIO_TOKEN}" list
```

---

## Contributer(s)

* [Uğur "vigo" Özyılmazel](https://github.com/vigo) - Creator, maintainer

---

## Contribute

All PR’s are welcome!

1. `fork` (https://github.com/vigo/putio-cli/fork)
1. Create your `branch` (`git checkout -b my-feature`)
1. `commit` yours (`git commit -am 'add some functionality'`)
1. `push` your `branch` (`git push origin my-feature`)
1. Than create a new **Pull Request**!

This project is intended to be a safe, welcoming space for collaboration, and
contributors are expected to adhere to the [code of conduct][coc].

---

## License

This project is licensed under MIT

[coc]: https://github.com/vigo/putio-cli/blob/main/CODE_OF_CONDUCT.md