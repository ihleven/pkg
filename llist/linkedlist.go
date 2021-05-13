package llist

import "fmt"

func main() {
	list := NewList(1)
	list.Add(2)
	list.Add(6)
	list.Add(345)
	list.Add(4)
	list.Add(782)
	fmt.Printf("Hello world! [ %v \n", list)
}

type elem struct {
	next    *elem
	content interface{}
}

func (e elem) String() string {
	if e.next != nil {
		return fmt.Sprintf("%v, %v", e.content, e.next)
	}
	return fmt.Sprintf("%v ]", e.content)

}

func (e *elem) Add(a interface{}) {
	elem := elem{content: a, next: e.next}
	e.next = &elem
}
func NewList(a interface{}) *elem {
	return &elem{content: a}
}
