// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"database/sql"
	"fmt"
	"net"
	"strings"
)

type IPManager struct {
	Uuid UUID
	db   *sql.DB
}

func NewIPManager(db *sql.DB, uuid UUID, subnet string, offset int, numAddresses int, except string) (*IPManager, error) {
	ipManager := IPManager{
		Uuid: uuid,
		db:   db,
	}
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		errMsg := "Invalid subnet format"
		Logger.Error(err, errMsg)
		return nil, fmt.Errorf("%s: %v", errMsg, err)
	}

	exceptionIP := net.ParseIP(except).To4()
	if exceptionIP == nil {
		errMsg := "Invalid exception IP address format"
		Logger.Error(err, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	start := ipNet.IP.To4()
	if start == nil {
		errMsg := "Invalid start IP address format"
		Logger.Error(err, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	tx, err := db.Begin()
	if err != nil {
		errMsg := "Cannot begin SQL transaction"
		Logger.Error(err, errMsg)
		return nil, GrpcErrorFromSql(err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				Logger.Error(err, "failed to rollback transaction")
				return
			}
		}
	}()

	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM gateway WHERE gateway_id = $1)", ipManager.Uuid).Scan(&exists)
	if err != nil {
		errMsg := "Failed to check gateway existence"
		Logger.Error(err, errMsg)
		return nil, fmt.Errorf("%s: %v", errMsg, err)
	}

	if exists {
		// Gateway already exists, return the existing IP manager
		Logger.Info(fmt.Sprintf("Gateway with ID %s already exists", ipManager.Uuid))
		return &ipManager, nil
	}
	var ipList []string
	ip := nextIP(start)
	for i := 0; i < offset+numAddresses; i++ {
		if i >= offset {
			if !ip.Equal(exceptionIP) {
				ipList = append(ipList, fmt.Sprintf("('%s', '%s')", ipManager.Uuid, ip.String()))
			}
		}
		ip = nextIP(ip)
	}

	if len(ipList) > 0 {
		query := fmt.Sprintf("INSERT INTO nat_available_ips (ip_manager_uuid, ip) VALUES %s", strings.Join(ipList, ", "))
		_, err = tx.Exec(query)
		if err != nil {
			errMsg := "Failed to insert available IPs"
			Logger.Error(err, errMsg)
			if err := tx.Rollback(); err != nil {
				errMsg = "failed to rollback transaction"
				return nil, fmt.Errorf("%s: %v", errMsg, err)
			}
			return nil, fmt.Errorf("%s: %v", errMsg, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		errMsg := "Cannot commit SQL transaction"
		Logger.Error(err, errMsg)
		return nil, GrpcErrorFromSql(err)
	}

	Logger.Info(fmt.Sprintf("Inserted %d available IPs into nat_available_ips table for IP Manager %s",
		len(ipList), ipManager.Uuid))

	return &ipManager, nil
}

func (ipManager *IPManager) AllocateIP(tx *sql.Tx) (string, error) {
	var ip string
	err := tx.QueryRow("SELECT ip FROM nat_available_ips WHERE ip_manager_uuid = $1 LIMIT 1", ipManager.Uuid).Scan(&ip)
	if err != nil {
		errMsg := "Failed to query available IPs"
		if err == sql.ErrNoRows {
			errMsg = "No available IP addresses"
		}
		Logger.Error(err, errMsg)
		return "", fmt.Errorf("%s: %v", errMsg, err)
	}

	_, err = tx.Exec("DELETE FROM nat_available_ips WHERE ip_manager_uuid = $1 AND ip = $2", ipManager.Uuid, ip)
	if err != nil {
		errMsg := "Failed to remove allocated IP from available IPs"
		Logger.Error(err, errMsg)
		return "", fmt.Errorf("%s: %v", errMsg, err)
	}

	_, err = tx.Exec("INSERT INTO nat_allocated_ips (ip_manager_uuid, ip) VALUES ($1, $2)", ipManager.Uuid, ip)
	if err != nil {
		errMsg := "Failed to insert allocated IP"
		Logger.Error(err, errMsg)
		return "", fmt.Errorf("%s: %v", errMsg, err)
	}

	Logger.Info(fmt.Sprintf("Allocated IP: %s for IP Manager %s", ip, ipManager.Uuid))

	return ip, nil
}

func (ipManager *IPManager) FreeIP(tx *sql.Tx, ip string) error {
	var allocatedIP string
	err := tx.QueryRow("SELECT ip FROM nat_allocated_ips WHERE ip_manager_uuid = $1 AND ip = $2", ipManager.Uuid, ip).Scan(&allocatedIP)
	if err != nil {
		errMsg := fmt.Sprintf("IP address %s is not allocated", ip)
		if err != sql.ErrNoRows {
			errMsg = "Failed to query allocated IPs"
		}
		Logger.Error(err, errMsg)
		return fmt.Errorf("%s: %v", errMsg, err)
	}

	_, err = tx.Exec("DELETE FROM nat_allocated_ips WHERE ip_manager_uuid = $1 AND ip = $2", ipManager.Uuid, ip)
	if err != nil {
		errMsg := "Failed to remove IP from allocated IPs"
		Logger.Error(err, errMsg)
		return fmt.Errorf("%s: %v", errMsg, err)
	}

	_, err = tx.Exec("INSERT INTO nat_available_ips (ip_manager_uuid, ip) VALUES ($1, $2)", ipManager.Uuid, ip)
	if err != nil {
		errMsg := "Failed to insert freed IP into available IPs"
		Logger.Error(err, errMsg)
		return fmt.Errorf("%s: %v", errMsg, err)
	}

	Logger.Info(fmt.Sprintf("Freed IP: %s for IP Manager %s", ip, ipManager.Uuid))

	return nil
}

func nextIP(ip net.IP) net.IP {
	ip = ip.To4()
	next := make(net.IP, len(ip))
	copy(next, ip)

	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] > 0 {
			break
		}
	}
	return next
}
