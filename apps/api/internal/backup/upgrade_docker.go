package backup

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/dendianugerah/velld/internal/connection"
)

const (
	upgradeVolumeLabel = "velld.upgrade-volume"
	upgradeStopTimeout = 30
)

type dockerPortBind struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type dockerMountSpec struct {
	Type     string `json:"Type"`
	Source   string `json:"Source"`
	Target   string `json:"Target"`
	ReadOnly bool   `json:"ReadOnly,omitempty"`
}

type dockerMountInfo struct {
	Type        string `json:"Type"`
	Name        string `json:"Name"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	RW          bool   `json:"RW"`
}

type dockerNetworkSettings struct {
	NetworkID string   `json:"NetworkID"`
	IPAddress string   `json:"IPAddress"`
	Aliases   []string `json:"Aliases"`
}

type upgradeContainerInspect struct {
	ID    string `json:"Id"`
	Name  string `json:"Name"`
	Image string `json:"Image"`
	State struct {
		Running   bool   `json:"Running"`
		Status    string `json:"Status"`
		ExitCode  int    `json:"ExitCode"`
		OOMKilled bool   `json:"OOMKilled"`
		Error     string `json:"Error"`
	} `json:"State"`
	Config struct {
		Image        string              `json:"Image"`
		Env          []string            `json:"Env"`
		Cmd          []string            `json:"Cmd"`
		Entrypoint   []string            `json:"Entrypoint"`
		Labels       map[string]string   `json:"Labels"`
		User         string              `json:"User"`
		WorkingDir   string              `json:"WorkingDir"`
		ExposedPorts map[string]struct{} `json:"ExposedPorts"`
	} `json:"Config"`
	HostConfig struct {
		Binds         []string                    `json:"Binds"`
		NetworkMode   string                      `json:"NetworkMode"`
		PortBindings  map[string][]dockerPortBind `json:"PortBindings"`
		RestartPolicy struct {
			Name              string `json:"Name"`
			MaximumRetryCount int    `json:"MaximumRetryCount"`
		} `json:"RestartPolicy"`
		ExtraHosts []string          `json:"ExtraHosts"`
		ShmSize    int64             `json:"ShmSize"`
		Mounts     []dockerMountSpec `json:"Mounts"`
	} `json:"HostConfig"`
	Mounts          []dockerMountInfo `json:"Mounts"`
	NetworkSettings struct {
		Networks map[string]dockerNetworkSettings `json:"Networks"`
		Ports    map[string][]dockerPortBind      `json:"Ports"`
	} `json:"NetworkSettings"`
}

type dockerImageInspect struct {
	ID     string `json:"Id"`
	Config struct {
		Env        []string `json:"Env"`
		Cmd        []string `json:"Cmd"`
		Entrypoint []string `json:"Entrypoint"`
	} `json:"Config"`
}

type dockerEndpoint struct {
	host string
	port int
}

func upgradeImageFamily(image string) (dbType string, defaultPort int) {
	base := image
	base, _, _ = strings.Cut(base, "@")
	base = base[strings.LastIndex(base, "/")+1:]
	base, _, _ = strings.Cut(base, ":")
	switch strings.ToLower(base) {
	case "postgres", "postgis", "timescaledb", "pgvector":
		return "postgresql", 5432
	case "mysql", "percona":
		return "mysql", 3306
	case "mariadb":
		return "mariadb", 3306
	case "mongo", "mongodb":
		return "mongodb", 27017
	default:
		return "", 0
	}
}

func imageMatchesConnectionType(connType, image string) (int, bool) {
	family, port := upgradeImageFamily(image)
	if family == "" {
		return 0, false
	}
	switch connType {
	case "postgresql":
		return port, family == "postgresql"
	case "mysql", "mariadb":
		return port, family == "mysql" || family == "mariadb"
	case "mongodb":
		return port, family == "mongodb"
	default:
		return 0, false
	}
}

func validateTargetUpgradeImage(connType, currentImage, targetImage string) error {
	targetImage = strings.TrimSpace(targetImage)
	if targetImage == "" {
		return fmt.Errorf("target_image is required")
	}
	if strings.ContainsAny(targetImage, " \t\r\n\x00") {
		return fmt.Errorf("target_image must not contain whitespace")
	}
	if targetImage == strings.TrimSpace(currentImage) {
		return fmt.Errorf("container is already running %s", targetImage)
	}
	if _, ok := imageMatchesConnectionType(connType, targetImage); !ok {
		return fmt.Errorf("target image %q is not a supported %s image", targetImage, connType)
	}
	return nil
}

func normalizeEndpoint(host string, port int) dockerEndpoint {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	if host == "localhost" {
		host = "127.0.0.1"
	}
	return dockerEndpoint{host: host, port: port}
}

func (d *dockerEngine) selfContainerNetworks(ctx context.Context) (map[string]dockerNetworkSettings, bool) {
	if _, err := os.Stat("/.dockerenv"); err != nil {
		return nil, false
	}
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return nil, false
	}
	var self upgradeContainerInspect
	status, err := d.do(ctx, http.MethodGet, "/containers/"+hostname+"/json", nil, &self)
	if err != nil || status != http.StatusOK {
		return nil, true
	}
	return self.NetworkSettings.Networks, true
}

func containerEndpoints(c upgradeContainerInspect, defaultPort int, ownNetworks map[string]dockerNetworkSettings, inContainer bool) []dockerEndpoint {
	var endpoints []dockerEndpoint
	name := strings.TrimPrefix(c.Name, "/")

	for netName, netCfg := range c.NetworkSettings.Networks {
		if own, shared := ownNetworks[netName]; shared && own.NetworkID == netCfg.NetworkID {
			if netName != "bridge" && name != "" {
				endpoints = append(endpoints, normalizeEndpoint(name, defaultPort))
				for _, alias := range netCfg.Aliases {
					endpoints = append(endpoints, normalizeEndpoint(alias, defaultPort))
				}
			}
			if netCfg.IPAddress != "" {
				endpoints = append(endpoints, normalizeEndpoint(netCfg.IPAddress, defaultPort))
			}
		} else if !inContainer && netCfg.IPAddress != "" {
			endpoints = append(endpoints, normalizeEndpoint(netCfg.IPAddress, defaultPort))
		}
	}

	host := "127.0.0.1"
	if inContainer {
		host = "host.docker.internal"
	}
	if c.HostConfig.NetworkMode == "host" {
		endpoints = append(endpoints, normalizeEndpoint(host, defaultPort))
	}

	portKey := strconv.Itoa(defaultPort) + "/tcp"
	for _, binding := range c.NetworkSettings.Ports[portKey] {
		published, err := strconv.Atoi(binding.HostPort)
		if err != nil || published <= 0 || published > 65535 {
			continue
		}
		ip := net.ParseIP(binding.HostIP)
		switch {
		case ip != nil && ip.IsLoopback():
			if !inContainer {
				endpoints = append(endpoints, normalizeEndpoint(binding.HostIP, published))
			}
		case binding.HostIP == "" || binding.HostIP == "0.0.0.0" || binding.HostIP == "::":
			endpoints = append(endpoints, normalizeEndpoint(host, published))
			endpoints = append(endpoints, normalizeEndpoint("127.0.0.1", published))
		case ip != nil && ip.To4() != nil:
			endpoints = append(endpoints, normalizeEndpoint(binding.HostIP, published))
		}
	}
	return endpoints
}

func (d *dockerEngine) findContainerForConnection(ctx context.Context, conn *connection.StoredConnection) (*upgradeContainerInspect, int, error) {
	var containers []struct {
		ID    string `json:"Id"`
		Image string `json:"Image"`
	}
	if _, err := d.do(ctx, http.MethodGet, "/containers/json", nil, &containers); err != nil {
		return nil, 0, fmt.Errorf("list Docker containers: %w", err)
	}

	ownNetworks, inContainer := d.selfContainerNetworks(ctx)
	target := normalizeEndpoint(conn.Host, conn.Port)

	for _, item := range containers {
		defaultPort, ok := imageMatchesConnectionType(conn.Type, item.Image)
		if !ok {
			continue
		}
		var details upgradeContainerInspect
		status, err := d.do(ctx, http.MethodGet, "/containers/"+item.ID+"/json", nil, &details)
		if err != nil || status != http.StatusOK || !details.State.Running {
			continue
		}
		for _, ep := range containerEndpoints(details, defaultPort, ownNetworks, inContainer) {
			if ep == target {
				return &details, defaultPort, nil
			}
		}
	}
	return nil, 0, errContainerNotFound
}

func (d *dockerEngine) inspectContainer(ctx context.Context, id string) (*upgradeContainerInspect, error) {
	var details upgradeContainerInspect
	status, err := d.do(ctx, http.MethodGet, "/containers/"+id+"/json", nil, &details)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("container %s not found", id)
	}
	return &details, nil
}

func (d *dockerEngine) inspectImage(ctx context.Context, ref string) (*dockerImageInspect, error) {
	var details dockerImageInspect
	status, err := d.do(ctx, http.MethodGet, "/images/"+url.PathEscape(ref)+"/json", nil, &details)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("image %s not found", ref)
	}
	return &details, nil
}

func (d *dockerEngine) stopContainer(ctx context.Context, id string, timeoutSeconds int) error {
	status, err := d.do(ctx, http.MethodPost, fmt.Sprintf("/containers/%s/stop?t=%d", id, timeoutSeconds), nil, nil)
	if status == http.StatusNotModified {
		return nil
	}
	return err
}

func (d *dockerEngine) startContainer(ctx context.Context, id string) error {
	status, err := d.do(ctx, http.MethodPost, "/containers/"+id+"/start", nil, nil)
	if status == http.StatusNotModified {
		return nil
	}
	return err
}

func (d *dockerEngine) renameContainer(ctx context.Context, id, newName string) error {
	_, err := d.do(ctx, http.MethodPost, "/containers/"+id+"/rename?name="+url.QueryEscape(newName), nil, nil)
	return err
}

func (d *dockerEngine) createVolume(ctx context.Context, name string) error {
	_, err := d.do(ctx, http.MethodPost, "/volumes/create", map[string]any{
		"Name":   name,
		"Labels": map[string]string{upgradeVolumeLabel: "true"},
	}, nil)
	return err
}

func (d *dockerEngine) removeVolume(name string) {
	ctx, cancel := context.WithTimeout(context.Background(), restoreCheckCleanupTimeout)
	defer cancel()
	_, _ = d.do(ctx, http.MethodDelete, "/volumes/"+url.PathEscape(name)+"?force=true", nil, nil)
}

func (d *dockerEngine) connectNetwork(ctx context.Context, networkName, containerID string, aliases []string) error {
	body := map[string]any{"Container": containerID}
	if len(aliases) > 0 {
		body["EndpointConfig"] = map[string]any{"Aliases": aliases}
	}
	_, err := d.do(ctx, http.MethodPost, "/networks/"+url.PathEscape(networkName)+"/connect", body, nil)
	return err
}

func (d *dockerEngine) containerLogsTail(ctx context.Context, id string, tail int) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("http://docker/containers/%s/logs?stdout=true&stderr=true&tail=%d", id, tail), nil)
	if err != nil {
		return ""
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if err != nil || len(raw) == 0 {
		return ""
	}
	return strings.TrimSpace(demuxDockerLogs(raw))
}

// demuxDockerLogs strips Docker's 8-byte multiplexed stream headers when present.
func demuxDockerLogs(raw []byte) string {
	var out bytes.Buffer
	pos := 0
	for pos+8 <= len(raw) {
		stream := raw[pos]
		if (stream != 1 && stream != 2) || raw[pos+1] != 0 || raw[pos+2] != 0 || raw[pos+3] != 0 {
			return string(raw)
		}
		size := int(binary.BigEndian.Uint32(raw[pos+4 : pos+8]))
		pos += 8
		if size < 0 || pos+size > len(raw) {
			return string(raw)
		}
		out.Write(raw[pos : pos+size])
		pos += size
	}
	if out.Len() == 0 {
		return string(raw)
	}
	return out.String()
}

func isDatabaseDataPath(dbType, target string) bool {
	cleaned := path.Clean(target)
	switch dbType {
	case "postgresql":
		return cleaned == "/var/lib/postgresql/data" || cleaned == "/var/lib/postgresql" || strings.HasPrefix(cleaned, "/var/lib/postgresql/data/")
	case "mysql", "mariadb":
		return cleaned == "/var/lib/mysql" || strings.HasPrefix(cleaned, "/var/lib/mysql/")
	case "mongodb":
		return cleaned == "/data/db" || cleaned == "/data/configdb" || strings.HasPrefix(cleaned, "/data/db/")
	default:
		return false
	}
}

func filterUserEnv(containerEnv, imageEnv []string) []string {
	imageSet := make(map[string]struct{}, len(imageEnv))
	for _, entry := range imageEnv {
		imageSet[entry] = struct{}{}
	}
	result := make([]string, 0, len(containerEnv))
	for _, entry := range containerEnv {
		if _, isDefault := imageSet[entry]; !isDefault {
			result = append(result, entry)
		}
	}
	return result
}

func userNetworkAliases(origID, origName string, aliases []string) []string {
	shortID := origID
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}
	var out []string
	for _, alias := range aliases {
		if alias == "" || alias == origID || alias == shortID || alias == origName {
			continue
		}
		out = append(out, alias)
	}
	return out
}

type upgradedContainerPlan struct {
	body            map[string]any
	extraNetworks   map[string][]string
	createdVolumes  []string
	containerName   string
}

func buildUpgradedContainerPlan(
	orig *upgradeContainerInspect,
	origImage *dockerImageInspect,
	targetImage string,
	dbType string,
	shortSuffix string,
) upgradedContainerPlan {
	origName := strings.TrimPrefix(orig.Name, "/")

	var imageEnv, imageCmd, imageEntrypoint []string
	if origImage != nil {
		imageEnv = origImage.Config.Env
		imageCmd = origImage.Config.Cmd
		imageEntrypoint = origImage.Config.Entrypoint
	}

	body := map[string]any{
		"Image": targetImage,
		"Env":   filterUserEnv(orig.Config.Env, imageEnv),
	}
	if len(orig.Config.Entrypoint) > 0 && !slices.Equal(orig.Config.Entrypoint, imageEntrypoint) {
		body["Entrypoint"] = orig.Config.Entrypoint
	}
	if len(orig.Config.Cmd) > 0 && !slices.Equal(orig.Config.Cmd, imageCmd) {
		body["Cmd"] = orig.Config.Cmd
	}
	if len(orig.Config.Labels) > 0 {
		body["Labels"] = orig.Config.Labels
	}
	if orig.Config.User != "" {
		body["User"] = orig.Config.User
	}
	if orig.Config.WorkingDir != "" {
		body["WorkingDir"] = orig.Config.WorkingDir
	}
	if len(orig.Config.ExposedPorts) > 0 {
		body["ExposedPorts"] = orig.Config.ExposedPorts
	}

	hostConfig := map[string]any{}
	if orig.HostConfig.NetworkMode != "" {
		hostConfig["NetworkMode"] = orig.HostConfig.NetworkMode
	}
	if orig.HostConfig.RestartPolicy.Name != "" {
		hostConfig["RestartPolicy"] = orig.HostConfig.RestartPolicy
	}
	if len(orig.HostConfig.ExtraHosts) > 0 {
		hostConfig["ExtraHosts"] = orig.HostConfig.ExtraHosts
	}
	if orig.HostConfig.ShmSize > 0 {
		hostConfig["ShmSize"] = orig.HostConfig.ShmSize
	}

	// Preserve exact host port bindings, filling any ephemeral HostPort with the actual bound port
	// so existing clients and Velld connections continue hitting the exact same address.
	if len(orig.HostConfig.PortBindings) > 0 {
		bindings := make(map[string][]dockerPortBind, len(orig.HostConfig.PortBindings))
		for portKey, list := range orig.HostConfig.PortBindings {
			actual := orig.NetworkSettings.Ports[portKey]
			resolved := make([]dockerPortBind, 0, len(list))
			for i, b := range list {
				if b.HostPort == "" && i < len(actual) && actual[i].HostPort != "" {
					b.HostPort = actual[i].HostPort
					if b.HostIP == "" {
						b.HostIP = actual[i].HostIP
					}
				}
				resolved = append(resolved, b)
			}
			bindings[portKey] = resolved
		}
		hostConfig["PortBindings"] = bindings
	}

	// Keep non-data mounts; isolate database data paths so the original data stays untouched for rollback.
	var createdVolumes []string
	var binds []string
	for _, bind := range orig.HostConfig.Binds {
		parts := strings.Split(bind, ":")
		if len(parts) < 2 {
			continue
		}
		dest := parts[1]
		if !isDatabaseDataPath(dbType, dest) {
			binds = append(binds, bind)
		}
	}
	var mounts []dockerMountSpec
	for _, m := range orig.Mounts {
		if isDatabaseDataPath(dbType, m.Destination) {
			// If the container used a named Docker volume for data, provision a fresh named volume
			// so the upgraded container retains named-volume persistence without mutating the old volume.
			if m.Type == "volume" && m.Name != "" && !isAnonymousVolumeName(m.Name) {
				newVol := m.Name + "-upgrade-" + shortSuffix
				createdVolumes = append(createdVolumes, newVol)
				mounts = append(mounts, dockerMountSpec{
					Type:   "volume",
					Source: newVol,
					Target: m.Destination,
				})
			}
			continue
		}
	}
	for _, m := range orig.HostConfig.Mounts {
		if !isDatabaseDataPath(dbType, m.Target) {
			mounts = append(mounts, m)
		}
	}
	if len(binds) > 0 {
		hostConfig["Binds"] = binds
	}
	if len(mounts) > 0 {
		hostConfig["Mounts"] = mounts
	}
	body["HostConfig"] = hostConfig

	// Preserve primary network aliases and record any extra networks to connect before start.
	primaryNet := orig.HostConfig.NetworkMode
	if primaryNet == "" || primaryNet == "default" {
		primaryNet = "bridge"
	}
	extraNetworks := map[string][]string{}
	if len(orig.NetworkSettings.Networks) > 0 {
		endpointsCfg := map[string]any{}
		for netName, netCfg := range orig.NetworkSettings.Networks {
			aliases := userNetworkAliases(orig.ID, origName, netCfg.Aliases)
			if netName == primaryNet {
				if len(aliases) > 0 {
					endpointsCfg[netName] = map[string]any{"Aliases": aliases}
				}
			} else if netName != "bridge" && netName != "host" && netName != "none" {
				extraNetworks[netName] = aliases
			}
		}
		if len(endpointsCfg) > 0 {
			body["NetworkingConfig"] = map[string]any{"EndpointsConfig": endpointsCfg}
		}
	}

	return upgradedContainerPlan{
		body:           body,
		extraNetworks:  extraNetworks,
		createdVolumes: createdVolumes,
		containerName:  origName,
	}
}

func isAnonymousVolumeName(name string) bool {
	if len(name) != 64 {
		return false
	}
	for _, r := range name {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

func (d *dockerEngine) createUpgradedContainer(ctx context.Context, plan upgradedContainerPlan) (string, error) {
	for _, vol := range plan.createdVolumes {
		if err := d.createVolume(ctx, vol); err != nil {
			return "", fmt.Errorf("create upgrade volume %s: %w", vol, err)
		}
	}
	var created struct{ Id string }
	path := "/containers/create"
	if plan.containerName != "" {
		path += "?name=" + url.QueryEscape(plan.containerName)
	}
	if _, err := d.do(ctx, http.MethodPost, path, plan.body, &created); err != nil {
		return "", err
	}
	for netName, aliases := range plan.extraNetworks {
		if err := d.connectNetwork(ctx, netName, created.Id, aliases); err != nil {
			d.remove(created.Id)
			return "", fmt.Errorf("connect container to network %s: %w", netName, err)
		}
	}
	return created.Id, nil
}

// resolvePublishedPort returns the currently published host port for containerPort on c, or 0 if not published.
func resolvePublishedPort(c *upgradeContainerInspect, containerPort int) int {
	bindings := c.NetworkSettings.Ports[strconv.Itoa(containerPort)+"/tcp"]
	if len(bindings) == 0 {
		return 0
	}
	port, _ := strconv.Atoi(bindings[0].HostPort)
	return port
}
