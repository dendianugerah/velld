package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type dockerDiscovery struct {
	client       *http.Client
	selfHostname string
}

func newDockerDiscovery(socket, selfHostname string) *dockerDiscovery {
	dialer := &net.Dialer{}
	return &dockerDiscovery{
		selfHostname: selfHostname,
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return dialer.DialContext(ctx, "unix", socket)
			}},
		},
	}
}

type dockerContainer struct {
	ID    string
	Image string
}

type dockerNetwork struct {
	NetworkID string
	IPAddress string
	Aliases   []string
}

type dockerBinding struct {
	HostIP   string
	HostPort string
}

type dockerInspect struct {
	Name   string
	Config struct {
		Image string
		Env   []string
	}
	State           struct{ Running bool }
	NetworkSettings struct {
		Networks map[string]dockerNetwork
		Ports    map[string][]dockerBinding
	}
	HostConfig struct{ NetworkMode string }
}

type dockerEndpoint struct {
	host string
	port int
}

type dockerCandidate struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	DBName string `json:"database"`

	config    ConnectionConfig
	endpoints []dockerEndpoint
}

func dockerImageName(image string) string {
	image, _, _ = strings.Cut(image, "@")
	image = image[strings.LastIndex(image, "/")+1:]
	image, _, _ = strings.Cut(image, ":")
	return image
}

func dockerImage(image string) (dbType string, port int) {
	switch dockerImageName(image) {
	case "postgres", "postgis", "timescaledb", "pgvector":
		return "postgresql", 5432
	case "mysql", "mariadb", "percona":
		return "mysql", 3306
	case "mongo", "mongodb":
		return "mongodb", 27017
	default:
		return "", 0
	}
}

func dockerConnection(id string, container dockerInspect, ownNetworks map[string]dockerNetwork, inContainer bool) (dockerCandidate, bool) {
	dbType, port := dockerImage(container.Config.Image)
	if dbType == "" || !container.State.Running {
		return dockerCandidate{}, false
	}

	name := strings.TrimPrefix(container.Name, "/")
	endpoints := dockerEndpoints(container, name, port, ownNetworks, inContainer)
	if len(endpoints) == 0 {
		return dockerCandidate{}, false
	}

	env := make(map[string]string, len(container.Config.Env))
	for _, entry := range container.Config.Env {
		if key, value, ok := strings.Cut(entry, "="); ok {
			env[key] = value
		}
	}
	config := ConnectionConfig{Name: name, Type: dbType, Host: endpoints[0].host, Port: endpoints[0].port}
	switch dbType {
	case "postgresql":
		config.Username, config.Password = env["POSTGRES_USER"], env["POSTGRES_PASSWORD"]
		if config.Username == "" {
			config.Username = "postgres"
		}
		config.Database = env["POSTGRES_DB"]
		if config.Database == "" {
			config.Database = config.Username
		}
	case "mysql":
		if dockerImageName(container.Config.Image) == "mariadb" {
			for _, key := range []string{"USER", "PASSWORD", "DATABASE", "ROOT_PASSWORD"} {
				if value := env["MARIADB_"+key]; value != "" {
					env["MYSQL_"+key] = value
				}
			}
		}
		config.Username, config.Password = env["MYSQL_USER"], env["MYSQL_PASSWORD"]
		config.Database = env["MYSQL_DATABASE"]
		if config.Username == "" || config.Password == "" {
			config.Username, config.Password = "root", env["MYSQL_ROOT_PASSWORD"]
		}
	case "mongodb":
		config.Username, config.Password = env["MONGO_INITDB_ROOT_USERNAME"], env["MONGO_INITDB_ROOT_PASSWORD"]
		config.Database = env["MONGO_INITDB_DATABASE"]
		if config.Username != "" {
			config.MongoAuthSource = "admin"
		}
	}

	return dockerCandidate{
		ID: id, Name: name, Type: dbType, Host: config.Host, Port: config.Port, DBName: config.Database,
		config: config, endpoints: endpoints,
	}, true
}

func dockerEndpoints(container dockerInspect, name string, port int, ownNetworks map[string]dockerNetwork, inContainer bool) []dockerEndpoint {
	var endpoints []dockerEndpoint
	var networkNames []string
	for networkName := range container.NetworkSettings.Networks {
		networkNames = append(networkNames, networkName)
	}
	sort.Strings(networkNames)
	for _, networkName := range networkNames {
		network := container.NetworkSettings.Networks[networkName]
		if own, shared := ownNetworks[networkName]; shared && own.NetworkID == network.NetworkID {
			if networkName != "bridge" && name != "" {
				endpoints = append(endpoints, dockerEndpoint{name, port})
				for _, alias := range network.Aliases {
					endpoints = append(endpoints, dockerEndpoint{alias, port})
				}
			}
			if network.IPAddress != "" {
				endpoints = append(endpoints, dockerEndpoint{network.IPAddress, port})
			}
		}
	}
	host := "127.0.0.1"
	if inContainer {
		host = "host.docker.internal"
	}
	if container.HostConfig.NetworkMode == "host" {
		endpoints = append(endpoints, dockerEndpoint{host, port})
	}
	for _, binding := range container.NetworkSettings.Ports[strconv.Itoa(port)+"/tcp"] {
		published, err := strconv.Atoi(binding.HostPort)
		if err != nil || published <= 0 || published > 65535 {
			continue
		}
		ip := net.ParseIP(binding.HostIP)
		switch {
		case ip != nil && ip.IsLoopback():
			if !inContainer {
				endpoints = append(endpoints, dockerEndpoint{binding.HostIP, published})
			}
		case binding.HostIP == "" || binding.HostIP == "0.0.0.0":
			endpoints = append(endpoints, dockerEndpoint{host, published})
		case ip != nil && ip.To4() != nil:
			endpoints = append(endpoints, dockerEndpoint{binding.HostIP, published})
		}
	}
	return endpoints
}

func (d *dockerDiscovery) get(ctx context.Context, path string, dest any) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker"+path, nil)
	if err != nil {
		return false, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Docker API returned %d", resp.StatusCode)
	}
	return true, json.NewDecoder(resp.Body).Decode(dest)
}

func (d *dockerDiscovery) scan(ctx context.Context) ([]dockerCandidate, string) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	candidates := []dockerCandidate{}
	var containers []dockerContainer
	if found, err := d.get(ctx, "/containers/json", &containers); err != nil || !found {
		return candidates, "Docker is unavailable. Check that the API can access /var/run/docker.sock."
	}

	var self dockerInspect
	if d.selfHostname != "" {
		if _, err := d.get(ctx, "/containers/"+d.selfHostname+"/json", &self); err != nil {
			return candidates, "Could not read the API container's Docker networks."
		}
	}
	var hint string
	for _, container := range containers {
		if dbType, _ := dockerImage(container.Image); dbType == "" {
			continue
		}
		var details dockerInspect
		found, err := d.get(ctx, "/containers/"+container.ID+"/json", &details)
		if err != nil {
			hint = "Some Docker containers could not be inspected."
			continue
		}
		if found {
			if candidate, ok := dockerConnection(container.ID, details, self.NetworkSettings.Networks, d.selfHostname != ""); ok {
				candidates = append(candidates, candidate)
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Name < candidates[j].Name })
	return candidates, hint
}

func (c dockerCandidate) alreadyAdded(connections map[string]map[dockerEndpoint]bool) bool {
	for _, endpoint := range c.endpoints {
		if connections[c.Type][normalizeDockerEndpoint(endpoint.host, endpoint.port)] {
			return true
		}
	}
	return false
}

func normalizeDockerEndpoint(host string, port int) dockerEndpoint {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if host == "localhost" {
		host = "127.0.0.1"
	}
	return dockerEndpoint{host, port}
}
