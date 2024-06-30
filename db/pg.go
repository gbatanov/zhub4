package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type MapAddr struct {
	ShortAddr uint16
	MacAddr   uint64
}

type PostgresDB struct {
	Db *sql.DB
}

func (db PostgresDB) CloseDb() {
	db.Db.Close()
}

func OpenDb() (*PostgresDB, error) {
	host := "192.168.88.82"
	port := 5432
	user := "postgres"
	password := "12345678"
	dbname := "zhub4"

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return &PostgresDB{Db: nil}, err
	}
	return &PostgresDB{Db: db}, nil
}

// Загружаем маппинг адресов из БД
func (db PostgresDB) ReadMap(devicessAddressMap *map[uint16]uint64) error {

	rows, err := db.Db.Query(`SELECT short_addr, mac_addr  FROM "map_addr"`)
	if err != nil {
		log.Printf("%s\n", err.Error())
		return err
	}

	defer rows.Close()
	for rows.Next() {

		var short_addr string
		var mac_addr string

		err = rows.Scan(&short_addr, &mac_addr)
		if err != nil {
			log.Printf("%s\n", err.Error())
			return err
		}
		var shortAddr uint16
		var macAddr uint64
		fd := short_addr + " " + mac_addr
		r, err := fmt.Sscanf(fd, "%4x %16x\n", &shortAddr, &macAddr)
		if err == nil && r > 0 {
			(*devicessAddressMap)[shortAddr] = macAddr
		}
	}

	return nil
}

// Полностью перезаписываем таблицу map_addr значениями из файла
// Записи в таблице полносттью аналогичны записям в файле - формат 16-титеричных чисел
func (db PostgresDB) ImportMap(devicessAddressMap *map[uint16]uint64) error {
	_, err := db.Db.Exec(`DROP TABLE IF EXISTS public.map_addr;
CREATE TABLE IF NOT EXISTS public.map_addr
(
    short_addr char(4)   NOT NULL,
    mac_addr char(16)   NOT NULL
)
TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.map_addr  OWNER to postgres;`)
	if err != nil {
		return err
	}

	for a, b := range *devicessAddressMap {
		err = db.InsertIntoMap(a, b)
		if err != nil {
			return err
		}
	}
	return nil
}

// Вставка одной записи в мап
func (db PostgresDB) InsertIntoMap(shortAddr uint16, macAddr uint64) error {
	as := fmt.Sprintf("%04x", shortAddr)
	bs := fmt.Sprintf("%016x", macAddr)

	_, err := db.Db.Exec(fmt.Sprintf(`INSERT INTO map_addr VALUES ('%s','%s')`, as, bs))

	return err

}

// Замена короткого адреса для одной записи по мак-адресу
func (db PostgresDB) UpdateMap(shortAddr uint16, macAddr uint64) error {
	as := fmt.Sprintf("%04x", shortAddr)
	bs := fmt.Sprintf("%016x", macAddr)
	_, err := db.Db.Exec(fmt.Sprintf(`UPDATE map_addr SET short_addr='%s' WHERE mac_addr = '%s'`, as, bs))

	return err
}

/*
func funcDb() {
	host := "192.168.88.82"
	port := 5432
	user := "postgres"
	password := "12345678"
	dbname := "zhub4"

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	// close database
	defer db.Close()

	// check db
	err = db.Ping()
	if err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	fmt.Println("Connected!")

	rows, err := db.Query(`SELECT count("id")  FROM "events"`)
	if err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	defer rows.Close()
	for rows.Next() {

		var cnt int

		err = rows.Scan(&cnt)
		if err != nil {
			log.Printf("%s\n", err.Error())
			return
		}

		fmt.Println(cnt)
	}

	if err != nil {
		log.Printf("%s\n", err.Error())
		return
	}
}
*/
