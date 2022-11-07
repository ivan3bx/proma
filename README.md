# proma

```text
CLI tool for extracting data from the Mastodon API.

Usage:
  proma [command]

Available Commands:
  auth        Authenticate with a Mastodon server.
  links       Extract links from any saved bookmarks
  help        Help about any command

Flags:
  -c, --config string   config file (default is $HOME/.proma.json)
  -h, --help            help for proma
  -s, --server string   server name (default "mastodon.social")
  -v, --verbose         verbose mode

Use "proma [command] --help" for more information about a command.
```

## Example: extracting links from saved bookmarks

```json
$ proma links --limit 2
[
  {
    "profileName": "megmaker",
    "profileURL": "https://mastodon.social/@megmaker",
    "URL": "https://mastodon.social/@megmaker/109291500135871673",
    "linkRef": "https://www.pcmag.com/how-to/how-to-get-started-on-mastodon-and-leave-twitter-behind"
  },
  {
    "profileName": "mykalmachon",
    "profileURL": "https://indieweb.social/@mykalmachon",
    "URL": "https://indieweb.social/@mykalmachon/109299730701767558",
    "linkRef": "https://mykal.codes/posts/why-every-company-is-a-tech-company/"
  }
]
```