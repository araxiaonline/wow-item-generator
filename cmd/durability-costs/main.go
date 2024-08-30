package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

type DurabilityCost struct {
	ID                 uint32
	WeaponSubClassCost [21]uint32
	ArmorSubClassCost  [8]uint32
}

type DBCHeader struct {
	Magic           [4]byte
	RecordCount     uint32
	FieldCount      uint32
	RecordSize      uint32
	StringBlockSize uint32
}

// read a dbc file and return the header information
func ReadDBCHeader(filepath string) (DBCHeader, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return DBCHeader{}, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var header DBCHeader
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return DBCHeader{}, fmt.Errorf("failed to read DBC header: %v", err)
	}

	if string(header.Magic[:]) != "WDBC" {
		return DBCHeader{}, fmt.Errorf("invalid DBC file: wrong magic identifier")
	}

	return header, nil
}

func GetStringOffset(filepath string, header DBCHeader) (int64, error) {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to open file for header update: %v", err)
	}
	defer file.Close()

	stringBlocks, err := file.Seek(-int64(header.StringBlockSize), io.SeekEnd)
	if err != nil {
		return 0, fmt.Errorf("failed to seek to string block: %v", err)
	}

	return stringBlocks, nil
}

func GetStringBlock(filepath string, header DBCHeader) ([]byte, error) {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for header update: %v", err)
	}
	defer file.Close()

	stringBlocks, err := file.Seek(-int64(header.StringBlockSize), io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to string block: %v", err)
	}

	StringBytes := make([]byte, header.StringBlockSize)
	_, err = file.ReadAt(StringBytes, stringBlocks)
	if err != nil {
		return nil, fmt.Errorf("failed to read string block: %v", err)
	}

	return StringBytes, nil
}

// Reader for durability costs costs from a DBC file
func ReadDurabilityCosts(filepath string) ([]DurabilityCost, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var durabilityCosts []DurabilityCost

	// Skip header (20 bytes)
	header := make([]byte, 20)
	if _, err := file.Read(header); err != nil {
		return nil, fmt.Errorf("failed to read header: %v", err)
	}

	recordCount := binary.LittleEndian.Uint32(header[4:8])
	fmt.Printf("Record count: %d\n", recordCount)

	for i := 0; i < int(recordCount); i++ {
		var cost DurabilityCost
		if err := binary.Read(file, binary.LittleEndian, &cost.ID); err != nil {
			return nil, fmt.Errorf("failed to read ID at record %d: %v", i, err)
		}
		fmt.Printf("Read ID: %d\n", cost.ID)

		for j := 0; j < len(cost.WeaponSubClassCost); j++ {
			if err := binary.Read(file, binary.LittleEndian, &cost.WeaponSubClassCost[j]); err != nil {
				return nil, fmt.Errorf("failed to read WeaponSubClassCost[%d] at record %d: %v", j, i, err)
			}
			fmt.Printf("Read WeaponSubClassCost[%d]: %d\n", j, cost.WeaponSubClassCost[j])
		}

		// Read ArmorSubClassCost
		for j := 0; j < len(cost.ArmorSubClassCost); j++ {
			if err := binary.Read(file, binary.LittleEndian, &cost.ArmorSubClassCost[j]); err != nil {
				return nil, fmt.Errorf("failed to read ArmorSubClassCost[%d] at record %d: %v", j, i, err)
			}
			fmt.Printf("Read ArmorSubClassCost[%d]: %d\n", j, cost.ArmorSubClassCost[j])
		}

		durabilityCosts = append(durabilityCosts, cost)
	}

	remainingBytes := make([]byte, 1024) // Read up to 1024 bytes beyond the expected records
	n, _ := file.Read(remainingBytes)
	if n > 0 {
		fmt.Printf("Extra data at the end of file (%d bytes): %X\n", n, remainingBytes[:n])
	}

	return durabilityCosts, nil
}

func WriteDurabilityCost(filepath string, cost DurabilityCost) error {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read the header
	header := make([]byte, 20)
	if _, err := file.Read(header); err != nil {
		return fmt.Errorf("failed to read header: %v", err)
	}

	// Update the record count in the header
	recordCount := binary.LittleEndian.Uint32(header[4:8])
	recordCount++
	binary.LittleEndian.PutUint32(header[4:8], recordCount)

	// Write the updated header back to the file
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to the beginning: %v", err)
	}
	if _, err := file.Write(header); err != nil {
		return fmt.Errorf("failed to write updated header: %v", err)
	}

	// Move to the end of the file to append the new record
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("failed to seek to the end: %v", err)
	}

	buffer := new(bytes.Buffer)
	if err := binary.Write(buffer, binary.LittleEndian, cost.ID); err != nil {
		return fmt.Errorf("failed to write ID: %v", err)
	}

	// Write WeaponSubClassCost in little-endian format (uint32)
	for _, value := range cost.WeaponSubClassCost {
		if err := binary.Write(buffer, binary.LittleEndian, value); err != nil {
			return fmt.Errorf("failed to write WeaponSubClassCost: %v", err)
		}
	}

	// Write ArmorSubClassCost in little-endian format (uint32)
	for _, value := range cost.ArmorSubClassCost {
		if err := binary.Write(buffer, binary.LittleEndian, value); err != nil {
			return fmt.Errorf("failed to write ArmorSubClassCost: %v", err)
		}
	}

	// Align the record size to the expected size in the DBC file
	// If the record size is less than the expected size, pad with zeros
	expectedRecordSize := 116
	actualRecordSize := buffer.Len()

	if actualRecordSize < expectedRecordSize {
		padding := make([]byte, expectedRecordSize-actualRecordSize)
		buffer.Write(padding)
	}

	// Write the new record to the file
	if _, err := file.Write(buffer.Bytes()); err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	return nil
}

func AppendDurabilityCostsToFile(filepath string, durabilityCosts []DurabilityCost) error {
	// Step 1: Open the file without O_APPEND to update the header
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for header update: %v", err)
	}
	defer file.Close()

	header, err := ReadDBCHeader(filepath)
	if err != nil {
		return fmt.Errorf("failed to read DBC header: %v", err)
	}

	header.RecordCount += uint32(len(durabilityCosts))

	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to find beginning of the file: %v", err)
	}

	err = binary.Write(file, binary.LittleEndian, &header)
	if err != nil {
		return fmt.Errorf("failed to write updated header: %v", err)
	}

	// get the string block offset to write
	offset, err := GetStringOffset(filepath, header)
	if err != nil {
		return fmt.Errorf("failed to get string block offset: %v", err)
	}

	savedStrBlock, err := GetStringBlock(filepath, header)
	if err != nil {
		return fmt.Errorf("failed to get string block: %v", err)
	}

	_, err = file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("failed to seek to string block: %v", err)
	}

	for _, cost := range durabilityCosts {
		if err := binary.Write(file, binary.LittleEndian, cost.ID); err != nil {
			return fmt.Errorf("failed to write ID: %v", err)
		}

		for _, weaponCost := range cost.WeaponSubClassCost {
			if err := binary.Write(file, binary.LittleEndian, weaponCost); err != nil {
				return fmt.Errorf("failed to write WeaponSubClassCost: %v", err)
			}
		}

		for _, armorCost := range cost.ArmorSubClassCost {
			if err := binary.Write(file, binary.LittleEndian, armorCost); err != nil {
				return fmt.Errorf("failed to write ArmorSubClassCost: %v", err)
			}
		}
	}

	// Write the string block back to the file
	_, err = file.Write(savedStrBlock)
	if err != nil {
		return fmt.Errorf("failed to write string block: %v", err)
	}

	return nil
}

func main() {

	costs, err := ReadDurabilityCosts("DurabilityCosts.dbc")
	if err != nil {
		log.Fatal(err)
	}

	Row300 := costs[len(costs)-1]
	toAdd := 450 // Number of new durability to add

	newRows := []DurabilityCost{}

	for i := 301; i <= toAdd; i++ {
		newRow := Row300
		newRow.ID = uint32(i)

		for j := 0; j < len(Row300.WeaponSubClassCost); j++ {
			newRow.WeaponSubClassCost[j] += uint32(20) * (uint32(i) - 300)
		}
		for j := 0; j < len(Row300.ArmorSubClassCost); j++ {
			newRow.ArmorSubClassCost[j] += uint32(10) * (uint32(i) - 300)
		}

		newRows = append(newRows, newRow)
	}

	AppendDurabilityCostsToFile("DurabilityCosts.dbc", newRows)
	ReadDurabilityCosts("DurabilityCosts.dbc")
}
