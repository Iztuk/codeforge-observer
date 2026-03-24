package storage

import (
	"database/sql"
	"time"
)

type HostInfo struct {
	Name     string `json:"name"`
	Upstream string `json:"upstream"`
	Contract string `json:"contract"`
	Resource string `json:"resource"`
}

func CreateHost(host HostInfo, db *sql.DB) error {
	now := time.Now()

	query := `
	INSERT INTO hosts (name, upstream, api_contract_file, resource_contract_file, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, host.Name, host.Upstream, host.Contract, host.Resource, now, now)
	if err != nil {
		return err
	}

	return nil
}

func ReadHosts(db *sql.DB) ([]HostInfo, error) {
	var hosts []HostInfo

	query := `SELECT name, upstream, api_contract_file, resource_contract_file FROM hosts`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var host HostInfo

		err := rows.Scan(&host.Name, &host.Upstream, &host.Contract, &host.Resource)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}

func UpdateHost(host HostInfo, db *sql.DB) error {
	now := time.Now()

	query := `
	UPDATE hosts
	WHERE name = ?
	SET name = ?, upstream = ?, api_contract_file = ?, resource_contract_file = ?, updated_at = ?
	`

	_, err := db.Exec(query, host.Name, host.Upstream, host.Contract, host.Resource, now)
	if err != nil {
		return err
	}

	return nil
}

func DeleteHost(hostName string, db *sql.DB) error {
	query := `
	DELETE FROM hosts WHERE name = ?
	`

	_, err := db.Exec(query, hostName)
	if err != nil {
		return err
	}

	return nil
}
