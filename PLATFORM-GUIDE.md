# Platform-Specific Guide

Quick reference for using bandfetch on different operating systems.

## Linux

### Installation

#### From Pre-built Binary

```bash
# AMD64 (most common)
wget https://github.com/cx009/netperf/releases/download/v1.0.0/bandfetch-v1.0.0-linux-amd64.tar.gz
tar -xzf bandfetch-v1.0.0-linux-amd64.tar.gz
sudo mv bandfetch-linux-amd64 /usr/local/bin/bandfetch
sudo chmod +x /usr/local/bin/bandfetch

# ARM64 (Raspberry Pi, ARM servers)
wget https://github.com/cx009/netperf/releases/download/v1.0.0/bandfetch-v1.0.0-linux-arm64.tar.gz
tar -xzf bandfetch-v1.0.0-linux-arm64.tar.gz
sudo mv bandfetch-linux-arm64 /usr/local/bin/bandfetch
sudo chmod +x /usr/local/bin/bandfetch
```

#### Build from Source

```bash
# Install Go (if not already installed)
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone and build
git clone https://github.com/cx009/netperf.git
cd netperf
make build

# Optional: Install system-wide
sudo cp bin/bandfetch /usr/local/bin/
```

### Usage

```bash
# Basic usage
bandfetch -list urls.txt

# Save downloads
bandfetch -list urls.txt -save -out ~/downloads

# High concurrency
bandfetch -list urls.txt -workers 32
```

### Systemd Service (Optional)

Create `/etc/systemd/system/bandfetch.service`:

```ini
[Unit]
Description=Bandwidth Fetch Service
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/netperf
ExecStart=/usr/local/bin/bandfetch -list /path/to/urls.txt
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

---

## Windows

### Installation

#### From Pre-built Binary

1. Download from [Releases](https://github.com/cx009/netperf/releases)
   - `bandfetch-v1.0.0-windows-amd64.zip` (most common)
   - `bandfetch-v1.0.0-windows-arm64.zip` (Surface Pro X, etc.)

2. Extract the ZIP file

3. (Optional) Add to PATH:
   - Right-click "This PC" → Properties
   - Advanced system settings → Environment Variables
   - Edit "Path" under System variables
   - Add the folder containing `bandfetch.exe`

#### Build from Source

```powershell
# Install Go from https://golang.org/dl/

# Clone and build
git clone https://github.com/cx009/netperf.git
cd netperf
.\build-release.ps1
```

Or using Make (requires WSL or MinGW):
```powershell
make build
```

### Usage

#### Command Prompt
```cmd
bandfetch.exe -list urls.txt
bandfetch.exe -list urls.txt -save -out C:\Downloads
```

#### PowerShell
```powershell
.\bandfetch.exe -list urls.txt
.\bandfetch.exe -list urls.txt -save -out "$env:USERPROFILE\Downloads"
```

### Windows-Specific Tips

- Use double quotes for paths with spaces:
  ```cmd
  bandfetch.exe -list "C:\My Files\urls.txt"
  ```

- For scheduled tasks, use Task Scheduler:
  1. Open Task Scheduler
  2. Create Basic Task
  3. Set trigger (time/event)
  4. Action: Start a program → `bandfetch.exe`
  5. Add arguments: `-list C:\path\to\urls.txt`

---

## macOS

### Installation

#### From Pre-built Binary

```bash
# Intel Macs (AMD64)
curl -LO https://github.com/cx009/netperf/releases/download/v1.0.0/bandfetch-v1.0.0-darwin-amd64.tar.gz
tar -xzf bandfetch-v1.0.0-darwin-amd64.tar.gz
sudo mv bandfetch-darwin-amd64 /usr/local/bin/bandfetch
sudo chmod +x /usr/local/bin/bandfetch

# Apple Silicon M1/M2/M3 (ARM64)
curl -LO https://github.com/cx009/netperf/releases/download/v1.0.0/bandfetch-v1.0.0-darwin-arm64.tar.gz
tar -xzf bandfetch-v1.0.0-darwin-arm64.tar.gz
sudo mv bandfetch-darwin-arm64 /usr/local/bin/bandfetch
sudo chmod +x /usr/local/bin/bandfetch
```

First run may show security warning:
```bash
# Allow in System Preferences → Security & Privacy
# Or remove quarantine attribute:
xattr -d com.apple.quarantine /usr/local/bin/bandfetch
```

#### Using Homebrew (if available)

```bash
brew tap cx009/netperf
brew install bandfetch
```

#### Build from Source

```bash
# Install Go
brew install go

# Clone and build
git clone https://github.com/cx009/netperf.git
cd netperf
make build

# Optional: Install system-wide
sudo cp bin/bandfetch /usr/local/bin/
```

### Usage

```bash
# Basic usage
bandfetch -list urls.txt

# Save to Downloads folder
bandfetch -list urls.txt -save -out ~/Downloads

# High performance mode
bandfetch -list urls.txt -workers 24
```

### macOS-Specific Tips

- Use iTerm2 or Terminal for best experience
- For launchd automation, create `~/Library/LaunchAgents/com.yourname.bandfetch.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.yourname.bandfetch</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/bandfetch</string>
        <string>-list</string>
        <string>/path/to/urls.txt</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
```

---

## Platform Comparison

| Feature               | Linux | Windows | macOS |
|-----------------------|-------|---------|-------|
| Binary size           | ~8MB  | ~8MB    | ~8MB  |
| Install location      | `/usr/local/bin` | Any folder | `/usr/local/bin` |
| Service/daemon        | systemd | Task Scheduler | launchd |
| Terminal              | bash/zsh | cmd/PowerShell | zsh/bash |
| File paths            | `/path/to/file` | `C:\path\to\file` | `/path/to/file` |

---

## Common Issues

### Linux

**Permission denied**
```bash
chmod +x bandfetch-linux-amd64
# Or for system install:
sudo chmod +x /usr/local/bin/bandfetch
```

**Command not found**
```bash
# Add to PATH in ~/.bashrc or ~/.zshrc
export PATH=$PATH:/usr/local/bin
```

### Windows

**"Windows protected your PC"**
- Click "More info" → "Run anyway"
- Or right-click → Properties → Unblock

**PowerShell execution policy**
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### macOS

**"bandfetch cannot be opened because the developer cannot be verified"**
```bash
# Remove quarantine
xattr -d com.apple.quarantine bandfetch-darwin-arm64

# Or allow in System Preferences:
# Security & Privacy → General → Allow anyway
```

**Homebrew installation fails**
```bash
# Use manual installation method instead
```

---

## Performance Tips by Platform

### Linux
- Use `tmpfs` for maximum speed:
  ```bash
  bandfetch -list urls.txt -out /tmp/downloads
  ```
- Increase file descriptor limit:
  ```bash
  ulimit -n 4096
  ```

### Windows
- Disable Windows Defender scanning for download folder temporarily
- Use SSD for `-out` directory
- Close other network-intensive applications

### macOS
- Disable Gatekeeper scanning temporarily:
  ```bash
  sudo spctl --master-disable  # Re-enable after: sudo spctl --master-enable
  ```
- Use wired connection for better performance
- Increase worker count on Apple Silicon (M1/M2/M3)

---

## Architecture-Specific Notes

### ARM64 (Linux/Windows)
- Raspberry Pi 4 and newer supported
- AWS Graviton instances supported
- Surface Pro X supported (Windows ARM64)

### Apple Silicon (M1/M2/M3)
- Native ARM64 binary (not Rosetta)
- Best performance on macOS 12+
- Excellent power efficiency

### AMD64
- Universal compatibility
- Most tested platform
- Default choice for servers

---

## Next Steps

- Read the [main README](README.md) for usage examples
- Check [BUILD.md](BUILD.md) for building from source
- See [FEATURES.md](FEATURES.md) for feature details
