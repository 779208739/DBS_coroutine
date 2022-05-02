// package models

// import (
// )

// type Lock_Info struct {
// 	T     string `json:"title"`
// 	Line  int    `json:"line"`
// 	Table int    `json:"table"`
// }

// func GetLock() []Lock_Info {
// 	Lockinfo := make([]Lock_Info, 0)
// 	for lock := range controller.Xlock_Table {
// 		Lockinfo = append(Lockinfo, Lock_Info{T: "互斥锁", Line: lock.Line, Table: lock.Table})
// 	}

// 	for lock := range controller.Slock_Table {
// 		Lockinfo = append(Lockinfo, Lock_Info{T: "共享锁", Line: lock.Line, Table: lock.Table})
// 	}

// 	return Lockinfo
// }

// func GetUsr(usr int) []controller.Transaction {
// 	t := make([]controller.Transaction, 0)

// 	switch usr {
// 	case 1:
// 		t = append(t, controller.ScanText("txt/t1.txt")...)
// 	case 2:
// 		t = append(t, controller.ScanText("txt/t2.txt")...)
// 	case 3:
// 		t = append(t, controller.ScanText("txt/t3.txt")...)
// 	}

// 	return t
// }
