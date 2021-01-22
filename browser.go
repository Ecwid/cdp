package cdp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Browser ...
type Browser struct {
	cmd      *exec.Cmd
	wsClient *WSClient
	deadline time.Duration
}

// GetWSClient ...
func (c Browser) GetWSClient() *WSClient {
	return c.wsClient
}

// Crash ...
func (c Browser) Crash() {
	c.wsClient.sendOverProtocol("", "Browser.crash", nil)
}

// Close close browser
func (c Browser) Close() error {
	// Close close browser and websocket connection
	exited := make(chan int)
	go func() {
		state, _ := c.cmd.Process.Wait()
		exited <- state.ExitCode()
	}()
	c.wsClient.sendOverProtocol("", "Browser.close", nil)
	select {
	case <-exited:
		return nil
	case <-time.After(c.deadline):
		if err := c.cmd.Process.Kill(); err != nil {
			return err
		}
		return errors.New("browser is not closing gracefully, process was killed")
	}
}

// Launch launch a new browser process
func Launch(ctx context.Context, userFlags ...string) (*Browser, error) {
	browser := &Browser{deadline: 10 * time.Second}
	var path string
	bin := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/usr/bin/google-chrome",
		"headless-shell",
		"chromium",
		"chromium-browser",
		"google-chrome",
		"google-chrome-stable",
		"google-chrome-beta",
		"google-chrome-unstable",
	}
	for _, c := range bin {
		if _, err := exec.LookPath(c); err == nil {
			path = c
			break
		}
	}

	userDataDir, err := ioutil.TempDir("", "tmp")
	if err != nil {
		return nil, err
	}

	flags := []string{
		"--no-first-run",
		"--no-default-browser-check",
		"--remote-debugging-port=0",
		"--hide-scrollbars",
		"--password-store=basic",
		"--use-mock-keychain",
		"--enable-automation",
		"--disable-gpu",
		"--disable-default-apps",
		"--disable-extensions",
		"--disable-browser-side-navigation",
		"--disable-features=site-per-process,TranslateUI,BlinkGenPropertyTrees",
		"--disable-background-timer-throttling",
		"--disable-backgrounding-occluded-windows",
		"--disable-renderer-backgrounding",
		"--disable-hang-monitor",
		"--enable-features=NetworkService,NetworkServiceInProcess",
		"--user-data-dir=" + userDataDir,
	}

	for _, f := range userFlags {
		flags = append(flags, f)
	}

	if os.Getuid() == 0 {
		flags = append(flags, "--no-sandbox")
	}

	browser.cmd = exec.CommandContext(ctx, path, flags...)
	stderr, err := browser.cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	defer stderr.Close()
	if err := browser.cmd.Start(); err != nil {
		return nil, err
	}
	webSocketURL, err := addrFromStderr(stderr)
	if err != nil {
		return nil, err
	}
	browser.wsClient, err = NewWebSocketClient(webSocketURL)
	return browser, err
}

func addrFromStderr(rc io.ReadCloser) (string, error) {
	defer rc.Close()
	const prefix = "DevTools listening on"
	var (
		url     = ""
		scanner = bufio.NewScanner(rc)
		lines   []string
	)
	for scanner.Scan() {
		line := scanner.Text()
		if s := strings.TrimPrefix(line, prefix); s != line {
			url = strings.TrimSpace(s)
			break
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if url == "" {
		return "", fmt.Errorf("chrome stopped too early; stderr:\n%s", strings.Join(lines, "\n"))
	}
	return url, nil
}
