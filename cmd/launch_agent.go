package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const launchAgentLabel = "com.automationhub.master"

type launchAgent struct {
	Label            string
	Program          string
	ProgramArguments []string
	KeepAlive        bool
	RunAtLoad        bool
}

// newLaunchAgent describes a LaunchAgent that runs the hub binary in
// executableDir as master with the given config.
func newLaunchAgent(executableDir, configPath string) launchAgent {
	return launchAgent{
		Label:            launchAgentLabel,
		Program:          executableDir + "/hub",
		ProgramArguments: []string{"master", "--config", configPath},
		KeepAlive:        true,
		RunAtLoad:        true,
	}
}

func launchAgentPath(home string) string {
	return filepath.Join(home, "Library", "LaunchAgents", launchAgentLabel+".plist")
}

// writeLaunchAgent renders the plist of a into path.
func writeLaunchAgent(path string, a launchAgent) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create LaunchAgents directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create plist: %w", err)
	}
	t := template.Must(template.New("launchdConfig").Parse(macLaunchAgentTemplate()))
	renderErr := t.Execute(f, a)
	if renderErr != nil {
		renderErr = fmt.Errorf("render plist: %w", renderErr)
	}
	return errors.Join(renderErr, f.Close())
}

func macLaunchAgentTemplate() string {
	return `<?xml version='1.0' encoding='UTF-8'?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd" >
<plist version='1.0'>
  <dict>
    <key>Label</key><string>{{.Label}}</string>
    <key>EnvironmentVariables</key>
    <dict>
      <key>PATH</key>
      <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:</string>
    </dict>
    <key>ProgramArguments</key>
    <array>
      <string>{{.Program}}</string>
      {{range $val := .ProgramArguments}}<string>{{$val}}</string>
      {{end}}
    </array>
    <key>StandardOutPath</key><string>/tmp/{{.Label}}.out.log</string>
    <key>StandardErrorPath</key><string>/tmp/{{.Label}}.err.log</string>
    <key>KeepAlive</key><{{.KeepAlive}}/>
    <key>RunAtLoad</key><{{.RunAtLoad}}/>
  </dict>
</plist>
`
}
