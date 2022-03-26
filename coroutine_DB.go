package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
	"sync"
)

type Data struct {
	x_lock X_Lock
	s_lock S_Lock
	val    int
}
type User struct { //include authenticity level and user_id
	status int //1 == free,2 == busy(wait for fin), 3 == block
	Id     int
	T      []Transaction
}
type DataTable struct {
	mu       sync.Mutex
	DataList []Data
	TableId  int
}

const (
	Line_level  = 1
	Page_level  = 2
	Table_level = 3
)

type Transaction struct {
	activity string
	val      int
	line     int
	table    int
}

//line level
type X_Lock struct {
	usr_id  int
	t_queue []int
}
type S_Lock struct {
	usr_id  []int
	t_queue []int
}

var isready chan int

func (usr *User) writeData(val int, t *DataTable, finished chan bool) bool {
	usr.status = 2

	//lock_table.mu.Lock()
	t.DataList = append(t.DataList, Data{X_Lock{}, S_Lock{}, val})
	usr.addLock(true, len(t.DataList)-1, t)

	fmt.Println("lock!")
	//defer lock_table.mu.Unlock()

	//finish
	fin := <-finished

	usr.deleteLock(true, len(t.DataList)-1, t)

	fmt.Println("Unlock!")

	usr.status = 1
	return fin

}

func (usr *User) editData(val, line, table int, t *DataTable, finished chan bool) bool {
	usr.status = 2
	//lock_table.mu.Lock()

	if !usr.checkLock(true, line, t) {
		usr.status = 3
		return false
	}

	t.mu.Lock()
	t.DataList[line] = Data{X_Lock{}, S_Lock{}, val}
	usr.addLock(true, line, t)
	t.mu.Unlock()

	//defer lock_table.mu.Unlock()

	fin := <-finished //block

	usr.deleteLock(true, line, t)
	usr.status = 1
	return fin
}

func (usr *User) deleteData(line int, t *DataTable) bool {
	usr.status = 2

	if line < 0 || line > cap(t.DataList) {
		fmt.Println("Warming: Invalid:line", line)
		return false
	}

	tmpList := t.DataList[:line-1]
	tmpList2 := t.DataList[line:]
	t.DataList = append(tmpList, tmpList2...)
	usr.status = 1

	return true
}

func (usr *User) readData(line, table int, t *DataTable, finished chan bool) bool {
	line -= 1
	//mutex for read LockTable
	usr.status = 2

	if !usr.checkLock(false, line, t) {
		usr.status = 3
		return false
	}
	usr.addLock(false, line, t)
	val := t.DataList[line].val

	fin := <-finished //block

	usr.deleteLock(false, line, t)

	fmt.Println("user:", usr.Id, " read data:  ", val)
	usr.status = 1
	return fin
}

func Controller() {
	//1. initial

	isready = make(chan int)
	Finished := [3]chan bool{
		make(chan bool),
		make(chan bool),
		make(chan bool),
	}

	Usr_list := []User{{2, 0, scanText("t1.txt")}, {2, 1, scanText("t2.txt")}, {2, 2, scanText("t3.txt")}}
	t := []DataTable{{sync.Mutex{}, []Data{}, 0}, {sync.Mutex{}, []Data{}, 1}, {sync.Mutex{}, []Data{}, 2}}
	fmt.Println("start DBS")

	for usr := range Usr_list { //start coroutine,each coroutine represents 1 user
		go Usr_list[usr].run(Finished[usr], &t)
	}

	go ready(Finished)

	for {
		//time.Sleep((time.Second * 1))

		var fin_usr int

		_, err := fmt.Scan(&fin_usr)
		if err == io.EOF || fin_usr == -1 {
			break
		}

		if fin_usr > 2 || fin_usr < 0 {
			fmt.Println("Error: User ", fin_usr, " does not exist")
			break
		}

		switch Usr_list[fin_usr].status {
		case 1:
			fmt.Println("Error: User ", fin_usr, " is free")
		case 2:
			Finished[fin_usr] <- true
			fmt.Println("Finish ->: ", fin_usr)
		case 3:
			fmt.Println("Error: User ", fin_usr, " is block")
		default:
			fmt.Println("status 0")
		}
	}

}

func ready(fin [3]chan bool) {
	id := <-isready
	fmt.Println("ready <- ", id)
	fin[id] <- true

}
func scanText(filename string) []Transaction {
	var t []Transaction
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		var temp_t Transaction

		s, err := reader.ReadString(' ')
		if err == nil {
			switch s {
			case "append ":
				temp_t.activity = "append"

			case "edit ":
				temp_t.activity = "edit"

			case "read ":
				temp_t.activity = "read"
			}
		} else if err == io.EOF {
			break
		} else {
			fmt.Println("error in scan text")
		}

		s, err = reader.ReadString(' ')
		s = s[0 : len(s)-1]
		if err == nil {
			temp_t.val, _ = strconv.Atoi(s)
		}

		s, err = reader.ReadString(' ')
		s = s[0 : len(s)-1]
		if err == nil {
			temp_t.line, _ = strconv.Atoi(s)
		}

		s, err = reader.ReadString('\n')
		s = s[0 : len(s)-1]
		if err == nil {
			temp_t.table, _ = strconv.Atoi(s)
		}
		t = append(t, temp_t)

	}

	return t

}

func (usr *User) run(fin chan bool, t *[]DataTable) bool {
	finished := <-fin
	fmt.Println("User ", usr.Id, " start running")
	for len(usr.T) != 0 {
		if usr.status != 3 {
			inst := usr.T[0].activity
			val := usr.T[0].val
			d := usr.T[0].line
			ind_of_tb := usr.T[0].table

			var taskover bool
			switch inst {
			case "append":
				taskover = usr.writeData(d, &((*t)[ind_of_tb]), fin)

			case "read":
				taskover = usr.readData(d, ind_of_tb, &((*t)[ind_of_tb]), fin)

			case "edit":
				taskover = usr.editData(val, d, ind_of_tb, &((*t)[ind_of_tb]), fin)

			}

			if taskover {
				usr.T = usr.T[1:len(usr.T)]
			}

			//time.Sleep(1 * time.Second)
		} else {
			s := <-fin
			fmt.Println(s)
			usr.status = 1
		}
	}

	fmt.Println("user: ", usr.Id, " done!")
	return finished
}
func (usr *User) addLock(isedit bool, line int, t *DataTable) {

	if !isedit {
		t.DataList[line].s_lock.usr_id = append(t.DataList[line].s_lock.usr_id, usr.Id)
	} else {
		t.DataList[line].x_lock = X_Lock{usr.Id, []int{}}
	}

}

func (usr *User) checkLock(isedit bool, line int, t *DataTable) bool {

	if isedit {
		if len(t.DataList[line].s_lock.usr_id) != 0 {
			fmt.Println("block : s_lock")
			t.DataList[line].s_lock.t_queue = append(t.DataList[line].s_lock.t_queue, usr.Id)
			return false
		}
	}

	if !reflect.DeepEqual(t.DataList[line].x_lock, X_Lock{}) {
		fmt.Println("block:x_lock")
		t.DataList[line].x_lock.t_queue = append(t.DataList[line].x_lock.t_queue, usr.Id)
		return false
	}

	return true
}

func (usr *User) deleteLock(isedit bool, line int, t *DataTable) {

	if !isedit {

		for i, s_lock := range t.DataList[line].s_lock.usr_id {
			if s_lock == usr.Id {
				t.DataList[line].s_lock.usr_id = append(t.DataList[line].s_lock.usr_id[0:i], t.DataList[line].s_lock.usr_id[i+1:]...)

				if len(t.DataList[line].s_lock.t_queue) > 0 {
					ready_tmp := t.DataList[line].s_lock.t_queue[0]
					defer func() {
						isready <- ready_tmp
					}()

					t.DataList[line].s_lock.t_queue = t.DataList[line].s_lock.t_queue[1:]

				}
			}
		}

	} else {
		if len(t.DataList[line].x_lock.t_queue) > 0 {
			ready_tmp := t.DataList[line].x_lock.t_queue[0]
			defer func() {
				isready <- ready_tmp
			}()
		}
		fmt.Println("delete x_lock")
		t.DataList[line].x_lock = X_Lock{}
	}
}

func main() {
	Controller()
}
