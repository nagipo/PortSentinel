# Port Sentinel

Cross-platform TCP port watcher built with Go + Fyne.

中文版文件: `README_zh-TW.md`

## Features

- Monitor TCP ports with status (FREE / IN_USE / UNKNOWN).
- Show PID, process name, command line/path, and last updated time.
- Add custom ports and toggle preset ports.
- Pin ports to keep them at the top of the list.
- One-click refresh or auto refresh.
- Terminate processes (with optional force).

## Requirements

- Go 1.22 or newer.
- Fyne v2 (via Go modules).
- CGO enabled (default for Fyne builds). If you previously disabled it, run:
  - `go env -w CGO_ENABLED=1`
- OS tools used for scanning:
  - Windows: `netstat`, `tasklist`, `wmic`, `taskkill`
  - macOS/Linux: `lsof` (preferred), `netstat` (fallback), `ps`

## Fyne Build Dependencies

Fyne uses CGO and system libraries. Make sure the native build tools are installed.

Windows (recommended)
- Install MSYS2.
- Install the MinGW-w64 toolchain (e.g. `mingw-w64-x86_64-gcc`).
- Ensure `gcc` is on PATH before running `go build`.

macOS
- Install Xcode Command Line Tools:
  - `xcode-select --install`

Linux (Debian/Ubuntu)
- Install common packages:
  - `build-essential`, `pkg-config`, `libgl1-mesa-dev`, `xorg-dev`, `libx11-dev`, `libxcursor-dev`, `libxrandr-dev`, `libxinerama-dev`, `libxi-dev`, `libxkbcommon-dev`, `libwayland-dev`, `libegl1-mesa-dev`, `libasound2-dev`

## Build / Run

```bash
go mod download
go run ./cmd/portsentinel
```

Build a binary:

```bash
go build -o portsentinel ./cmd/portsentinel
```

## Configuration

Config file location:

- `<UserConfigDir>/portsentinel/config.json`

The file includes preset ports, custom ports, pinned ports, and UI settings.

## Notes

- If a tool is missing or parsing fails, ports may show `UNKNOWN`.
- Terminate requires permissions to kill the target PID.

## License

MIT. See `LICENSE`.
