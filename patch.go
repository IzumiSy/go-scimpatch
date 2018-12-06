package scimpatch

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	Add     = "add"
	Remove  = "remove"
	Replace = "replace"
)

type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type Modification struct {
	Schemas []string `json:"schemas"`
	Ops     []Patch  `json:"Operations"`
}

func (m Modification) Validate() error {
	if len(m.Schemas) != 1 && m.Schemas[0] != PatchOpUrn {
		return fmt.Errorf("Invalid parameter: %+v", m.Schemas)
	}

	if len(m.Ops) == 0 {
		return fmt.Errorf("Invalid parameter: no ops")
	}

	for _, patch := range m.Ops {
		switch strings.ToLower(patch.Op) {
		case Add:
			if patch.Value == nil {
				return fmt.Errorf("Invalid parameter: value is not present")
			} else if len(patch.Path) == 0 {
				if _, ok := patch.Value.(map[string]interface{}); !ok {
					return fmt.Errorf("Invalid parameter: path is not present")
				}
			}
		case Replace:
			if patch.Value == nil {
				return fmt.Errorf("Invalid parameter: value is not present")
			} else if len(patch.Path) == 0 {
				return fmt.Errorf("Invalid parameter: path is not present")
			}
		case Remove:
			if len(patch.Path) == 0 {
				return fmt.Errorf("Invalid parameter: path is not present")
			}

		default:
			return fmt.Errorf("Invalid operation: must be one of [add|remove|replace]")
		}
	}

	return nil
}

func ApplyPatch(patch Patch, subj *Resource, schema *Schema) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				err = r.(error)
			default:
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	err, psPtr, pathPtr := buildPatchState(patch, schema)
	if err != nil {
		return err
	}

	ps := *psPtr
	path := *pathPtr

	err = applyAzureADRemoveSupport(&ps, &path)
	if err != nil {
		return err
	}

	v := reflect.ValueOf(patch.Value)
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch strings.ToLower(patch.Op) {
	case Add:
		ps.applyPatchAdd(path, fixValueWithType(ps.destAttr, v), subj)
	case Replace:
		ps.applyPatchReplace(path, fixValueWithType(ps.destAttr, v), subj)
	case Remove:
		ps.applyPatchRemove(path, subj)
	default:
		err = fmt.Errorf("Invalid operator: %s", patch.Op)
	}
	return
}

// [AzureAD対策]
// TODO: これが必要な理由をあとでかく
func applyAzureADRemoveSupport(ps *patchState, path *Path) error {
	patch := (*ps).patch

	if ps.destAttr != nil {
		if strings.ToLower(patch.Op) == "remove" && ps.destAttr.MultiValued && patch.Value != nil {
			v := reflect.ValueOf(patch.Value)
			if v.Kind() == reflect.Interface {
				v = v.Elem()
			}
			valueKind := v.Kind()

			var value reflect.Value
			switch valueKind {
			case reflect.Map:
				mapValue := v.Interface().(map[string]interface{})
				value = reflect.ValueOf(mapValue["value"])
			case reflect.Slice, reflect.Array:
				arrayValue := v.Interface().([]interface{})
				head := arrayValue[0].(map[string]interface{})
				value = reflect.ValueOf(head["value"])
			}

			patch.Path = patch.Path + "[value eq \"" + value.String() + "\"]"
			err, psPtr, pathPtr := buildPatchState(patch, ps.sch)
			if err != nil {
				return err
			}

			*ps = *psPtr
			*path = *pathPtr
		}
	}

	return nil
}

func buildPatchState(patch Patch, schema *Schema) (error, *patchState, *Path) {
	ps := patchState{patch: patch, sch: schema}

	var err error
	var path Path
	if len(patch.Path) == 0 {
		path = nil
	} else {
		path, err = NewPath(patch.Path)
		if err != nil {
			return err, nil, nil
		}
		fmt.Printf("%+v\n", path)
		path.CorrectCase(schema, true)

		if attr := schema.GetAttribute(path, true); attr != nil {
			ps.destAttr = attr
		} else {
			return fmt.Errorf("No attribute found for path: %s", patch.Path), nil, nil
		}
	}

	return nil, &ps, &path
}

// [AzureAD対策]
// 特定のSCIMクライアントがstring型のフィールドに対して配列値を送ってくることがある
// なので、配列やオブジェクトが値として送られてきた場合にはここで取り出して単一値として与えたい。
// 雑な実装として配列値が与えられた場合には必ず配列の先頭のオブジェクトから常にvalueフィールドを対象のデータとして取り出すことにする。
func fixValueWithType(destAttr *Attribute, value reflect.Value) reflect.Value {
	valueKind := value.Kind()

	// ここでps.destAttrのチェックをしているのはimplicit pathというpathを直接指定せずにデータの追加をするためのテストケースがあるため
	// 実際にはSCIMのRFCを読んでも規約としては存在していなさそうなので、path指定は必須項目としたいが一旦ここではnilチェックに留めておく。
	if destAttr != nil {
		isValueAndTypeUnmatched :=
			(destAttr.Type == "string" && valueKind != reflect.String) ||
				(destAttr.Type == "boolean" && valueKind != reflect.Bool)

		if isValueAndTypeUnmatched {
			var v reflect.Value

			// Mapであればそこからvalueフィールドを取り出す
			// Slice/Arrayであればその先頭要素のオブジェクトからvalueフィールドを取り出す
			switch valueKind {
			case reflect.Map:
				mapValue := value.Interface().(map[string]interface{})
				v = reflect.ValueOf(mapValue["value"])
			case reflect.Slice, reflect.Array:
				arrayValue := value.Interface().([]interface{})
				head := arrayValue[0].(map[string]interface{})
				v = reflect.ValueOf(head["value"])
			}

			// これもAzureADだがbooleanの値をPascalCaseで送ってくるためパースできない。
			// なのでもしその文字列がTrue/Falseであればそれをbool値に変換する
			switch v.Interface().(string) {
			case "True":
				v = reflect.ValueOf(true)
			case "False":
				v = reflect.ValueOf(false)
			}

			return v
		}
	}

	return value
}

type patchState struct {
	patch    Patch
	destAttr *Attribute
	sch      *Schema
}

func (ps *patchState) throw(err error) {
	if err != nil {
		panic(err)
	}
}

func (ps *patchState) applyPatchRemove(p Path, subj *Resource) {
	basePath, lastPath := p.SeparateAtLast()
	baseChannel := make(chan interface{}, 1)
	if basePath == nil {
		go func() {
			baseChannel <- subj.Complex
			close(baseChannel)
		}()
	} else {
		baseChannel = subj.Get(basePath, ps.sch)
	}

	var baseAttr AttributeSource = ps.sch
	if basePath != nil {
		baseAttr = ps.sch.GetAttribute(basePath, true)
	}

	for base := range baseChannel {
		baseVal := reflect.ValueOf(base)
		if baseVal.IsNil() {
			continue
		}
		if baseVal.Kind() == reflect.Interface {
			baseVal = baseVal.Elem()
		}

		switch baseVal.Kind() {
		case reflect.Map:
			keyVal := reflect.ValueOf(lastPath.Base())
			if ps.destAttr.MultiValued {
				if lastPath.FilterRoot() == nil {
					baseVal.SetMapIndex(keyVal, reflect.Value{})
				} else {
					origVal := baseVal.MapIndex(keyVal)
					baseAttr = baseAttr.GetAttribute(lastPath, false)
					reverseRoot := &filterNode{
						data: Not,
						typ:  LogicalOperator,
						left: lastPath.FilterRoot().(*filterNode),
					}
					newElemChannel := MultiValued(origVal.Interface().([]interface{})).Filter(reverseRoot, baseAttr)
					newArr := make([]interface{}, 0)
					for newElem := range newElemChannel {
						newArr = append(newArr, newElem)
					}
					if len(newArr) == 0 {
						baseVal.SetMapIndex(keyVal, reflect.Value{})
					} else {
						baseVal.SetMapIndex(keyVal, reflect.ValueOf(newArr))
					}
				}
			} else {
				baseVal.SetMapIndex(keyVal, reflect.Value{})
			}
		case reflect.Array, reflect.Slice:
			keyVal := reflect.ValueOf(lastPath.Base())
			for i := 0; i < baseVal.Len(); i++ {
				elemVal := baseVal.Index(i)
				if elemVal.Kind() == reflect.Interface {
					elemVal = elemVal.Elem()
				}
				switch elemVal.Kind() {
				case reflect.Map:
					elemVal.SetMapIndex(keyVal, reflect.Value{})
				default:
					ps.throw(fmt.Errorf("Array base contains non-map: %s", ps.patch.Path))
				}
			}
		default:
			ps.throw(fmt.Errorf("Base evaluated to non-map and non-array: %s", ps.patch.Path))
		}
	}
}

func (ps *patchState) applyPatchReplace(p Path, v reflect.Value, subj *Resource) {
	basePath, lastPath := p.SeparateAtLast()
	baseChannel := make(chan interface{}, 1)
	if basePath == nil {
		go func() {
			baseChannel <- subj.Complex
			close(baseChannel)
		}()
	} else {
		baseChannel = subj.Get(basePath, ps.sch)
	}

	for base := range baseChannel {
		baseVal := reflect.ValueOf(base)
		if baseVal.IsNil() {
			continue
		}
		if baseVal.Kind() == reflect.Interface {
			baseVal = baseVal.Elem()
		}
		baseVal.SetMapIndex(reflect.ValueOf(lastPath.Base()), v)
	}
}

func (ps *patchState) applyPatchAdd(p Path, v reflect.Value, subj *Resource) {
	if p == nil {
		if v.Kind() != reflect.Map {
			ps.throw(fmt.Errorf("Invalid parameter for add operation"))
		}
		for _, k := range v.MapKeys() {
			v0 := v.MapIndex(k)
			if err := ApplyPatch(Patch{Op: Add, Path: k.String(), Value: v0.Interface()}, subj, ps.sch); err != nil {
				ps.throw(err)
			}
		}
	} else {
		basePath, lastPath := p.SeparateAtLast()
		baseChannel := make(chan interface{}, 1)

		if basePath == nil {
			go func() {
				baseChannel <- subj.Complex
				close(baseChannel)
			}()
		} else {
			baseChannel = subj.Get(basePath, ps.sch)
		}

		for base := range baseChannel {
			baseVal := reflect.ValueOf(base)
			if baseVal.IsNil() {
				continue
			}
			if baseVal.Kind() == reflect.Interface {
				baseVal = baseVal.Elem()
			}

			switch baseVal.Kind() {
			case reflect.Map:
				keyVal := reflect.ValueOf(lastPath.Base())
				if ps.destAttr.MultiValued {
					origVal := baseVal.MapIndex(keyVal)
					if !origVal.IsValid() {
						switch v.Kind() {
						case reflect.Array, reflect.Slice:
							baseVal.SetMapIndex(keyVal, v)
						default:
							baseVal.SetMapIndex(keyVal, reflect.ValueOf([]interface{}{v.Interface()}))
						}
					} else {
						if origVal.Kind() == reflect.Interface {
							origVal = origVal.Elem()
						}
						var newArr MultiValued
						switch v.Kind() {
						case reflect.Array, reflect.Slice:
							for i := 0; i < v.Len(); i++ {
								newArr = MultiValued(origVal.Interface().([]interface{})).Add(v.Index(i).Interface())
							}
						default:
							newArr = MultiValued(origVal.Interface().([]interface{})).Add(v.Interface())
						}
						baseVal.SetMapIndex(keyVal, reflect.ValueOf(newArr))
					}
				} else {
					baseVal.SetMapIndex(keyVal, v)
				}
			case reflect.Array, reflect.Slice:
				for i := 0; i < baseVal.Len(); i++ {
					elemVal := baseVal.Index(i)
					if elemVal.Kind() == reflect.Interface {
						elemVal = elemVal.Elem()
					}
					switch elemVal.Kind() {
					case reflect.Map:
						elemVal.SetMapIndex(reflect.ValueOf(lastPath.Base()), v)
					default:
						ps.throw(fmt.Errorf("Array base contains non-map: %s", ps.patch.Path))
					}
				}
			default:
				ps.throw(fmt.Errorf("Base evaluated to non-map and non-array: %s", ps.patch.Path))
			}
		}
	}
}
