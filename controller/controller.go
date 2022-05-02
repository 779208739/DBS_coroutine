package controller

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

type AllInfo struct {
	U []User
	L []Lock_Info
}

type Data struct {
	x_lock X_Lock
	s_lock S_Lock
	val    int
}
type User struct { //include authenticity level and user_id
	Status int `json:"status,string"`
	//1 == free,2 == busy(wait for fin), 3 == block
	Id  int
	Act int `json:"Act"`
	T   []TransAction
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

type TransAction struct {
	Activity string `json:"act"`
	Val      int    `json:"val"`
	Line     int    `json:"line"`
	Table    int    `json:"table"`
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
type map_key struct {
	Table int
	Line  int
}
type Lock_Info struct {
	T     string `json:"title"`
	Line  int    `json:"line"`
	Table int    `json:"table"`
}

var Slock_Table map[map_key]S_Lock
var Xlock_Table map[map_key]X_Lock
var isready chan int
var isfresh chan bool
var Usr_list []User
var waitlist []int

func (usr *User) writeData(val int, t *DataTable, finished chan bool) bool {
	usr.Status = 2

	//lock_table.mu.Lock()
	t.DataList = append(t.DataList, Data{X_Lock{}, S_Lock{}, val})
	usr.addLock2(true, len(t.DataList)-1, t.TableId)

	//defer lock_table.mu.Unlock()

	//finish
	fin := <-finished

	usr.deleteLock2(true, len(t.DataList)-1, t.TableId)

	fmt.Println("Unlock!")

	usr.Status = 1
	return fin

}

func (usr *User) editData(val, line, table int, t *DataTable, finished chan bool) bool {
	usr.Status = 2
	//lock_table.mu.Lock()

	if !usr.checkLock2(true, t.TableId, line) {
		usr.Status = 3
		return false
	}

	t.mu.Lock()
	t.DataList[line] = Data{X_Lock{}, S_Lock{}, val}
	usr.addLock2(true, line, t.TableId)
	t.mu.Unlock()

	//defer lock_table.mu.Unlock()

	fin := <-finished //block

	usr.deleteLock2(true, line, t.TableId)
	usr.Status = 1
	return fin
}

func (usr *User) deleteData(line int, t *DataTable) bool {
	usr.Status = 2

	if !usr.checkLock2(true, t.TableId, line) {
		usr.Status = 3
		return false
	}
	if line < 0 || line > cap(t.DataList) {
		fmt.Println("Warming: Invalid:line", line)
		return false
	}

	tmpList := t.DataList[:line-1]
	tmpList2 := t.DataList[line:]
	t.DataList = append(tmpList, tmpList2...)
	usr.Status = 1

	return true
}

func (usr *User) readData(line, table int, t *DataTable, finished chan bool) bool {
	line -= 1
	var val int
	usr.Status = 2

	if !usr.checkLock2(false, t.TableId, line) {
		usr.Status = 3
		fmt.Println(usr.Id, " Block!")
		return false
	}
	usr.addLock2(false, line, t.TableId)
	if table > 3 || len(t.DataList) > line+1 {
		val = -100
	} else {
		val = t.DataList[line].val
	}

	fin := <-finished //block

	usr.deleteLock2(false, line, t.TableId)

	fmt.Println("user:", usr.Id, " read data:  ", val)
	usr.Status = 1
	return fin
}

func Controller(fin chan int, info *AllInfo) {
	//1. initial

	isready = make(chan int)
	isfresh = make(chan bool)
	Finished := [3]chan bool{
		make(chan bool),
		make(chan bool),
		make(chan bool),
	}
	Xlock_Table = make(map[map_key]X_Lock)
	Slock_Table = make(map[map_key]S_Lock)

	Usr_list = make([]User, 3)
	Usr_list[0] = User{4, 0, 0, ScanText("txt/t1.txt")}
	Usr_list[1] = User{4, 1, 0, ScanText("txt/t2.txt")}
	Usr_list[2] = User{4, 2, 0, ScanText("txt/t3.txt")}
	//Usr_list := []User{{2, 0, 0, ScanText("txt/t1.txt")}, {2, 1, 0, ScanText("txt/t2.txt")}, {2, 2, 0, ScanText("txt/t3.txt")}}
	t := []DataTable{{sync.Mutex{}, []Data{}, 0}, {sync.Mutex{}, []Data{}, 1}, {sync.Mutex{}, []Data{}, 2}}
	fmt.Println("start DBS")

	go Usr_list[0].run(Finished[0], &t)
	go Usr_list[1].run(Finished[1], &t)
	go Usr_list[2].run(Finished[2], &t)

	go ready(Finished)
	go fresh(info)
	isfresh <- true
	for {
		var fin_usr int

		fin_usr = <-fin

		if fin_usr > 2 || fin_usr < 0 {
			fmt.Println("Error: User ", fin_usr, " does not exist")
			break
		}

		switch Usr_list[fin_usr].Status {
		case 1:
			fmt.Println("Error: User ", fin_usr, " is free")
		case 2, 4:
			Finished[fin_usr] <- true
			fmt.Println("Finish ->: ", fin_usr)
		case 3:
			fmt.Println("Error: User ", fin_usr, " is block")
		default:
			fmt.Println("Status 0")
		}
	}

}

func ready(fin [3]chan bool) {
	id := <-isready
	fmt.Println("ready <- ", id)
	fin[id] <- true

}
func ScanText(filename string) []TransAction {
	var t []TransAction
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		var temp_t TransAction

		s, err := reader.ReadString(' ')
		if err == nil {
			switch s {
			case "append ":
				temp_t.Activity = "append"

			case "edit ":
				temp_t.Activity = "edit"

			case "read ":
				temp_t.Activity = "read"
			}
		} else if err == io.EOF {
			break
		} else {
			fmt.Println("error in scan text")
		}

		s, err = reader.ReadString(' ')
		s = s[0 : len(s)-1]
		if err == nil {
			temp_t.Val, _ = strconv.Atoi(s)
		}

		s, err = reader.ReadString(' ')
		s = s[0 : len(s)-1]
		if err == nil {
			temp_t.Line, _ = strconv.Atoi(s)
		}

		s, err = reader.ReadString('\n')
		s = s[0 : len(s)-1]
		if err == nil {
			temp_t.Table, _ = strconv.Atoi(s)
		}
		t = append(t, temp_t)

	}

	return t

}

func (usr *User) run(fin chan bool, t *[]DataTable) bool {
	finished := <-fin
	fmt.Println("User ", usr.Id, " start running")
	for len(usr.T) != 0 {
		if usr.Status != 3 {
			inst := usr.T[0].Activity
			val := usr.T[0].Val
			d := usr.T[0].Line
			ind_of_tb := usr.T[0].Table

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
				usr.Act++
			}

			//time.Sleep(1 * time.Second)
		} else {
			s := <-fin
			fmt.Println(s)
			usr.Status = 1
		}

		isfresh <- true
	}

	fmt.Println("user: ", usr.Id, " done!")
	return finished
}
func (usr *User) addLock(isedit bool, line int, t *DataTable) {

	fmt.Println("Add lock!  User:", usr.Id)
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

func (usr *User) addLock2(isedit bool, line, table int) {
	defer func() {
		isfresh <- true
	}()

	fmt.Println("Lock! ", usr.Id)
	key := map_key{table, line}
	if isedit {
		Xlock_Table[key] = X_Lock{usr.Id, waitlist}
		waitlist = nil
	} else {
		_, ok := Slock_Table[key]
		if ok {
			usr_id := append(Slock_Table[key].usr_id, usr.Id)
			t_queue := Slock_Table[key].t_queue

			Slock_Table[key] = S_Lock{usr_id, t_queue}
		} else {
			usr_id := []int{usr.Id}
			t_queue := Slock_Table[key].t_queue

			Slock_Table[key] = S_Lock{usr_id, t_queue}

		}
	}

}

func (usr *User) deleteLock2(isedit bool, line, table int) {
	defer func() {
		isfresh <- true
	}()

	key := map_key{table, line}
	if isedit {
		if len(Xlock_Table[key].t_queue) == 0 {
			delete(Xlock_Table, key)
		} else {
			t_read, t_edit := dispatchAfterDelete(Xlock_Table[key].t_queue)

			fmt.Println("t_read", t_read)
			defer func() {
				for i := range t_read {
					isready <- t_read[i]
				}
			}()

			if len(t_read) == 0 {
				usr := Xlock_Table[key].t_queue[0]
				t := Xlock_Table[key].t_queue[1:]

				l := X_Lock{usr, t}
				Xlock_Table[key] = l
			} else {
				waitlist = append(waitlist, t_edit...)
				delete(Xlock_Table, key)
			}

		}
	} else {
		if len(Slock_Table[key].t_queue) == 0 && len(Slock_Table[key].usr_id) == 1 {
			delete(Slock_Table, key)
			fmt.Println("delete lock")
		} else {
			if len(Slock_Table[key].usr_id) != 1 {
				var usr_id []int
				for index := range Slock_Table[key].usr_id {
					if Slock_Table[key].usr_id[index] == usr.Id {
						usr_id = append(Slock_Table[key].usr_id[0:index], Slock_Table[key].usr_id[index+1:]...)
						break
					}
				}

				Slock_Table[key] = S_Lock{usr_id, Slock_Table[key].t_queue}
			} else {
				ready_tmp := Slock_Table[key].t_queue[0]
				defer func() {
					isready <- ready_tmp
					delete(Slock_Table, key)
				}()
				if len(Slock_Table[key].t_queue) > 1 {
					waitlist = append(waitlist, Slock_Table[key].t_queue[1:]...)
				}
			}
		}
	}

}

//after deleting Xlock, goroutines which would read in waitlist will work together
func dispatchAfterDelete(u []int) ([]int, []int) {
	var ans1, ans2 []int
	for i := range u {
		if Usr_list[u[i]].T[0].Activity == "read" {
			ans1 = append(ans1, u[i])
		} else {
			ans2 = append(ans2, u[i])
		}
	}

	return ans1, ans2

}
func (usr *User) checkLock2(isedit bool, table, line int) bool {
	defer func() {
		isfresh <- true
	}()
	key := map_key{table, line}

	if isedit {
		_, ok := Xlock_Table[key]
		_, ok2 := Slock_Table[key]
		if ok {
			t := append(Xlock_Table[key].t_queue, usr.Id)
			l := X_Lock{Xlock_Table[key].usr_id, t}
			Xlock_Table[key] = l
			return false
		}
		if ok2 {
			t := append(Slock_Table[key].t_queue, usr.Id)
			l := S_Lock{Slock_Table[key].usr_id, t}
			Slock_Table[key] = l
			return false
		}
	} else {
		_, ok := Xlock_Table[key]
		if ok {
			t := append(Xlock_Table[key].t_queue, usr.Id)
			l := X_Lock{Xlock_Table[key].usr_id, t}

			fmt.Println("t_queue:", t)
			Xlock_Table[key] = l
			return false
		}
	}

	return true
}

func GetLock() []Lock_Info {
	Lockinfo := make([]Lock_Info, 0)
	for lock := range Xlock_Table {
		Lockinfo = append(Lockinfo, Lock_Info{T: "互斥锁", Line: lock.Line, Table: lock.Table})
	}

	for lock := range Slock_Table {
		Lockinfo = append(Lockinfo, Lock_Info{T: "共享锁", Line: lock.Line, Table: lock.Table})
	}

	return Lockinfo
}

func GetUsr(usr int) []TransAction {
	t := make([]TransAction, 0)

	switch usr {
	case 1:
		t = append(t, ScanText("txt/t1.txt")...)
	case 2:
		t = append(t, ScanText("txt/t2.txt")...)
	case 3:
		t = append(t, ScanText("txt/t3.txt")...)
	}

	return t
}

func fresh(a *AllInfo) {
	for true {
		_ = <-isfresh
		a.L = GetLock()
		U_tmp := make([]User, 0)
		for i := range Usr_list {
			U_tmp = append(U_tmp, User{Usr_list[i].Status, Usr_list[i].Id, Usr_list[i].Act, Usr_list[i].T})
		}
		a.U = U_tmp
	}
}
