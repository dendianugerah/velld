package connection

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func decodeConnectionRequest(r *http.Request) (ConnectionConfig, error) {
	var request struct {
		ConnectionConfig
		URL *string `json:"connection_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return ConnectionConfig{}, fmt.Errorf("Invalid connection request.")
	}
	if request.URL != nil {
		if err := applyConnectionURL(&request.ConnectionConfig, *request.URL); err != nil {
			return ConnectionConfig{}, err
		}
	}
	return request.ConnectionConfig, nil
}

func applyConnectionURL(config *ConnectionConfig, raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Hostname() == "" || strings.Contains(u.Hostname(), ",") || u.Fragment != "" {
		return fmt.Errorf("Enter a valid connection URL with one host.")
	}
	var engine string
	var port int
	switch u.Scheme {
	case "postgres", "postgresql":
		engine, port = "postgresql", 5432
	case "mysql", "mariadb":
		engine, port = "mysql", 3306
	case "mongo", "mongodb":
		engine, port = "mongodb", 27017
	default:
		return fmt.Errorf("Use a PostgreSQL, MySQL/MariaDB, or MongoDB connection URL.")
	}
	if config.Type != "" && config.Type != engine {
		return fmt.Errorf("The URL does not match the selected database engine.")
	}
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("Use a port between 1 and 65535.")
		}
	}
	query, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return fmt.Errorf("Invalid connection URL options.")
	}
	config.Type, config.Host, config.Port = engine, u.Hostname(), port
	config.Username, config.Password = u.User.Username(), ""
	if password, ok := u.User.Password(); ok {
		config.Password = password
	}
	config.Database = strings.TrimPrefix(u.Path, "/")
	config.MongoAuthSource = query.Get("authSource")
	config.SSL = query.Get("sslmode") != "disable" && query.Get("ssl") != "false" && query.Get("tls") != "false"
	return nil
}
