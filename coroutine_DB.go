package main

import (
	"fmt"
)

type Data struct {
	val int
}
type DataTable struct {
	DataList []Data
}
type Lock struct {
	mode int
}
type LockTable struct {
	Table []Lock
}

func (t *DataTable) createData(val int) {
	d := Data{val}

	t.DataList = append(t.DataList, d)

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
func main() {
	var Dt DataTable
	Dt.createData(222)
	Dt.createData(233)
	Dt.createData(244)
	Dt.deleteData(10)
	fmt.Println(Dt.DataList)

}
