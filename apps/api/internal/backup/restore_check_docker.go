package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const restoreCheckLabel = "velld.restore-check"

type dockerEngine struct {
	client *http.Client
}

func newDockerEngine() *dockerEngine {
	socket := "/var/run/docker.sock"
	if host := os.Getenv("DOCKER_HOST"); strings.HasPrefix(host, "unix://") {
		socket = strings.TrimPrefix(host, "unix://")
	}
	dialer := &net.Dialer{}
	return &dockerEngine{client: &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, "unix", socket)
		},
	}}}
}

func (d *dockerEngine) do(ctx context.Context, method, path string, body, dest any) (int, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, "http://docker"+path, reader)
	if err != nil {
		return 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		var failure struct{ Message string }
		_ = json.NewDecoder(resp.Body).Decode(&failure)
		if resp.StatusCode == http.StatusNotFound {
			return resp.StatusCode, nil
		}
		return resp.StatusCode, fmt.Errorf("Docker API %s %s returned %d: %s", method, path, resp.StatusCode, failure.Message)
	}
	if dest != nil {
		return resp.StatusCode, json.NewDecoder(resp.Body).Decode(dest)
	}
	_, err = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, err
}

func (d *dockerEngine) ping(ctx context.Context) error {
	_, err := d.do(ctx, http.MethodGet, "/_ping", nil, nil)
	return err
}

func (d *dockerEngine) ensureImage(ctx context.Context, image string) error {
	status, err := d.do(ctx, http.MethodGet, "/images/"+image+"/json", nil, nil)
	if err != nil || status != http.StatusNotFound {
		return err
	}
	name, tag, _ := strings.Cut(image, ":")
	if tag == "" {
		tag = "latest"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"http://docker/images/create?fromImage="+url.QueryEscape(name)+"&tag="+url.QueryEscape(tag), nil)
	if err != nil {
		return err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pull %s: Docker API returned %d", image, resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	for {
		var progress struct{ Error string }
		if err := decoder.Decode(&progress); err == io.EOF {
			return nil
		} else if err != nil {
			return fmt.Errorf("pull %s: %w", image, err)
		}
		if progress.Error != "" {
			return fmt.Errorf("pull %s: %s", image, progress.Error)
		}
	}
}

type throwawayContainer struct {
	id   string
	host string
	port int
}

type containerInspect struct {
	NetworkSettings struct {
		Networks map[string]struct{ IPAddress string }
		Ports    map[string][]struct{ HostIP, HostPort string }
	}
	HostConfig struct{ NetworkMode string }
}

// selfNetwork returns a Docker network shared with this process when it runs in a container.
func (d *dockerEngine) selfNetwork(ctx context.Context) string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return ""
	}
	var self containerInspect
	if status, err := d.do(ctx, http.MethodGet, "/containers/"+hostname+"/json", nil, &self); err != nil || status != http.StatusOK {
		return ""
	}
	if self.HostConfig.NetworkMode == "host" {
		return ""
	}
	for name := range self.NetworkSettings.Networks {
		return name
	}
	return ""
}

func (d *dockerEngine) start(ctx context.Context, image string, containerPort int, env []string) (*throwawayContainer, error) {
	if err := d.ensureImage(ctx, image); err != nil {
		return nil, err
	}
	network := d.selfNetwork(ctx)
	port := strconv.Itoa(containerPort) + "/tcp"
	hostConfig := map[string]any{"AutoRemove": false}
	if network != "" {
		hostConfig["NetworkMode"] = network
	} else {
		hostConfig["PortBindings"] = map[string]any{port: []map[string]string{{"HostIp": "127.0.0.1", "HostPort": ""}}}
	}
	var created struct{ Id string }
	if _, err := d.do(ctx, http.MethodPost, "/containers/create", map[string]any{
		"Image": image, "Env": env, "Labels": map[string]string{restoreCheckLabel: "true"},
		"ExposedPorts": map[string]any{port: map[string]any{}}, "HostConfig": hostConfig,
	}, &created); err != nil {
		return nil, err
	}
	container := &throwawayContainer{id: created.Id}
	if _, err := d.do(ctx, http.MethodPost, "/containers/"+created.Id+"/start", nil, nil); err != nil {
		d.remove(created.Id)
		return nil, err
	}
	var details containerInspect
	if _, err := d.do(ctx, http.MethodGet, "/containers/"+created.Id+"/json", nil, &details); err != nil {
		d.remove(created.Id)
		return nil, err
	}
	if network != "" {
		container.host, container.port = details.NetworkSettings.Networks[network].IPAddress, containerPort
	} else if bindings := details.NetworkSettings.Ports[port]; len(bindings) > 0 {
		container.host = "127.0.0.1"
		container.port, _ = strconv.Atoi(bindings[0].HostPort)
	}
	if container.host == "" || container.port == 0 {
		d.remove(created.Id)
		return nil, fmt.Errorf("could not find an address for the restore check container")
	}
	return container, nil
}

func (d *dockerEngine) remove(id string) {
	ctx, cancel := context.WithTimeout(context.Background(), restoreCheckCleanupTimeout)
	defer cancel()
	if _, err := d.do(ctx, http.MethodDelete, "/containers/"+id+"?force=true&v=true", nil, nil); err != nil {
		fmt.Printf("Warning: failed to remove restore check container %s: %v\n", id, err)
	}
}

func (d *dockerEngine) removeLeftovers(ctx context.Context) {
	filters, _ := json.Marshal(map[string][]string{"label": {restoreCheckLabel}})
	var containers []struct{ Id string }
	if _, err := d.do(ctx, http.MethodGet, "/containers/json?all=true&filters="+url.QueryEscape(string(filters)), nil, &containers); err != nil {
		return
	}
	for _, container := range containers {
		d.remove(container.Id)
	}
}
