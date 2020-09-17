package rule

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// 类型
const float32Typ = "float32"
const float64Typ = "float64"
const int32Typ = "int32"
const int64Typ = "int64"
const uint32Typ = "uint32"
const uint64Typ = "uint64"
const stringTyp = "string"
const boolTyp = "bool"
const enumTyp = "enum"
const byteTyp = "byte"
const messageTyp = "message"

// 规则
const eqTyp = "eq"
const ltTyp = "lt"
const gtTyp = "gt"
const gteTyp = "gte"
const lteTyp = "lte"
const inTyp = "in"
const notInTyp = "not_in"
const lenTyp = "len"
const minLenTyp = "min_len"
const maxLenTyp = "max_len"
const patternTyp = "pattern"
const prefixTyp = "prefix"
const suffixTyp = "suffix"
const containsTyp = "contains"
const notContainsTyp = "not_contains"
const minItemsTyp = "min_items"
const maxItemsTyp = "max_items"
const uniqueTyp = "unique"
const typeTyp = "type"
const rangeTyp = "range"

var tienum = map[string]string{
	eqTyp:          eqTpl,
	ltTyp:          ltTpl,
	gtTyp:          gtTpl,
	gteTyp:         gteTpl,
	lteTyp:         lteTpl,
	inTyp:          inTpl,
	notInTyp:       notInTpl,
	lenTyp:         lenTpl,
	minLenTyp:      minLenTpl,
	maxLenTyp:      maxLenTpl,
	patternTyp:     patternTpl,
	prefixTyp:      prefixTpl,
	suffixTyp:      suffixTpl,
	containsTyp:    containsTpl,
	notContainsTyp: notContainsTpl,
	minItemsTyp:    minItemsTpl,
	maxItemsTyp:    maxItemsTpl,
	uniqueTyp:      uniqueTpl,
	typeTyp:        typeTpl,
	rangeTyp:       rangeTpl,
}

// TemplateInfo 用以生成最终的 rule 模版
type TemplateInfo struct {
	Field protogen.Field // field 内容
	Key   string         // key 名 正常情况为 m.GetX() repeated 情况为 item
	Value string         // value
}

// Rule 获取规则 目前从注释中正则获取
type Rule struct {
	Key   string // 规则类型
	Value string // 规则内容
}

// RegisterFunctions 注册方法
func RegisterFunctions(tpl *template.Template) {
	tpl.Funcs(map[string]interface{}{
		"msgTyp":    msgTyp,
		"errname":   errName,
		"pkg":       pkgName,
		"slice":     slicefunc,
		"accessor":  accessor,
		"escape":    escape,
		"goType":    protoTypeToGoType,
		"rangeRule": rangeRulefunc,
		"validate":  validatefunc,
		"message":   messagefunc,
	})
}

// msgTyp 返回 msg 名
func msgTyp(message protogen.Message) string {
	return message.GoIdent.GoName
}

// errName 返回 err 名
func errName(message protogen.Message) string {
	return msgTyp(message) + "ValidationError"
}

// pkgName 返回包名
func pkgName(file protogen.File) string {
	return string(file.GoPackageName)
}

// slicefunc [1,2,3] 解析成数组
func slicefunc(s string) (r []string) {
	re := regexp.MustCompile(`^\[(.*)\]$`)
	matched := re.FindStringSubmatch(s)

	if len(matched) <= 1 {
		return
	}
	ss := strings.Split(matched[1], ",")
	for _, v := range ss {
		r = append(r, v)
	}
	return
}

// accessor 获取 m.GetField 字符串
func accessor(field protogen.Field) string {
	return fmt.Sprintf("m.Get%s()", field.GoName)
}

// escape 转义字符串中的"并返回
func escape(s string) string {
	return strings.Replace(s, "\"", "", -1)
}

// protoTypeToGoType 转化 proto 数据类型为 go 数据类型
func protoTypeToGoType(kind protoreflect.Kind) (typ string) {
	switch kind {
	case protoreflect.BoolKind:
		return boolTyp
	case protoreflect.EnumKind:
		return enumTyp
	case protoreflect.Int32Kind:
		return int32Typ
	case protoreflect.Sint32Kind:
		return int32Typ
	case protoreflect.Uint32Kind:
		return uint32Typ
	case protoreflect.Int64Kind:
		return int64Typ
	case protoreflect.Sint64Kind:
		return int64Typ
	case protoreflect.Uint64Kind:
		return uint64Typ
	case protoreflect.Sfixed32Kind:
		return int32Typ
	case protoreflect.Fixed32Kind:
		return uint32Typ
	case protoreflect.FloatKind:
		return float32Typ
	case protoreflect.Sfixed64Kind:
		return int64Typ
	case protoreflect.Fixed64Kind:
		return uint64Typ
	case protoreflect.DoubleKind:
		return float64Typ
	case protoreflect.StringKind:
		return stringTyp
	case protoreflect.BytesKind:
		return byteTyp
	case protoreflect.MessageKind:
		return messageTyp
	case protoreflect.GroupKind:
		return ""
	default:
		return ""
	}
}

// rangeRulefunc 返回对 range 规则的判断
func rangeRulefunc(key string, value string) string {

	matched := regexp.MustCompile(`(\(|\[)(.+),(.+)(\)|\])`).FindStringSubmatch(value)
	if len(matched) < 5 {
		panic(key + "range value 不规范")
	}

	faultRule := map[string]string{
		"(": " <= ",
		"[": " < ",
		")": " >= ",
		"]": " > ",
	}

	v1 := faultRule[matched[1]]
	v2 := matched[2]
	v3 := matched[3]
	v4 := faultRule[matched[4]]

	return key + v1 + v2 + "&&" + key + v4 + v3
}

// validatefunc 返回 field 校验规则
func validatefunc(field protogen.Field) (ss []string) {
	rs := getRules(field.Comments) // 获取所有规则

	for _, v := range rs {
		s := getTemplateInfo(field, v)
		ss = append(ss, s)
	}
	return
}

// messagefunc 处理 message 间的互相调用 repeated 需要增加循环
func messagefunc(field protogen.Field) (str string) {
	if field.Desc.Kind() != protoreflect.MessageKind {
		return
	}

	str = `
		if v, ok := interface{}(` + accessor(field) + `).(interface{ validate() error }); ok {
			if err := v.validate(); err != nil {
				return ` + field.Parent.GoIdent.GoName + `ValidationError {
					field:  "` + field.GoName + `",
					reason: "embedded message failed validation " + err.Error(),
				}
			}
		}
`

	if field.Desc.IsList() {
		str = `
	for _, item := range ` + accessor(field) + ` {
		if v, ok := interface{}(item).(interface{ validate() error }); ok {
			if err := v.validate(); err != nil {
				return ` + field.Parent.GoIdent.GoName + `ValidationError {
					field:  "` + field.GoName + `",
					reason: "embedded message failed validation " + err.Error(),
				}
			}
		}
	}
`
	}
	return
}

func getTemplateInfo(field protogen.Field, r Rule) (s string) {
	ti := TemplateInfo{
		Field: field,
		Key:   accessor(field),
		Value: r.Value,
	}
	if v, ok := tienum[r.Key]; ok {
		s = v
	}

	if field.Desc.IsList() && r.Key != minItemsTyp && r.Key != maxItemsTyp && r.Key != uniqueTyp {
		s = `
			for _, item := range ` + accessor(field) + ` {
		` + s + `
			}
		`
		ti.Key = "item"
	}

	if s == "" {
		s = defaultTpl
	}

	tpl := template.New("rule")
	RegisterFunctions(tpl)
	template.Must(tpl.Parse(s))

	buf := &bytes.Buffer{}

	if err := tpl.Execute(buf, ti); err != nil {
		panic(err)
	}

	return buf.String()
}

// getRules 返回了每行符合正则的 rules 数组
func getRules(cs protogen.CommentSet) (rs []Rule) {
	ops := make([]string, 0, len(tienum))
	for op, _ := range tienum {
		ops = append(ops, op)
	}

	r := "@(" + strings.Join(ops, "|") + "):\\s*(.+)\\s*"
	re := regexp.MustCompile(r)

	for _, line := range strings.Split(string(cs.Leading), "\n") {
		matched := re.FindStringSubmatch(line)

		if len(matched) < 3 {
			continue
		}

		r := Rule{
			Key:   matched[1],
			Value: matched[2],
		}

		rs = append(rs, r)
	}

	return
}

func inKinds(item protoreflect.Kind, items []protoreflect.Kind) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}
