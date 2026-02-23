# Port Sentinel

使用 Go + Fyne 製作的跨平台 TCP Port 監看工具。

English README: `README.md`

## 專案背景與技術選擇

- 本專案為使用 Codex CLI 進行實驗的產出。
- 本人並非 Go 語言工程師，僅具備基礎語法認識，因此程式尚未經過人工審核。
- 選題目的為解決工作上的實務困擾。
- 採用 Go 的原因在於可避免工具執行時依賴 Node.js、JVM 等執行環境，並可編譯為單一執行檔；此外，就本專案範圍而言，Go 的環境建置與程式複雜度相對易於掌握。
- 選擇 Fyne 作為 UI，主要考量其依賴較輕量，並避免形成 Web UI + Go 後端的較複雜架構。

## 功能

- 監看 TCP port 狀態（FREE / IN_USE / UNKNOWN）。
- 顯示 PID、程序名稱、命令列/路徑、最後更新時間。
- 可新增自訂 port，並切換預設 port。
- 支援釘選（可複數），釘選項目會固定在列表上方。
- 一鍵刷新或自動刷新。
- 可終止程序（支援強制終止）。

## 編譯環境需求

- Go 1.22 或以上。
- Fyne v2（透過 Go modules 取得）。
- 必須啟用 CGO（Fyne 需要）。如果你曾關閉，請執行：
  - `go env -w CGO_ENABLED=1`
- 依賴作業系統工具進行掃描：
  - Windows: `netstat`, `tasklist`, `wmic`, `taskkill`
  - macOS/Linux: `lsof`（優先）、`netstat`（備援）、`ps`

## Fyne 編譯依賴

Fyne 會使用 CGO 與系統原生函式庫，請先安裝建置工具。

Windows（建議）
- 安裝 MSYS2。
- 安裝 MinGW-w64 工具鏈（例如 `mingw-w64-x86_64-gcc`）。
- 確保 `gcc` 在 PATH 中，才能執行 `go build`。

macOS
- 安裝 Xcode Command Line Tools：
  - `xcode-select --install`

Linux（Debian/Ubuntu）
- 安裝常見套件：
  - `build-essential`, `pkg-config`, `libgl1-mesa-dev`, `xorg-dev`, `libx11-dev`, `libxcursor-dev`, `libxrandr-dev`, `libxinerama-dev`, `libxi-dev`, `libxkbcommon-dev`, `libwayland-dev`, `libegl1-mesa-dev`, `libasound2-dev`

## 建立與執行

```bash
go mod download
go run ./cmd/portsentinel
```

建立可執行檔：

```bash
go build -o portsentinel ./cmd/portsentinel
```

## 設定檔

設定檔位置：

- `<UserConfigDir>/portsentinel/config.json`

內容包含 preset ports、custom ports、pinned ports 與 UI 設定。

## 備註

- 若缺少工具或解析失敗，port 可能顯示 `UNKNOWN`。
- 終止程序需要足夠的權限。

## 授權

MIT，詳見 `LICENSE`。
