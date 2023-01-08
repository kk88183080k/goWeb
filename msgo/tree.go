package msgo

import (
	"strings"
)

// TreeNode / 为根节点
// 以/user/login 为例
//
//	(第一级)
//
// |---user (第二级)
// |--------login (第三级)
type TreeNode struct {
	name           string
	child          []*TreeNode
	routerFullPath string
	leaf           bool
}

// Put path  example
//
//	/user/login
//	/user/logout
//	/user/info
//	/user/get/:id
//	/user/del/:id
//	/user/update/:id
//	/user/search
//	/user/**
func (root *TreeNode) Put(path string) {

	temp := root

	//if root.child == nil {
	//	root.child = []&TreeNode{}
	//}

	pathSub := strings.Split(path, "/")
	for i, v := range pathSub {

		if i == 0 {
			continue
		}

		// 判断子节点中是否存在
		exists := false
		for _, tNode := range root.child {
			if tNode.name == v {
				exists = true
				root = tNode
				break
			}
		}

		// 当子节点中没有时添加
		if !exists {
			node := &TreeNode{v, make([]*TreeNode, 0), "", false}
			if i == len(pathSub)-1 {
				node.leaf = true
			}
			child := append(root.child, node)
			root.child = child
			root = node
		}
	}

	root = temp
}

// Get  /user/login/
func (t *TreeNode) Get(path string) *TreeNode {
	temp := t
	pathSub := strings.Split(path, "/")
	rs := ""
	for i, v := range pathSub {
		if i == 0 {
			continue
		}

		exists := false
		for _, node := range t.child {
			if v == node.name || "*" == node.name || strings.Contains(node.name, ":") {
				exists = true
				rs += "/" + node.name
				t = node
				node.routerFullPath = rs
				//log.Printf("routerFullPath:%s \n", node.routerFullPath)
				if i == len(pathSub)-1 {
					return node
				}
				break
			}
		}

		if !exists {
			for _, node1 := range t.child {
				if node1.name == "**" {
					node1.routerFullPath = rs + "/" + node1.name
					//log.Printf("routerFullPath not exists:%s \n", node1.routerFullPath)
					return node1
				}
			}
		}
	}
	t = temp
	return nil
}
