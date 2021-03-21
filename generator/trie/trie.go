package trie

import (
	"fmt"
)

// TODO(SSH): cover with tests
func GetShortNames(names []string) (map[string]*string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	t, err := newArgsTrie(names)
	if err != nil {
		return nil, err
	}
	shortNames := make(map[string]*string, len(names))
	for _, name := range names {
		shortNames[name] = t.getShortName(name)
	}
	return shortNames, nil
}

type argsTrie struct {
	isValid bool
	next    map[rune]*argsTrie
}

func newArgsTrie(names []string) (*argsTrie, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("zero arguments passed")
	}
	root := &argsTrie{
		next: make(map[rune]*argsTrie),
	}
	for _, name := range names {
		if err := root.append(name); err != nil {
			return nil, err
		}
	}
	return root, nil
}

func (t *argsTrie) getShortName(name string) *string {
	nameRunes := []rune(name)
	return t.getShortNameAux(nameRunes, 0, 0)
}

func (t *argsTrie) getShortNameAux(name []rune, currIdx int, lastSplitPos int) *string {
	if currIdx >= len(name) {
		if t.isValid == false {
			panic(fmt.Sprintf("name %s not found in the trie", string(name)))
		}
		if len(t.next) != 0 || lastSplitPos >= len(name) {
			return nil
		}
		result := string(name[:lastSplitPos+1])
		return &result
	}
	if t.isValid {
		lastSplitPos = currIdx + 1
	} else if len(t.next) > 1 {
		lastSplitPos = currIdx
	}
	nextNode := t.next[name[currIdx]]
	if len(t.next) > 1 {
		lastSplitPos = currIdx
	}
	currIdx++
	return nextNode.getShortNameAux(name, currIdx, lastSplitPos)
}

func (t *argsTrie) append(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("passed name is empty")
	}
	return t.appendAux([]rune(name), name)
}

func (t *argsTrie) appendAux(acc []rune, originalName string) error {
	if len(acc) == 0 {
		if t.isValid {
			return fmt.Errorf("failed to append the argument %s: it is already in the trie", originalName)
		}
		t.isValid = true
		return nil
	}
	if _, ok := t.next[acc[0]]; !ok {
		t.next[acc[0]] = &argsTrie{
			next: make(map[rune]*argsTrie),
		}
	}
	return t.next[acc[0]].appendAux(acc[1:], originalName)
}
