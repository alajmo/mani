package components

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TTree struct {
	Tree     *tview.TreeView
	Root     *tview.Flex
	RootNode *tview.TreeNode
	Filter   *tview.InputField

	List          []*TNode
	Title         string
	RootTitle     string
	FilterValue   *string
	SelectEnabled bool

	IsNodeSelected   func(name string) bool
	ToggleSelectNode func(name string)
	SelectAll        func()
	UnselectAll      func()
	FilterNodes      func()
	DescribeNode     func(name string)
	EditNode         func(name string)
}

type TNode struct {
	ID          string // The reference
	DisplayName string // What is shown
	Type        string

	TreeNode *tview.TreeNode
	Children *[]TNode
}

func (t *TTree) Create() {
	title := misc.Colorize(t.RootTitle, *misc.TUITheme.Item)
	rootNode := tview.NewTreeNode(title)
	rootNode.SetColor(misc.STYLE_DEFAULT.Fg)
	rootNode.SetSelectable(false)

	t.IsNodeSelected = func(name string) bool { return false }
	t.ToggleSelectNode = func(name string) {}
	t.SelectAll = func() {}
	t.UnselectAll = func() {}
	t.FilterNodes = func() {}
	t.DescribeNode = func(name string) {}
	t.EditNode = func(name string) {}

	tree := tview.NewTreeView().
		SetRoot(rootNode).
		SetCurrentNode(rootNode)
	tree.SetGraphics(true)

	filter := CreateFilter()

	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tree, 0, 1, true).
		AddItem(filter, 1, 0, false)
	root.SetTitleAlign(misc.STYLE_TITLE.Align).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)

	t.Root = root
	t.Filter = filter
	t.RootNode = rootNode
	t.Tree = tree

	if t.Title != "" {
		title := misc.Colorize(t.Title, *misc.TUITheme.Title)
		t.Root.SetTitle(title)
	}

	// Methods
	t.IsNodeSelected = func(name string) bool { return false }

	// Filter
	t.Filter.SetChangedFunc(func(_ string) {
		t.applyFilter()
		t.FilterNodes()
	})

	t.Filter.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentFocus := misc.App.GetFocus()
		if currentFocus == filter {
			switch event.Key() {
			case tcell.KeyEscape:
				t.ClearFilter()
				t.FilterNodes()
				misc.App.SetFocus(tree)
				return nil
			case tcell.KeyEnter:
				t.applyFilter()
				t.FilterNodes()
				misc.App.SetFocus(tree)
			}
			return event
		}
		return event
	})

	// Input
	t.Tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if t.SelectEnabled {
				node := t.Tree.GetCurrentNode()
				name := node.GetReference().(string)
				t.ToggleSelectNode(name)
			}
		case tcell.KeyCtrlD:
			current := t.Tree.GetCurrentNode()
			_, _, _, height := t.Tree.GetInnerRect()
			visibleNodes := t.getVisibleNodes()
			currentIndex := t.findNodeIndex(visibleNodes, current)
			newIndex := min(currentIndex+height/2, len(visibleNodes)-1)
			if newIndex > 0 && newIndex < len(visibleNodes) {
				t.Tree.SetCurrentNode(visibleNodes[newIndex])
			}
			return nil
		case tcell.KeyCtrlU:
			current := t.Tree.GetCurrentNode()
			_, _, _, height := t.Tree.GetInnerRect()
			visibleNodes := t.getVisibleNodes()
			currentIndex := t.findNodeIndex(visibleNodes, current)
			newIndex := max(currentIndex-height/2, 0)
			if newIndex >= 0 && newIndex < len(visibleNodes) {
				t.Tree.SetCurrentNode(visibleNodes[newIndex])
			}
			return nil
		case tcell.KeyCtrlF:
			current := t.Tree.GetCurrentNode()
			_, _, _, height := t.Tree.GetInnerRect()
			visibleNodes := t.getVisibleNodes()
			currentIndex := t.findNodeIndex(visibleNodes, current)
			newIndex := min(currentIndex+height, len(visibleNodes)-1)
			if newIndex > 0 && newIndex < len(visibleNodes) {
				t.Tree.SetCurrentNode(visibleNodes[newIndex])
			}
			return nil
		case tcell.KeyCtrlB:
			current := t.Tree.GetCurrentNode()
			_, _, _, height := t.Tree.GetInnerRect()
			visibleNodes := t.getVisibleNodes()
			currentIndex := t.findNodeIndex(visibleNodes, current)
			newIndex := max(currentIndex-height, 0)
			if newIndex >= 0 && newIndex < len(visibleNodes) {
				t.Tree.SetCurrentNode(visibleNodes[newIndex])
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ': // Toggle item (space)
				if t.SelectEnabled {
					node := t.Tree.GetCurrentNode()
					name := node.GetReference().(string)
					t.ToggleSelectNode(name)
				}
				return nil
			case 'a': // Select all
				if t.SelectEnabled {
					t.SelectAll()
				}
				return nil
			case 'c': // Unselect all all
				if t.SelectEnabled {
					t.UnselectAll()
				}
				return nil
			case 'f': // Filter rows
				ShowFilter(filter, *t.FilterValue)
				return nil
			case 'F': // Remove filter
				CloseFilter(filter)
				*t.FilterValue = ""
				return nil
			case 'o': // Edit in editor
				item := tree.GetCurrentNode()
				name := item.GetReference().(string)
				t.EditNode(name)
				return nil
			case 'd': // Open description modal
				item := tree.GetCurrentNode()
				name := item.GetReference().(string)
				t.DescribeNode(name)
				return nil
			case 'g': // Top
				tree.SetCurrentNode(rootNode)
				misc.App.QueueEvent(tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone))
				return nil
			case 'G': // Bottom
				children := rootNode.GetChildren()
				last := children[len(children)-1]
				name := last.GetReference().(string)

				if name == "" {
					children = last.GetChildren()
					last = children[len(children)-1]
				}

				tree.SetCurrentNode(last)
				misc.App.QueueEvent(tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone))
				return nil
			}
		}
		return event
	})

	// Events
	var previousNode *tview.TreeNode
	var previousColor tcell.Color
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		if previousNode != nil {
			previousNode.SetColor(previousColor)
		}
		if node != nil {
			previousColor = node.GetColor()
			previousNode = node
			node.SetColor(misc.STYLE_ITEM_FOCUSED.Bg)
		}
	})
	t.Tree.SetFocusFunc(func() {
		InitFilter(t.Filter, *t.FilterValue)

		misc.PreviousPane = t.Tree
		misc.PreviousModel = t
		misc.SetActive(t.Root.Box, t.Title, true)
	})
	t.Tree.SetBlurFunc(func() {
		misc.PreviousPane = t.Tree
		misc.PreviousModel = t
		misc.SetActive(t.Root.Box, t.Title, false)
	})
}

func (t *TTree) UpdateProjects(paths []dao.TNode) {
	t.RootNode.ClearChildren()

	var itree []dao.TreeNode
	for i := range paths {
		itree = dao.AddToTree(itree, paths[i])
	}

	t.List = []*TNode{}
	for i := range itree {
		t.BuildProjectTree(t.RootNode, itree[i])
	}
}

func (t *TTree) UpdateProjectsStyle() {
	for _, node := range t.List {
		t.setNodeSelect(node)
	}
}

func (t *TTree) BuildProjectTree(node *tview.TreeNode, tnode dao.TreeNode) {
	// Project
	if len(tnode.Children) == 0 {
		pathName := misc.Colorize(tnode.Path, *misc.TUITheme.Item)
		childTreeNode := tview.NewTreeNode(pathName).
			SetReference(tnode.ProjectName).
			SetSelectable(true)

		node.AddChild(childTreeNode)
		childListNode := &TNode{
			ID:          tnode.ProjectName,
			DisplayName: tnode.Path,
			Type:        "project",
			TreeNode:    childTreeNode,
			Children:    &[]TNode{},
		}
		t.List = append(t.List, childListNode)
		return
	}

	// Directory
	pathName := misc.Colorize(tnode.Path, *misc.TUITheme.ItemDir)
	parentTreeNode := tview.NewTreeNode(pathName).
		SetReference("").
		SetSelectable(false)
	node.AddChild(parentTreeNode)

	parentListNode := &TNode{
		ID:          tnode.ProjectName,
		DisplayName: tnode.Path,
		TreeNode:    parentTreeNode,
		Type:        "directory",
	}
	t.List = append(t.List, parentListNode)
	for i := range tnode.Children {
		t.BuildProjectTree(parentTreeNode, tnode.Children[i])
	}
}

func (t *TTree) UpdateTasks(nodes []TNode) {
	t.RootNode.ClearChildren()
	t.List = []*TNode{}

	for _, parentNode := range nodes {
		// Parent
		displayName := misc.Colorize(parentNode.DisplayName, *misc.TUITheme.Item)
		parentTreeNode := tview.NewTreeNode(displayName).
			SetReference(parentNode.ID).
			SetSelectable(true)
		t.RootNode.AddChild(parentTreeNode)

		parentListNode := &TNode{
			DisplayName: parentNode.DisplayName,
			ID:          parentNode.DisplayName,
			Type:        parentNode.Type,
			TreeNode:    parentTreeNode,
			Children:    &[]TNode{},
		}
		t.List = append(t.List, parentListNode)

		// Children
		for _, childNode := range *parentNode.Children {
			displayName := misc.Colorize(parentNode.DisplayName, *misc.TUITheme.Item)
			childTreeNode := tview.
				NewTreeNode(displayName).
				SetSelectable(false)
			parentTreeNode.AddChild(childTreeNode)

			listChildNode := &TNode{
				DisplayName: childNode.DisplayName,
				Type:        childNode.Type,
				TreeNode:    childTreeNode,
				Children:    &[]TNode{},
			}
			*parentListNode.Children = append(*parentListNode.Children, *listChildNode)
		}
	}
}

func (t *TTree) UpdateTasksStyle() {
	for _, node := range t.List {
		if t.IsNodeSelected(node.DisplayName) {
			displayName := misc.Colorize(node.DisplayName, *misc.TUITheme.ItemSelected)
			node.TreeNode.SetText(displayName)
			for _, child := range *node.Children {
				displayName := misc.Colorize(child.DisplayName, *misc.TUITheme.ItemSelected)
				child.TreeNode.SetText(displayName)
			}
		} else {
			displayName := misc.Colorize(node.DisplayName, *misc.TUITheme.Item)
			node.TreeNode.SetText(displayName)
			for _, child := range *node.Children {
				if child.Type == "task-ref" {
					displayName := misc.Colorize(child.DisplayName, *misc.TUITheme.ItemRef)
					child.TreeNode.SetText(displayName)
				} else {
					displayName := misc.Colorize(child.DisplayName, *misc.TUITheme.Item)
					child.TreeNode.SetText(displayName)
				}
			}
		}
	}
}

func (t *TTree) ToggleSelectCurrentNode(id string) {
	for i := 0; i < len(t.List); i++ {
		node := t.List[i]
		if node.ID == id {
			t.setNodeSelect(node)
			return
		}
	}
}

func (t *TTree) setNodeSelect(node *TNode) {
	if t.IsNodeSelected(node.ID) {
		displayName := misc.Colorize(node.DisplayName, *misc.TUITheme.ItemSelected)
		node.TreeNode.SetText(displayName)
		for _, childNode := range *node.Children {
			displayName := misc.Colorize(childNode.DisplayName, *misc.TUITheme.ItemSelected)
			childNode.TreeNode.SetText(displayName)
		}
		return
	}

	switch node.Type {
	case "directory":
		displayName := misc.Colorize(node.DisplayName, *misc.TUITheme.ItemDir)
		node.TreeNode.SetText(displayName)
	case "task":
		displayName := misc.Colorize(node.DisplayName, *misc.TUITheme.Item)
		node.TreeNode.SetText(displayName)
		for _, childNode := range *node.Children {
			if childNode.Type == "task-ref" {
				displayName := misc.Colorize(childNode.DisplayName, *misc.TUITheme.ItemRef)
				childNode.TreeNode.SetText(displayName)
			} else {
				displayName := misc.Colorize(childNode.DisplayName, *misc.TUITheme.Item)
				childNode.TreeNode.SetText(displayName)
			}
		}
	case "project":
		displayName := misc.Colorize(node.DisplayName, *misc.TUITheme.Item)
		node.TreeNode.SetText(displayName)
	default:
		displayName := misc.Colorize(node.DisplayName, *misc.TUITheme.Item)
		node.TreeNode.SetText(displayName)
	}
}

func (t *TTree) FocusFirst() {
	t.Tree.SetCurrentNode(t.RootNode)
}

func (t *TTree) FocusLast() {
	children := t.RootNode.GetChildren()
	last := children[len(children)-1]
	name := last.GetReference().(string)

	if name == "" {
		children = last.GetChildren()
		last = children[len(children)-1]
	}

	t.Tree.SetCurrentNode(last)
}

func (t *TTree) ClearFilter() {
	CloseFilter(t.Filter)
	*t.FilterValue = ""
}

func (t *TTree) applyFilter() {
	*t.FilterValue = t.Filter.GetText()
}

func (t *TTree) getVisibleNodes() []*tview.TreeNode {
	var nodes []*tview.TreeNode
	var walk func(*tview.TreeNode)
	walk = func(node *tview.TreeNode) {
		if node == nil {
			return
		}
		ref := node.GetReference()
		if ref != nil && ref.(string) != "" {
			nodes = append(nodes, node)
		}
		if node.IsExpanded() {
			for _, child := range node.GetChildren() {
				walk(child)
			}
		}
	}
	walk(t.RootNode)
	return nodes
}

func (t *TTree) findNodeIndex(nodes []*tview.TreeNode, target *tview.TreeNode) int {
	for i, node := range nodes {
		if node == target {
			return i
		}
	}
	return 0
}
