# camindo-watermark

> **Veraltet – wird nicht mehr weiterentwickelt.** Ab camindo CMS 9.2.0 verwendet camindo
> das Programm [image-watermark](https://github.com/andreasmayercode/image-watermark) (ab 1.0.0)
> direkt mit `-namespace camindo-watermark` und rechnet die Codes selbst um (jetzt auch mit den
> Ziffern 0–9). Mit camindo-watermark markierte Bilder behalten ihren Code. Bitte installieren
> Sie image-watermark statt dieses Programms.

[English version](README.md)

Die Anbindung des unsichtbaren Bild-Wasserzeichens
[image-watermark](https://github.com/andreasmayercode/image-watermark) an
[camindo CMS](https://camindo.de): ein einzelnes, statisch gelinktes Programm (Linux
x86-64/ARM64, macOS, Windows, ~6 MB), das das Wasserzeichen mit den Codes von camindo
einbettet, erkennt und bei Bedarf selbst die Prüf-Website bereitstellt.

Das Verfahren selbst – Wasserzeichen im Frequenzraum mit Synchronisationsmuster,
Robustheit gegen Skalieren, JPEG, Zuschnitt, Drehung und Collagen, Messwerte und Grenzen –
stammt aus der eigenständigen, CMS-unabhängigen Bibliothek image-watermark, die ihre
eigenen Autoren in einem eigenen Repository entwickeln; siehe deren README.
camindo-watermark ist ein schlankes Programm darauf und ergänzt:

- **Codes statt Zahlen:** `-code GS-NRD` speichert einen Code aus 1–6 Zeichen,
  Großbuchstaben A–Z und `-` (Kleinbuchstaben werden umgewandelt) – in camindo CMS ein
  Code je Instanz. Jedes Zeichen ist eine Ziffer 1–27 zur Basis 28 (0 = kein Zeichen),
  linksbündig: 28⁶ = 481 890 304 Codes passen genau in die 32 Bit der Kennung, ein Code
  kostet also weder Robustheit noch Rechenzeit. Ein siebtes Zeichen passt nicht mehr
  (27⁷ > 2³²). `-id` mit einer Zahl von 0 bis 4 294 967 295 geht weiterhin; `detect` zeigt
  Zahlen, die kein gültiger Code sind, nur als `id`, sonst zusätzlich als `code`.
- **Ein eigener Namensraum** `camindo-watermark`: Das Datenformat trägt
  `camindo-watermark/1` und `camindo-watermark/sync1` (unverändert seit 1.0 – jedes von
  camindo markierte Bild bleibt lesbar). Wasserzeichen des reinen Programms
  `image-watermark` werden nicht gefunden und umgekehrt.
- Die Schlüssel-Variable `CAMINDO_WM_KEY` und die Prüf-Website mit Code.

Download: [Releases](https://github.com/dirkintelligence/camindo-watermark/releases) –
je Plattform ein Archiv mit Programm, Prüf-Website (`web/`), Lizenz und Prüfsummen
(`SHA256SUMS`).

```
camindo-watermark embed  -in foto.jpg -out foto-wm.jpg -code GS-NRD -key-file watermark.key
camindo-watermark detect -in verdacht.jpg -key-file watermark.key
camindo-watermark serve  -listen 127.0.0.1:8080 -key-file watermark.key
```

## camindo CMS

In camindo CMS richtet der Systemadministrator das Programm ein (Einstellung
`paths.watermark`, Schlüsseldatei `paths.watermark_key_file`, angelegt mit
`camindo.php watermark key`). Jede Instanz bekommt ihren Code, jedes Upload- oder
Thumbnail-Feld einen Schalter „Unsichtbares Wasserzeichen einbetten“. Das CMS bettet den Code beim
Speichern einer Datei ein; trägt ein Bild ihn nicht (zu klein, zu strukturarm), wird die
Datei trotzdem gespeichert, das Problem protokolliert und die Eingabemaske zeigt eine
Warnung. Geprüft wird im Backend-Modul „Wasserzeichen“ oder mit
`camindo.php watermark detect`. Einzelheiten: camindo-Handbücher für Administratoren und
Systemadministratoren.

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

Der Schlüssel kommt per `-key-file` (empfohlen), über die Umgebungsvariable
`CAMINDO_WM_KEY` oder per `-key` (dann in der Prozessliste sichtbar).

### Einbinden in die Bildbearbeitung

```bash
camindo-watermark embed -in upload.jpg -out upload.jpg -code GS-NRD \
    -key-file ~/watermark/private/watermark.key -json
```

- Das Wasserzeichen in das **ausgelieferte** Bild einbetten bzw. in das größte.
  Daraus abgeleitete kleinere Fassungen ab ~400 px behalten es.
- `-code GS-NRD` speichert einen Code (siehe oben), `-id` eine Zahl.
- `-in`/`-out` dürfen dieselbe Datei sein (wird atomar ersetzt) oder `-` für
  stdin/stdout. Eingabe JPEG, PNG, GIF, WebP; Ausgabe nach Endung `.jpg`/`.png`.
- EXIF, ICC-Profil, XMP und IPTC bleiben bei JPEG erhalten.
- `-strength 1.5` macht es robuster (und eher sichtbar), `-quality 92` ist die
  JPEG-Qualität, `-no-mask` verteilt gleichmäßig (nur für Sonderfälle).
- Exit-Code 0 = eingebettet. Bei `verified: false` (Bild zu strukturarm) steht
  eine Warnung auf stderr, ebenso bei schwachem Synchronisationsmuster
  (`sync_confidence` unter 7: Zuschnitt und Drehung dann eher nicht lesbar).

JSON-Ausgabe von `embed -json`:

```json
{"ok":true,"id":132755228,"code":"GS-NRD","width":1920,"height":1280,"verified":true,"confidence":8.53,
 "texture":0.51,"delta_rms":3.18,"psnr_db":38.08,"sync_confidence":24.1}
```

### Erkennen

```bash
camindo-watermark detect -in bild.jpg -key-file … [-json] [-v]
```

Exit-Code 0 = gefunden, 1 = nicht gefunden, 2 = Fehler.

```json
{"status":"found","found":true,"id":132755228,"code":"GS-NRD","confidence":8.99,"crc_ok":true,
 "corrected_bits":0,"min_bit_z":1.35,"orientation":"none","width":960,"height":640}
```

Über das Synchronisationsmuster gelesen (hier: Bild um 12° gedreht):

```json
{"status":"found","found":true,"id":132755228,"code":"GS-NRD","confidence":8.42,"crc_ok":true,
 "corrected_bits":0,"min_bit_z":1.01,"orientation":"none","width":2146,"height":1654,
 "synced":true,"rotation_deg":12.01,"sync_confidence":24.29}
```

`code` steht nur da, wenn die Zahl ein gültiger Code ist. `orientation` ist immer einer
der acht Namen (nächstes Vielfaches von 90°), `rotation_deg` kommt bei freien Winkeln dazu. `status`: `found`, `present_unreadable` (deutliches Signal, Kennung nicht
lesbar) oder `none`. `confidence` ist die Signalstärke (unmarkierte Bilder
liegen meist unter 3, markierte zwischen 5 und 10).

## Prüf-Website

`web/index.html` ist die Upload-Seite (Deutsch/Englisch nach Browsersprache,
Hell/Dunkel). Sie schickt das Bild an `detect.php` im selben Verzeichnis.
Bilder werden nicht gespeichert.

**Variante A: PHP-Webspace.** `index.html` und `detect.php` in ein Verzeichnis
der Website kopieren, Binary und Schlüssel **außerhalb** des Document-Root. Die
Pfade stehen oben in `detect.php` (Standard: `../bin/camindo-watermark`,
`../private/watermark.key`) oder kommen aus den Umgebungsvariablen `CAMINDO_WM_BINARY` und
`CAMINDO_WM_KEY_FILE` (Apache: `SetEnv`). PHP braucht `proc_open` (bei manchen
Shared-Hosting-Tarifen gesperrt, dann Variante B). `upload_max_filesize` und
`post_max_size` müssen zur Bildgröße passen.

**Variante B: eingebauter Server** (VPS/Root-Server). Das Binary liefert Seite
und Endpunkt selbst aus:

```bash
camindo-watermark serve -listen 127.0.0.1:8080 -key-file … -max-mb 30
```

Davor ein Reverse-Proxy mit TLS, z. B. Apache:

```apache
ProxyPass        /wasserzeichen/ http://127.0.0.1:8080/
ProxyPassReverse /wasserzeichen/ http://127.0.0.1:8080/
```

systemd-Unit (`/etc/systemd/system/camindo-watermark.service`):

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

Der Server verarbeitet höchstens zwei Bilder gleichzeitig, lehnt Bilder über
120 Megapixel ab und setzt eine restriktive Content-Security-Policy.
Eine Drosselung je IP gibt es nicht; die gehört bei Bedarf in den Proxy.

## Bauen

Ein lokales Go wird nicht gebraucht. Gebaut wird im offiziellen Go-Container:

```bash
./build.sh          # dist/camindo-watermark-linux-amd64
./build.sh all      # zusätzlich linux-arm64 und darwin-arm64
```

Das Programm braucht die Bibliothek `github.com/andreasmayercode/image-watermark` in der
Version aus `go.mod`. Während der Entwicklung dient ein Checkout der Bibliothek neben
diesem Repository (`replace`-Anweisung in `go.mod`, `build.sh` bindet ihn in den Container
ein); für ein Release entfällt die Anweisung und die getaggte Bibliotheksversion gilt.

Releases baut GitHub Actions: Ein Tag `vX.Y.Z`, das zur Version in `main.go` passt,
erzeugt die Archive für Linux (amd64, arm64), macOS (arm64) und Windows (amd64) und hängt
sie an ein GitHub-Release.

Der Namensraum `camindo-watermark` in `main.go` darf sich nie ändern: Bestehende
Wasserzeichen würden unlesbar.

## Haftungsausschluss

camindo-watermark wird ohne jede Gewährleistung bereitgestellt, soweit das gesetzlich
zulässig ist (siehe Abschnitte 15 und 16 der `LICENSE`). Das gilt für den Quelltext ebenso
wie für die fertigen Programme (Binaries):

- Die Binaries werden von GitHub Actions automatisch aus dem getaggten Quelltext gebaut.
  Sie sind nicht signiert; macOS und Windows können deshalb beim ersten Start warnen.
  Prüfen Sie heruntergeladene Archive mit `SHA256SUMS`, oder bauen Sie das Programm selbst
  aus dem Quelltext (siehe „Bauen“).
- Sie führen das Programm auf eigene Verantwortung aus. Testen Sie es zuerst mit Kopien.
  Sind `-in` und `-out` dieselbe Datei, wird das Original ersetzt – bewahren Sie Originale
  gesondert auf.
- Ob Ihr Hosting-Vertrag das Ausführen eigener Programme erlaubt und welche Grenzen für
  Speicher und Rechenzeit gelten, regelt allein Ihr Anbieter. Prüfen Sie das vor der
  Installation. Die Beispiele in dieser Anleitung sind keine Empfehlung für einen
  bestimmten Anbieter oder Tarif.
- Ein gefundenes Wasserzeichen ist ein Indiz für die Herkunft, kein Beweis; ein nicht
  gefundenes beweist nicht, dass ein Bild nicht markiert war (siehe „Grenzen“ im README von image-watermark). Die
  Ergebnisse sind keine Rechtsberatung.

## Lizenz

Copyright (C) 2026 the camindo authors. camindo-watermark ist freie Software unter der
GNU Affero General Public License, Version 3 oder neuer (`LICENSE`). Wer eine geänderte
Fassung als Netzdienst betreibt (z. B. die Prüf-Website per `serve`), muss deren Quelltext
den Nutzern anbieten – die Seite verlinkt dafür auf dieses Repository.

Die Binaries enthalten die Bibliothek image-watermark (AGPL-3.0-or-later), die
Go-Standardbibliothek und `golang.org/x/image` (BSD-3-Clause), siehe
`THIRD-PARTY-NOTICES.md`.

## Rechte Dritter

Wenn Sie der Ansicht sind, dass camindo-watermark oder sein Einsatz Ihre Rechte verletzt,
schreiben Sie uns bitte zuerst an camindo@camindo.de. Wir prüfen den Hinweis umgehend und
ändern oder entfernen den betroffenen Teil, wo das nötig ist.

Fragen und Probleme zum Wasserzeichen-Verfahren selbst (Robustheit, Erkennung, Messwerte)
gehören in die [Issues von image-watermark](https://github.com/andreasmayercode/image-watermark/issues).
