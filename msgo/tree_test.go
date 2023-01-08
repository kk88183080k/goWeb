package msgo

import (
	"fmt"
	"testing"
)

func TestTreeNode_Put_Get(t *testing.T) {

	root := TreeNode{name: "/"}

	//root.Put("/user/find/:1")
	//root.Put("/user/delete/:1")
	//root.Put("/user/info/:1")
	//root.Put("/user/search")
	root.Put("/")
	//root.Put("/index")
	root.Put("/user/*/get")

	node := root.Get("/")
	fmt.Println(node)

	//node = root.Get("/index")
	//fmt.Println(node)

	/*node = root.Get("/user/find/:1")
	fmt.Println(node)

	node = root.Get("/user/search")
	fmt.Println(node)

	node = root.Get("/user/search34")
	fmt.Println(node)
	*/
	node = root.Get("/user/1/get")
	fmt.Println(node)
}
