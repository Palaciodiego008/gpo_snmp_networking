package main

import (
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
)

func main() {
	ipAddress := "192.168.1.1"
	ip := flag.String("ip", ipAddress, "IP address of the device to query")
	community := flag.String("community", "public", "SNMP community string")
	flag.Parse()

	if *ip == "" {
		log.Fatal("IP address is required")
	}

	params := &gosnmp.GoSNMP{
		Target:    *ip,
		Port:      161,
		Community: *community,
		Version:   gosnmp.Version2c,
		Timeout:   5,
	}

	ifDescrOid := ".1.3.6.1.2.1.2.2.1.2"
	ifInOctetsOid := ".1.3.6.1.2.1.2.2.1.10"
	ifOutOctetsOid := ".1.3.6.1.2.1.2.2.1.16"

	interfacesTable, err := getSNMPTable(params, []string{ifDescrOid, ifInOctetsOid, ifOutOctetsOid})
	if err != nil {
		log.Fatalf("Failed to get interfaces table: %s", err)
	}

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
	result := make(map[string]string)

	err := params.Connect()
	if err != nil {
		return nil, err
	}
	defer params.Conn.Close()

	for _, oid := range oids {
		err = params.Walk(oid, func(pdu gosnmp.SnmpPDU) error {
			result[pdu.Name] = pdu.Value.(string)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
