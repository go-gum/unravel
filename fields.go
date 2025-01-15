package unravel

import (
	"reflect"
	"slices"
	"strings"
)

type field struct {
	Name  string
	Type  reflect.Type
	Index []int
}

func fieldsToSerialize(ty reflect.Type, structTag string) []field {
	if ty.Kind() != reflect.Struct {
		panic("not a struct")
	}

	type Queued struct {
		Type        reflect.Type
		ParentIndex []int
	}

	type Candidate struct {
		Name     string
		Explicit bool
		Field    field
	}

	// initialize queue to walk
	queue := []Queued{{Type: ty}}

	candidates := map[string][]Candidate{}

	var order []string

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		for idx := range item.Type.NumField() {
			fi := item.Type.Field(idx)
			if !fi.IsExported() {
				continue
			}

			name, explicit := nameOf(fi, structTag)
			if name == "" {
				// this one is skipped
				continue
			}

			// derive index of this one. ensure we allocate a new slice by setting cap to
			// the length of the parents index
			parent := item.ParentIndex
			index := append(parent[:len(parent):len(parent)], fi.Index...)

			if fi.Anonymous && !explicit {
				// this is an embedded field. skip if not struct
				if fi.Type.Kind() != reflect.Struct {
					continue
				}

				// queue for later analysis
				queue = append(queue, Queued{fi.Type, index})
				continue
			}

			if len(candidates[name]) == 0 {
				order = append(order, name)
			}

			candidates[name] = append(candidates[name], Candidate{
				Name:     name,
				Explicit: explicit,
				Field: field{
					Name:  name,
					Index: index,
					Type:  fi.Type,
				},
			})
		}
	}

	var fields []field

	for _, name := range order {
		candidates := candidates[name]

		// INVARIANT Candidates are not empty here
		if len(candidates) == 0 {
			panic("candidates are empty")
		}

		// INVARIANT: verify that sorting holds:
		//  due to walking the type in bfs order, items in candidates are sorted by index length
		//  with the shortest index at the beginning.
		cmp := func(a, b Candidate) int { return len(a.Field.Index) - len(b.Field.Index) }
		if !slices.IsSortedFunc(candidates, cmp) {
			panic("candidates are not sorted")
		}

		var visible []Candidate

		// We take the prefix of candidates that have the same index length
		for idx := 0; idx < len(candidates); idx++ {
			if len(candidates[idx].Field.Index) == len(candidates[0].Field.Index) {
				visible = candidates[:idx+1]
			}
		}

		// if we have exactly one visible item, that one always wins
		if len(visible) == 1 {
			fields = append(fields, visible[0].Field)
			continue
		}

		// keep only explicit candidates
		explicit := slices.DeleteFunc(visible, func(c Candidate) bool { return !c.Explicit })

		// if we have exactly one explicit item, that one wins
		if len(explicit) == 1 {
			fields = append(fields, explicit[0].Field)
			continue
		}

		// No one single candidate found.
		// We ignore this fields and do not raise an error.
	}

	return fields
}

func nameOf(fi reflect.StructField, structTag string) (name string, explicit bool) {
	// parse json struct tag to get renamed alias
	tag := fi.Tag.Get(structTag)

	if tag == "" {
		// tag is empty, take the original name
		return fi.Name, false
	}

	if tag == "-" {
		// return empty name indicate: skip this field
		return "", true
	}

	idx := strings.IndexByte(tag, ',')
	switch {
	case idx == -1:
		// no comma, take the full tag as explicit name
		return tag, true

	case idx > 0:
		// non empty alias, take up to comma
		return tag[:idx], true

	default:
		// no alias before the comma, keep field name
		return fi.Name, false
	}
}
