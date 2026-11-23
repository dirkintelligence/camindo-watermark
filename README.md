# camindo-watermark

[Deutsche Fassung](README.de.md)

The [camindo CMS](https://camindo.de) integration of the invisible image watermark
[image-watermark](https://github.com/andreasmayercode/image-watermark): a single, statically
linked program (Linux x86-64/ARM64, macOS, Windows, ~6 MB) that embeds and detects the
watermark with camindo's codes and, if needed, serves the checking website itself.

The method itself – frequency-domain watermark with sync pattern, robustness against
scaling, JPEG, cropping, rotation and collages, measurements and limitations – comes from
the separate, CMS-independent library image-watermark, which is developed by its own
authors in its own repository; see its README. camindo-watermark is a thin program on it
and adds:

- **Codes instead of numbers:** `-code DR-CAM` stores a code of 1–6 characters, uppercase
  letters A–Z and `-` (lowercase is converted) – in camindo CMS one code per instance. Each
  character is a digit 1–27 in base 28 (0 = no character), left-aligned: 28⁶ = 481,890,304
  codes fit exactly into the 32 bits of the ID, so a code costs neither robustness nor
  computing time. A seventh character no longer fits (27⁷ > 2³²). `-id` with a number from
  0 to 4,294,967,295 still works; `detect` shows numbers that are not a valid code only as
  `id`, otherwise also as `code`.
- **Its own namespace** `camindo-watermark`: the data format carries `camindo-watermark/1`
  and `camindo-watermark/sync1` (unchanged since 1.0 – every image marked by camindo stays
  readable). Marks of the plain `image-watermark` program are not found, and vice versa.
- The key variable `CAMINDO_WM_KEY` and the checking website with the code.

Download: [Releases](https://github.com/dirkintelligence/camindo-watermark/releases) –
one archive per platform with the program, the checking website (`web/`), license and
checksums (`SHA256SUMS`).

```
camindo-watermark embed  -in photo.jpg -out photo-wm.jpg -code DR-CAM -key-file watermark.key
camindo-watermark detect -in suspect.jpg -key-file watermark.key
camindo-watermark serve  -listen 127.0.0.1:8080 -key-file watermark.key
```

## camindo CMS

In camindo CMS the program is set up by the system administrator (setting
`paths.watermark`, key file `paths.watermark_key_file`, created with
`camindo.php watermark key`). Each instance gets its code, each upload or thumbnail field a
switch “Embed invisible watermark”. The CMS embeds the code when a file is saved; if an image
cannot carry it (too small, too little texture), the file is saved anyway, the problem is
logged and the input mask shows a warning. Images are checked in the backend module
“Watermark” or with `camindo.php watermark detect`. Details: camindo manuals for
administrators and system administrators.

## Installation (Linux)

```bash
tar -xzf camindo-watermark-1.0.0-linux-amd64.tar.gz
mkdir -p ~/watermark/bin ~/watermark/private
cp camindo-watermark-1.0.0-linux-amd64/camindo-watermark ~/watermark/bin/camindo-watermark
chmod 755 ~/watermark/bin/camindo-watermark
head -c 32 /dev/urandom | base64 > ~/watermark/private/watermark.key
chmod 600 ~/watermark/private/watermark.key
~/watermark/bin/camindo-watermark version
```

The key is passed via `-key-file` (recommended), the environment variable
`CAMINDO_WM_KEY` or `-key` (then visible in the process list).

### Integrating into image processing

```bash
camindo-watermark embed -in upload.jpg -out upload.jpg -code DR-CAM \
    -key-file ~/watermark/private/watermark.key -json
```

- Embed the watermark into the image that is **delivered**, or into the largest one.
  Smaller versions derived from it, from ~400 px on, keep it.
- `-code DR-CAM` stores a code (see above), `-id` a number.
- `-in`/`-out` may be the same file (replaced atomically) or `-` for stdin/stdout. Input
  JPEG, PNG, GIF, WebP; output according to the extension `.jpg`/`.png`.
- EXIF, ICC profile, XMP and IPTC are preserved for JPEG.
- `-strength 1.5` makes it more robust (and more likely visible), `-quality 92` is the
  JPEG quality, `-no-mask` distributes evenly (only for special cases).
- Exit code 0 = embedded. With `verified: false` (image has too little texture) a warning
  is written to stderr, likewise with a weak sync pattern (`sync_confidence` below 7:
  cropping and rotation then probably not readable).

JSON output of `embed -json`:

```json
{"ok":true,"id":80500377,"code":"DR-CAM","width":1920,"height":1280,"verified":true,"confidence":8.53,
 "texture":0.51,"delta_rms":3.18,"psnr_db":38.08,"sync_confidence":24.1}
```

### Detection

```bash
camindo-watermark detect -in image.jpg -key-file … [-json] [-v]
```

Exit code 0 = found, 1 = not found, 2 = error.

```json
{"status":"found","found":true,"id":80500377,"code":"DR-CAM","confidence":8.99,"crc_ok":true,
 "corrected_bits":0,"min_bit_z":1.35,"orientation":"none","width":960,"height":640}
```

Read via the sync pattern (here: image rotated by 12°):

```json
{"status":"found","found":true,"id":80500377,"code":"DR-CAM","confidence":8.42,"crc_ok":true,
 "corrected_bits":0,"min_bit_z":1.01,"orientation":"none","width":2146,"height":1654,
 "synced":true,"rotation_deg":12.01,"sync_confidence":24.29}
```

`code` is only present if the number is a valid code. `orientation` is always one of the
eight names (nearest multiple of 90°); `rotation_deg` is added for arbitrary angles.
`status`: `found`, `present_unreadable` (clear signal, ID not readable) or `none`.
`confidence` is the signal strength (unmarked images are mostly below 3, marked ones
between 5 and 10).

## Checking website

`web/index.html` is the upload page (German/English depending on the browser language,
light/dark). It sends the image to `detect.php` in the same directory. Images are not
stored.

**Option A: PHP web hosting.** Copy `index.html` and `detect.php` into a directory of the
website, binary and key **outside** the document root. The paths are at the top of
`detect.php` (default: `../bin/camindo-watermark`, `../private/watermark.key`) or come from
the environment variables `CAMINDO_WM_BINARY` and `CAMINDO_WM_KEY_FILE` (Apache: `SetEnv`).
PHP needs `proc_open` (disabled on some shared hosting plans, then option B).
`upload_max_filesize` and `post_max_size` must match the image size.

**Option B: built-in server** (VPS/root server). The binary serves the page and the
endpoint itself:

```bash
camindo-watermark serve -listen 127.0.0.1:8080 -key-file … -max-mb 30
```

In front of it a reverse proxy with TLS, e.g. Apache:

```apache
ProxyPass        /watermark/ http://127.0.0.1:8080/
ProxyPassReverse /watermark/ http://127.0.0.1:8080/
```

systemd unit (`/etc/systemd/system/camindo-watermark.service`):

```ini
[Unit]
Description=camindo-watermark detection website
After=network.target

[Service]
ExecStart=/opt/watermark/bin/camindo-watermark serve -listen 127.0.0.1:8080 -key-file /opt/watermark/private/watermark.key
User=www-data
Restart=on-failure
NoNewPrivileges=true
ProtectSystem=strict
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

The server processes at most two images at a time, rejects images over 120 megapixels and
sets a restrictive Content Security Policy. There is no rate limiting per IP; if needed,
that belongs in the proxy.

## Building

No local Go installation is needed. The build runs in the official Go container:

```bash
./build.sh          # dist/camindo-watermark-linux-amd64
./build.sh all      # additionally linux-arm64 and darwin-arm64
```

The program needs the library `github.com/andreasmayercode/image-watermark` in the version
of `go.mod`. During development a checkout of the library next to this repository is used
(`replace` directive in `go.mod`, `build.sh` mounts it into the container); for a release
the directive is removed and the tagged library version is used.

Releases are built by GitHub Actions: a tag `vX.Y.Z` that matches the version in `main.go`
creates the archives for Linux (amd64, arm64), macOS (arm64) and Windows (amd64) and
attaches them to a GitHub release.

The namespace `camindo-watermark` in `main.go` must never change: existing watermarks would
become unreadable.

## Disclaimer

camindo-watermark is provided without any warranty, to the extent permitted by law (see
sections 15 and 16 of the `LICENSE`). This applies to the source code as well as to the
ready-made programs (binaries):

- The binaries are built automatically by GitHub Actions from the tagged source code. They
  are not signed, so macOS and Windows may show a warning on first launch. Verify
  downloaded archives with `SHA256SUMS`, or build the program yourself from source (see
  “Building”).
- You run the program at your own risk. Test it with copies first. If `-in` and `-out` are
  the same file, the original is replaced – keep originals separately.
- Whether your hosting contract allows running your own programs, and which limits apply
  to memory and CPU time, is solely up to your provider. Check this before installing. The
  examples in this guide are not a recommendation of any particular provider or plan.
- A detected watermark is an indication of origin, not proof; a watermark that is not
  detected does not prove that an image was not marked (see “Limitations” in the README of image-watermark). The results
  are not legal advice.

## License

Copyright (C) 2026 the camindo authors. camindo-watermark is free software under the GNU
Affero General Public License, version 3 or later (`LICENSE`). Anyone who runs a modified
version as a network service (e.g. the checking website via `serve`) must offer its source
code to the users – the page links to this repository for that purpose.

The binaries contain the library image-watermark (AGPL-3.0-or-later), the Go standard
library and `golang.org/x/image` (BSD-3-Clause), see `THIRD-PARTY-NOTICES.md`.

## Third-party rights

If you believe that camindo-watermark or its use infringes your rights, please write to us
first at camindo@camindo.de. We will look into it promptly and change or remove the
affected part where necessary.

Questions and problems concerning the watermark method itself (robustness, detection,
measurements) belong in the [issues of image-watermark](https://github.com/andreasmayercode/image-watermark/issues).
