package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
)

const (
	ifDescrOid     = ".1.3.6.1.2.1.2.2.1.2"
	ifInOctetsOid  = ".1.3.6.1.2.1.2.2.1.10"
	ifOutOctetsOid = ".1.3.6.1.2.1.2.2.1.16"
)

func main() {
	// Set up flags
	var (
		ip        = flag.String("ip", "192.168.1.1", "IP address of the device to query")
		community = flag.String("community", "public", "SNMP community string")
	)
	flag.Parse()

	if *ip == "" {
		log.Fatal("IP address is required")
	}

	// Set up SNMP parameters
	params, err := getSNMPParams(*ip, *community)
	if err != nil {
		log.Fatalf("Failed to create SNMP params: %s", err)
	}

	// Get interfaces table
	interfacesTable, err := getSNMPTable(params, []string{ifDescrOid, ifInOctetsOid, ifOutOctetsOid})
	if err != nil {
		log.Fatalf("Failed to get interfaces table: %s", err)
	}

	// Log interface statistics
	for _, row := range interfacesTable {
		ifIndex, err := strconv.Atoi(row[0].Value.(string))
		if err != nil {
			log.Printf("Failed to parse ifIndex: %s", err)
			continue
		}

		ifName := row[1].Value.(string)

		ifInOctets, err := strconv.Atoi(row[9].Value.(string))
		if err != nil {
			log.Printf("Failed to parse ifInOctets: %s", err)
			continue
		}

		ifOutOctets, err := strconv.Atoi(row[15].Value.(string))
		if err != nil {
			log.Printf("Failed to parse ifOutOctets: %s", err)
			continue
		}

		log.Printf("Interface %d (%s) statistics: ifInOctets=%d, ifOutOctets=%d", ifIndex, ifName, ifInOctets, ifOutOctets)
	}
}

func getSNMPParams(ip, community string) (*gosnmp.GoSNMP, error) {
	params := &gosnmp.GoSNMP{
		Target:    ip,
		Port:      161,
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   5,
	}

	err := params.Connect()
	if err != nil {
		return nil, err
	}

	defer params.Conn.Close()

	return params, nil
}

func getSNMPTable(params *gosnmp.GoSNMP, baseOid []string) (map[int][]gosnmp.SnmpPDU, error) {
	result, err := walkSNMP(params, baseOid)
	if err != nil {
		return nil, err
	}

	rows := make(map[int][]gosnmp.SnmpPDU)
	for oid, value := range result {
		oidParts := strings.Split(oid, ".")
		lastIndex := oidParts[len(oidParts)-1]
		rowIndex, err := strconv.Atoi(lastIndex)
		if err != nil {
			return nil, err
		}

		row, ok := rows[rowIndex]
		if !ok {
			row = make([]gosnmp.SnmpPDU, 0)
		}

		rows[rowIndex] = append(row, gosnmp.SnmpPDU{Value: value})
	}

	return rows, nil
}

func walkSNMP(params *gosnmp.GoSNMP, oids []string) (map[string]string, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	var result map[string]string
	if len(oids) == 0 {
		return result, nil
	}

	err := params.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SNMP: %v", err)
	}
	defer params.Conn.Close()

	for _, oid := range oids {
		err = params.Walk(oid, func(pdu gosnmp.SnmpPDU) error {
			result[pdu.Name] = pdu.Value.(string)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk OID %q: %v", oid, err)
		}
	}

	return result, nil
}
