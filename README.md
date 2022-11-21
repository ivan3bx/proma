# proma

```text
proma is a CLI tool for querying data via the Mastodon API.

Usage:
  proma [command]

Available Commands:
  auth        Authenticate with a Mastodon server.
  collect     Collects and aggregates tagged posts
  links       Extract links from any saved bookmarks
  help        Help about any command

Flags:
  -c, --config string     config file (default is $HOME/.proma.json)
  -h, --help              help for proma
  -s, --servers strings   server names to check (default [mastodon.social])
  -v, --verbose           verbose mode

Use "proma [command] --help" for more information about a command.
```

## Examples

### Extracting links from saved bookmarks

```bash
# creates a list of embedded links from bookmarked posts
./proma links --limit 2
```

```json
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

### Collecting posts by hashtag across multiple instances

```bash
# collects posts containing hashtags 'streetphotography' OR 'london'
./proma collect -t streetphotography,london -s mastodon.cloud,indieweb.social,social.linux.pizza
collecting from server: https://mastodon.cloud
collecting from server: https://indieweb.social
collecting from server: https://social.linux.pizza
...
```

```json
[
  {
    "uri": "https://photog.social/users/keirgravil/statuses/109377529017885305",
    "lang": "en",
    "content": "\u003cp\u003eCute little mushroom I spotted whilst out walking a section...",
    "tag_list": [
      "photography",
      "london",
      "autumn"
    ],
    "created_at": "2022-11-20T18:24:03Z"
  },
  {
    "uri": "https://mastodon.social/users/jesswade/statuses/109377446881326502",
    "lang": "en",
    "content": "...",
    "tag_list": [
      "london",
      "urbanphotography"
    ],
    "created_at": "2022-11-20T18:03:10Z"
  }
  ...
]
```
