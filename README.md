# Plex
> plexd

User-space UDP broker to replace IP multicasts where they are not supported by forwarding packets to subscribers manually.
For example, let's say you have a video stream and you want to deliver it in real-time via RTP to your friends - just use Plex!

## Building

Go to [`cmd/plexd`](cmd/plexd) directory and build with `go build`.
Take a look at example config at [`examples/config.toml`](examples/config.toml).

## License

This code is under MIT license. See [LICENSE](LICENSE) for more information

Copyright (C) 2020 Hexawolf (hexawolf at hexanet dot dev) 
