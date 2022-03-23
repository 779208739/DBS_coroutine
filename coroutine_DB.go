package main

import (
	"fmt"
	"io"
	"sync"
)

type Data struct {
	val int
}
type User struct { //include authenticity level and user_id
	busy bool
	Id   int
}
type DataTable struct {
	DataList []Data
	TableId  int
}

const (
	Line_level  = 1
	Page_level  = 2
	Table_level = 3
)

type X_Lock struct {
	level  int // 1==line_level,2==page_level,3==table_level
	usr_id int
	line   int
	table  int
}
type S_Lock struct {
	level  int
	usr_id int
	line   int
	table  int
}
type LockTable struct {
	mu       sync.Mutex
	XL_Table []X_Lock
	SL_Table []S_Lock
}

func (usr *User) writeData(val int, t *DataTable, lock_table *LockTable, finished chan bool) bool {

	//lock_table.mu.Lock()

	t.DataList = append(t.DataList, Data{val})

	lock_table.XL_Table = append(lock_table.XL_Table, X_Lock{1, usr.Id, len(t.DataList), t.TableId})

	fmt.Println("lock!")
	//defer lock_table.mu.Unlock()

	//finish
	finished <- true
	fmt.Println("Unlock!")

	return true

}

func (usr *User) editData(val, line, table int, lock_table *LockTable, t *DataTable, finished chan bool) bool {
	lock_table.mu.Lock()
	for n, i := range lock_table.XL_Table {
		for i.table == table && line == i.line && i.usr_id != usr.Id {
			fmt.Println("mutex in X_Lock:line ", n)
			defer lock_table.mu.Unlock()
			return false
		}
	}

	for n, i := range lock_table.SL_Table {
		for i.table == table && line == i.line && i.usr_id != usr.Id {
			fmt.Println("mutex in S_Lock:line ", n)
			defer lock_table.mu.Unlock()
			return false
		}
	}

	l_tmp := X_Lock{1, usr.Id, line, table}
	lock_table.XL_Table = append(lock_table.XL_Table, l_tmp)
	t.DataList[line] = Data{val}
	defer lock_table.mu.Unlock()

	fin := <-finished //block
	return fin
}

func (t *DataTable) deleteData(line int) bool {

	if line < 0 || line > cap(t.DataList) {
		fmt.Println("Warming: Invalid:line", line)
		return false
	}

	tmpList := t.DataList[:line-1]
	tmpList2 := t.DataList[line:]
	t.DataList = append(tmpList, tmpList2...)

	return true
}

func (usr *User) readData(line, table int, lock_table *LockTable, t *DataTable) bool {
	//mutex for read LockTable

	//lock_table.mu.Lock()
	for n, i := range lock_table.XL_Table {
		for i.table == table && line == i.line && i.usr_id != usr.Id {
			fmt.Println("mutex in X_Lock:line ", n+1, " table: ", table)
			//defer lock_table.mu.Unlock()
			return false
		}

	}

	fmt.Println("data:  ", t.DataList[line-1])
	//defer lock_table.mu.Unlock()

	return true
}

func printLockTable(lock_table LockTable) {
	fmt.Println("show X_LOCK:", "\r")
	for _, i := range lock_table.XL_Table {
		fmt.Println("user_id:", i.usr_id, "  table:", i.table, " line:", i.line, "\r")
	}

	fmt.Println("\r", "show S_LOCK:", "\r")
	for _, i := range lock_table.SL_Table {
		fmt.Println("user_id:", i.usr_id, "  table:", i.table, " line:", i.line)
	}
}

func Controller() {
	//1. initial

	Finished := [3]chan bool{
		make(chan bool),
		make(chan bool),
		make(chan bool),
	}
	Usr_list := []User{{false, 0}, {false, 1}, {false, 2}}
	var lock_table LockTable
	t := []DataTable{{[]Data{}, 0}, {[]Data{}, 1}, {[]Data{}, 2}}

	fmt.Println(len(t))
	for {
		var usr int
		var inst string
		_, err := fmt.Scan(&usr) //input user

		if err == io.EOF || usr == -1 {
			break
		}

		if Usr_list[usr].busy { //detect whether the user is free
			fmt.Println("User ", usr, " is busy")
			continue
		}

		if usr >= 3 {
			fmt.Println("invalid usr")
			continue
		}

		_, err = fmt.Scan(&inst) //input insructions

		if err == io.EOF {
			break
		}

		var val int
		var d int
		var ind_of_tb int
		switch inst {

		case "fin":
			fin := <-Finished[usr]
			fmt.Println(fin)

		case "append":
			fmt.Println("input data & index of DataTable", "\r")
			fmt.Scan(&d, &ind_of_tb)
			go Usr_list[usr].writeData(d, &t[ind_of_tb], &lock_table, Finished[usr])

		case "read":
			fmt.Println("input line & index of DataTable")
			fmt.Scan(&d, &ind_of_tb)
			go Usr_list[usr].readData(d, ind_of_tb, &lock_table, &t[ind_of_tb])

		case "edit":
			fmt.Println("input var & line & index of DataTable")
			fmt.Scan(&val, &d, &ind_of_tb)
			go Usr_list[usr].editData(val, d, ind_of_tb, &lock_table, &t[ind_of_tb], Finished[usr])

		case "printLock":
			fmt.Println("print lock")
			go printLockTable(lock_table)

		default:
			fmt.Println("invalid instruction")
		}

	}
}

func main() {
	Controller()
}
