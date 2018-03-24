<p align="center"><img src="https://y.zxq.co/jobeei.png"></p>

# CheeseGull [![Build Status](https://travis-ci.org/osuripple/cheesegull.svg?branch=master)](https://travis-ci.org/osuripple/cheesegull)

CheeseGull creates an unofficial "slave" database of the official osu! beatmap
database, trying to keep every beatmap up to date as much as possible, as well
as a cache middleman for osz files.

The main purpose for this is, as you can see from the owner of the repository,
for running a sort of "beatmap mirror" for Ripple's osu! direct. We originally
used an actual osu! beatmap mirror which had all of the beatmaps on osu!
downloaded, but it ended up taking wayyyy too much space, and the cheapest
server we could find that had at least 2 TB of HDD had an upload speed of
30mbit - as you can guess, this meant that for the Ripple users who didn't have
a third world connection the download speed was pretty poor.

CheeseGull tries to hold a replica of the osu! database as updated as possible.
Though of course, not having any way to see the latest updated beatmaps, or have
a system for subscribing to updates to beatmap, or anything else which could
help us identify what has been updated recently makes it very hard to do keep
an updated copy at all times (Takeaway: the osu! API is completely shit). In
order to do this, CheeseGull updates WIP, Pending or Qualified beatmaps when
at least 30 minutes have passed since the time they were checked, whereas for
all other beatmaps (including Graveyard, Ranked, Approved, etc) at least 4 days
must have passed. This is not a problem for ranked/approved (it's highly
unlikely for a ranked beatmap to ever change state, and Graveyard beatmaps
are rarely resurrected, so there's that).

Beatmap downloads are also provided by going at `/d/<id>`. In case the beatmap
is not stored in the local cache, the beatmap will be downloaded on-the-fly
(this assumes the machine's internet connection is fast enough to download a
beatmap before a HTTP timeout happens). In case the beatmap already is in the
cache, then well, as you can imagine, it is served straight from there. Oh, yes,
multiple people downloading a not cached beatmap at the same time is a case we
handle. Or should be able to handle, at least.

## [API docs](http://docs.ripple.moe/docs/cheesegull/cheesegull-api)

## Getting Started

You can find binaries of the latest release
[here.](https://github.com/osuripple/cheesegull/releases/latest)

If you want to compile from source, if you have Go installed it should only be
a `go get github.com/osuripple/cheesegull` away.

The only requirements at the moment are a MySQL server and an osu! account.
Check out `cheesegull --help` to see how you can set them up for cheesegull to
work properly.

## Contributing

No strict contribution guide at the moment. Just fork, clone, then

```sh
git checkout -b your-new-feature
code . # make changes
git add .
git commit -m "Added thing"
git push origin your-new-feature
```

Go to the GitHub website, and create a pull request.

## Sphinx set-up

If you want to test search using Sphinx, you will need to set it up.
[Here is the sphinx.conf used in production, you probably only need to change lines 23-35](https://gist.github.com/thehowl/3dc046e2a0ab93fa1ffe5f0eca085905)

To index the data, you'd then need to run `sudo indexer --all`. If you want to run
the indexer without having to shutdown sphinx, run `sudo indexer --all --rotate`.
In production, this is run as a cronjob every 5 minutes.

(No, we're not using ElasticSearch. Search is meant to be fast and not take too
much memory. Any Java solution can thus be tossed away since it does not suit
these basic two requirements.)

## License

Seeing as this project is not meant exclusively for usage by Ripple, the license,
unlike most other Ripple projects, is [MIT](LICENSE).
