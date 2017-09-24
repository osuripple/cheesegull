<p align="center"><img src="https://y.zxq.co/jobeei.png"></p>

# CheeseGull (Woo branch)

This is the branch for the v2 of CheeseGull: CheeseGull v2 Woo. Seeing as this
is a major release, everything has been rewritten from scratch, there are some
breaking changes and even the project's aim has been shifted from being a
beatmap mirror to being a beatmap cache middleman and a "slave" beatmap
database from the official beatmap database through the osu! API.

The goal of the project is to work as a sort of mirror from osu!
**for osu! private servers** (like, well, Ripple). The reason why the focus
shifted from being a full beatmap mirror to only "caching" beatmaps is that
HDD/"SATA" servers are starting to become lesser and lesser, and often hell
expensive. With some luck, we got our hands on a So You Start 15€ ARM server
with a couple of terabytes. Seemed a perfect fit to our needs, until it wasn't.
Mostly because of the relatively low dl/ul speed that often made the server
inaccessible. Thus, CheeseGull Woo was written having in mind a Scaleway server
with a 150 GB cache SSD. This makes our expenses on a "mirror" server much lower
while having a server with much higher upload speeds (from our tests,
scaleway if you're lucky enough can get up to 1 Gbit). We will probably have
some slower downloads when the beatmap is not in the cache (seeing as we need to
download the beatmap and send it to the user), however the speed will be
considerably higher when the beatmap is already in the cache, which should be
most cases.
