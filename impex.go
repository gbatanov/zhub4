package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gbatanov/zhub4/db"
	"github.com/gbatanov/zhub4/zigbee"
)

func BdImportFromFile() {
	log.Println("import")
	// Открываем файл, считываем в мап, закрываем файл
	// Открываем БД, записываем в таблицу, закрываем БД
	db, err := db.OpenDb()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer db.CloseDb()

	testMap, err := readMapFromFile()
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	err = db.ImportMap(&testMap)
	if err != nil {
		log.Println(err)
		os.Exit(3)
	}
	log.Println("import success")
}

// Читаем из базы и пишем в файл
func BdExportToFile() {
	log.Println("export")
	db, err := db.OpenDb()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer db.CloseDb()
	addressMap := make(map[uint16]uint64, 64) // Резервируем на 64 устройства

	err = db.ReadMap(&addressMap)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	filename := `map_addr_copy.cfg`

	fd, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	} else {
		for a, b := range addressMap {
			fmt.Fprintf(fd, "%04x %016x\n", a, b)
			fmt.Printf("%04x %016x\n", a, b)
		}
		fd.Sync()
		fd.Close()
	}
	log.Println(`Можно проверить новый файл и заменить при необходимости текущий`)
}

// Читаем  из файла в мап адресов
func readMapFromFile() (map[uint16]uint64, error) {

	zigbee.MapFileMutex.Lock()
	defer zigbee.MapFileMutex.Unlock()
	addressMap := make(map[uint16]uint64, 64) // Резервируем на 64 устройства

	filename := `map_addr.cfg` // Читаем из сохраненной копии, не из рабочей

	fd, err := os.OpenFile(filename, os.O_RDONLY, 0755)
	if err != nil {
		log.Println("ReadMap:: OpenFile error: ", err)
		return addressMap, err
	} else {

		var shortAddr uint16
		var macAddr uint64
		var r int
		var err error = nil
		for err == nil {
			r, err = fmt.Fscanf(fd, "%4x %16x\n", &shortAddr, &macAddr)
			if r > 0 {
				addressMap[shortAddr] = macAddr
			}
		}
		fd.Close()
		for a, b := range addressMap {
			log.Printf("%d: %d \n", a, b)
		}

	}

	return addressMap, nil
}
