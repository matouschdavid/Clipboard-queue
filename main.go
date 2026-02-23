package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"bytes"

	"github.com/matouschdavid/Clipboard-queue/pkg/monitor"
)

var version = "v0.1.0" // overridden by -ldflags "-X main.version=..."

const plistLabel = "com.matouschdavid.cbq"

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.BinaryPath}}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogPath}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogPath}}</string>
</dict>
</plist>
`

func plistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", plistLabel+".plist"), nil
}

func installAgent() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find binary path: %w", err)
	}
	// Resolve symlinks so Homebrew-installed binaries point to the real file.
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("could not resolve binary path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	logPath := filepath.Join(home, ".cbq", "cbq.log")

	dest, err := plistPath()
	if err != nil {
		return err
	}

	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct {
		Label, BinaryPath, LogPath string
	}{plistLabel, exePath, logPath}); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(dest, buf.Bytes(), 0644); err != nil {
		return err
	}

	if out, err := exec.Command("launchctl", "load", "-w", dest).CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl load failed: %w\n%s", err, out)
	}

	fmt.Printf("CBQ installed as a login item.\n  Binary: %s\n  Log:    %s\n", exePath, logPath)
	return nil
}

func uninstallAgent() error {
	dest, err := plistPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		fmt.Println("CBQ is not installed as a login item.")
		return nil
	}

	if out, err := exec.Command("launchctl", "unload", "-w", dest).CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl unload failed: %w\n%s", err, out)
	}
	if err := os.Remove(dest); err != nil {
		return err
	}

	fmt.Println("CBQ removed from login items.")
	return nil
}

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	install     := flag.Bool("install", false, "Install CBQ as a login item (autostart on login)")
	uninstall   := flag.Bool("uninstall", false, "Remove CBQ login item")
	flag.Parse()

	switch {
	case *showVersion:
		fmt.Println(version)
	case *install:
		if err := installAgent(); err != nil {
			log.Fatalf("Install failed: %v", err)
		}
	case *uninstall:
		if err := uninstallAgent(); err != nil {
			log.Fatalf("Uninstall failed: %v", err)
		}
	default:
		log.Printf("CBQ %s", version)
		monitor.Start()
	}
}
