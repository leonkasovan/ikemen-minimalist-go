# ikemen-minimalist

A minimalist [Ikemen Go](https://github.com/ikemen-engine/Ikemen_GO)-style sprite viewer with palette remapping, written in Go using SDL2 and OpenGL 3.3 core. Loads a character's SFF (sprite) and AIR (animation) files and displays them with real-time keyboard controls.

## Features

- **SFF v1 & v2 support** — Loads indexed (8-bit paletted) and RGBA sprites from both SFF versions.
- **AIR animation playback** — Parses M.U.G.E.N-style `.air` animation definitions with frame timing, offsets, flips, scaling, and rotation.
- **Palette remapping** — Cycle through palettes, remap individual sprite palettes, clear overrides back to defaults.
- **Blend modes** — Alpha, additive (`TransAdd`), and subtractive (`TransSub`) rendering.
- **OpenGL 3.3 core** — Indexed sprites rendered as `GL_R8` textures with palette lookups in the fragment shader; RGBA sprites rendered directly.
- **Draw list sorting** — Sprites sorted by layer, priority, and insertion order (M.U.G.E.N-compatible model).
- **Render parity diagnostics** — Logs SFF/AIR counts, missing AIR sprite references, and KFM regression checks before rendering.
- **Log capture** — Optional `-log <file>` flag to redirect stdout to a file (useful with the GUI-subsystem build).

## Screenshot

![screenshot](https://i.imgur.com/n3cP3t6.png)

## Controls

| Key | Action |
|-----|--------|
| `Esc` | Quit |
| `→`, `Space`, `N` | Next animation |
| `←`, `P` | Previous animation |
| `R` | Random animation |
| `1` | Alpha blend mode |
| `2` | Additive blend mode |
| `3` | Subtractive blend mode |
| `4` | Clear palette override / remap, use default palette |
| `[` | Previous palette override |
| `]` | Next palette override |
| `5` | Remap current sprite's default palette to next SFF palette |
| `6` | Remap current sprite's default palette to previous SFF palette |

The window title updates with the current action name, blend mode, and palette — no console needed for feedback.


## Render Parity Diagnostics

At startup the viewer prints a small Ikemen render parity report before creating the SDL window:

```text
Ikemen render parity diagnostics:
  SFF version: v2
  sprite header count: 281
  decoded sprite count: 281
  linked sprite count: 41
  palette count: ...
  AIR action count: 117
  drawable action count: 117
  action 5100 drawable: true
  action 5170 drawable: true
  missing AIR refs: none
```

For the bundled `bin/kfm.sff` and `bin/kfm.air`, these diagnostics are also treated as a regression check. The app fails early if the drawable action count is not `117`, if action `5100` or `5170` is not drawable, or if any AIR frame references a missing sprite.

This is the first render-parity step toward Ikemen original rendering: make the asset and animation layer trustworthy before adding PalFX, camera/localcoord, shadows, reflections, and more advanced render passes.

## Build

### Requirements

- [w64devkit](https://github.com/skeeto/w64devkit) (Windows) or a system with Go, GCC/Mingw, and SDL2 development headers (Linux/macOS).
- Go 1.22+
- SDL2, SDL2_image, SDL2_mixer, SDL2_ttf development libraries (for Linux builds).

### Make targets

```bash
make deps            # go mod tidy
make build           # default GUI-subsystem build (Windows: -H=windowsgui)
make build-gui       # explicit GUI build with -trimpath -tags static
make build-console   # console-subsystem build (visible terminal output)
make run SFF=char.sff AIR=char.air
make clean           # remove bin/
```

Output: `bin/ikemen-minimalist(.exe)`

### Windows / w64devkit note

The default build uses the **GUI subsystem** (`-H=windowsgui`), which avoids focus/input weirdness where the SDL window appears but keyboard events don't reach it reliably. To capture stdout without a console window, use the `-log` flag (see below). If you need visible terminal output, build the console variant:

```bash
make build-console
```

### Manual build

```bash
cd packages/gl && go generate   # if needed
CGO_ENABLED=1 go build -ldflags "-H=windowsgui -s -w" -o bin/ikemen-minimalist.exe ./src/
```

## Usage

```bash
bin/ikemen-minimalist [-log <file>] <character.sff> <character.air>
```

### Examples

```bash
# Basic usage
bin/ikemen-minimalist kfm.sff kfm.air

# With log capture (GUI build)
bin/ikemen-minimalist -log debug.log kfm.sff kfm.air

# Via Makefile
make run SFF=kfm.sff AIR=kfm.air
```

The `-log` flag is scanned from `os.Args` before positional argument parsing. When specified, `os.Stdout` is redirected to the given file so all `fmt.Println` output is captured. This is especially useful with the default GUI-subsystem build, which has no console window.

Test assets are included in `bin/`:
- `bin/kfm.sff` — Kung Fu Man sprite data
- `bin/kfm.air` — Kung Fu Man animation definitions

## File Format Support

### SFF v1 (ElecbyteSpr)

- PCX-encoded indexed sprites with embedded 256-color palette (VGA palette footer).
- RLE decompression (PCX-style run-length encoding).
- Palette shared across sprites or per-sprite.

### SFF v2 (ElecbyteSpr)

- Supports indexed and RGBA sprites.
- Compression formats: `RLE8` (format 2), `RLE5` (format 3), `LZ5` (format 4).
- PNG-encoded sprites (formats 10/11/12): paletted PNG (format 10) or full RGB/RGBA PNG.
- Uncompressed raw sprites (format 0, depth 8/24/32).
- Linked palettes — palettes can reference other palette entries.

### AIR (M.U.G.E.N animation files)

- `[Begin Action N]` section headers with frame definition lines.
- Frame format: `group,number,xoff,yoff,time[,flags][,?,xscale,yscale,angle]`
- Flags: `H` = horizontal flip, `V` = vertical flip.
- Semicolons denote comments.

## Rendering Pipeline

```
AnimationInstance.Draw()
    └─ resolves current AIR frame
    └─ resolves Sprite from SFF
    └─ resolves PaletteEntry (override → default → remap)
    └─ emits RenderParams to DrawList

DrawList
    └─ collects RenderParams
    └─ sorts by Layer → Priority → Insertion order
    └─ submits each to GLRenderer.RenderSprite()

GLRenderer
    └─ caches indexed sprites as GL_R8 textures
    └─ caches palettes as GL_RGBA8 256×1 textures
    └─ caches RGBA sprites as GL_RGBA8 textures
    └─ fragment shader: for indexed sprites, samples index texture and
       fetches palette color via texelFetch; for RGBA sprites, samples directly
    └─ renders with alpha/additive/subtractive blending
```

This is close to Ikemen's model: sprite pixel data is stable, palettes are independent, and each draw command decides which palette to bind.

## Project Structure

```
ikemen-minimalist-go/
├── src/
│   ├── main.go               # Entry point, SDL/GL setup, event loop, -log flag
│   ├── sff.go                # SFF v1/v2 loader: PCX, RLE8, RLE5, LZ5, PNG decoders
│   ├── air.go                # AIR animation parser & playback
│   ├── animation_instance.go # AnimationInstance: stateful sprite/animation runner
│   ├── drawlist.go           # DrawList: collects, sorts, and issues draw commands
│   ├── gl_renderer.go        # OpenGL 3.3 renderer with sprite/palette caching
│   └── render_params.go      # Data types: RenderParams, PaletteEntry, PalRemap, TransType
├── bin/                      # Built executable and test assets (kfm.sff, kfm.air)
├── packages/
│   ├── gl/                   # Local Go OpenGL bindings (multiple GL versions)
│   └── go-sdl2/              # Local Go SDL2 bindings (SDL, image, mixer, ttf, gfx)
├── Makefile                  # Build system with GUI/console targets
├── go.mod / go.sum           # Go module definition
├── README.md                 # This file
├── AGENTS.md                 # Project guidance for AI coding assistants
├── fixes.md                  # Record of applied bug fixes
└── review.md                 # Architecture review notes
```

## Dependencies

All dependencies are vendored as local packages under `packages/`:

- **[go-sdl2](https://github.com/veandco/go-sdl2)** — Go bindings for SDL2 (SDL, image, mixer, ttf, gfx). Used for window creation, OpenGL context, and event handling.
- **[gl](https://github.com/go-gl/gl)** — Go OpenGL bindings (Go-generated from Khronos specs). Used for OpenGL 3.3 core profile rendering.

Module replacements in `go.mod`:

```go
replace github.com/ikemen-minimalist/gl => ./packages/gl
replace github.com/ikemen-minimalist/go-sdl2 => ./packages/go-sdl2
```

## Known Issues & Tips

- **Keyboard not working?** Use `make build-console` or the `-log <file>` flag with the default GUI build. The GUI-subsystem avoids focus-fighting between the console and SDL window.
- **OpenGL errors?** Ensure your graphics drivers support OpenGL 3.3 core. The error message in `gl_renderer.go` suggests "update graphics drivers."
- **No animations displayed?** The program filters to only animations that reference sprites present in the SFF. If an AIR references sprites not in the SFF, those animations are skipped.
- **go.sum is empty** — This is normal for local-only module replacements. Run `make deps` (or `go mod tidy`) if you add external dependencies.

## License

This project is provided as-is for educational and reference purposes. The vendored packages (go-sdl2, gl) retain their original licenses.
