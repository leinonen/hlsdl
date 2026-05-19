# hlsdl

Downloads HLS `.ts` segments from an `.m3u8` playlist, then assembles them into a single MP4.

## Usage

```sh
go run main.go -base <base-url> -input <playlist.m3u8> [-out segments] [-workers 4] [-referer <url>]
```

```
-base      base URL for resolving relative segment paths (required)
-input     path to local .m3u8 playlist file (required)
-out       output directory for segments (default: segments)
-workers   concurrent download workers (default: 4)
-referer   Referer header sent with each request
```

## Assemble

After downloading, concatenate segments into one file:

```sh
./assemble.sh [segments-dir] [output.mp4]
```

Requires `ffmpeg`. Sorts segments numerically by the `-N` part of the filename.

## Example

```sh
go run main.go -base https://example.com/hls/ -input playlist.m3u8 -referer https://example.com
./assemble.sh segments video.mp4
```
