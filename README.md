# gpo_snmp_networking

This program is an example of how to use the GoSNMP package to retrieve statistics from a network device via SNMP (Simple Network Management Protocol). It starts by parsing the IP address and SNMP community string from command line arguments, or using default values. It then creates a GoSNMP object with these parameters, specifying the SNMP version and timeout value. 

The program then defines the OIDs (Object Identifiers) for the interface description, incoming octets, and outgoing octets. It uses the `getSNMPTable` function to retrieve these values from the device and store them in a slice of slices of `gosnmp.SnmpPDU` objects. 

Next, the program loops through each row in the table and extracts the `ifIndex`, `ifName`, `ifInOctets`, and `ifOutOctets` values. It converts the latter three values from string to integer format, and logs the statistics for each interface to the console.

Finally, the program defines two helper functions. The `getSNMPTable` function uses the `walkSNMP` function to retrieve SNMP data from the device and parse it into a table format. The `walkSNMP` function performs a SNMP walk operation on the specified OIDs, retrieving their values and storing them in a map with the OID as the key and the value as the value.
