package main

import (
	"fmt"
	"io"
	"os"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/nfi-hashicorp/yaml2tf/terraformfmt"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

// Modeled after hclwrite.TokensForValue, but for YAML.
//
// Generally we don't worry about whitespace, and assume the caller will format it.
// (🖕 YAML with your semantic whitespace!)
//
// I couldn't make this work with hclwrite body and block building
// because those don't give us enough control over ordering and comments.
func yamlIntoTFTokens(y *yaml.Node) []*hclwrite.Token {
	switch y.Kind {
	case yaml.DocumentNode:
		return yamlIntoTFTokens(y.Content[0])
	case yaml.MappingNode:
		toks := []*hclwrite.Token{}
		if y.HeadComment != "" {
			toks = append(toks, &hclwrite.Token{
				Type:  hclsyntax.TokenComment,
				Bytes: []byte(fmt.Sprintf("%s\n", y.HeadComment)),
			})
		}
		toks = append(toks, &hclwrite.Token{
			Type:  hclsyntax.TokenOBrace,
			Bytes: []byte{'{'},
		})
		if len(y.Content) > 0 {
			toks = append(toks, &hclwrite.Token{
				Type:  hclsyntax.TokenNewline,
				Bytes: []byte{'\n'},
			})
		}
		for i := 0; i < len(y.Content); i += 2 {
			k := y.Content[i]
			if k.Kind != yaml.ScalarNode && k.Tag != "!!str" {
				panic(fmt.Sprintf("[%d,%d] key is not a string: %v", y.Line, k.Column, k))
			}
			if k.HeadComment != "" {
				toks = append(toks, &hclwrite.Token{
					Type:  hclsyntax.TokenComment,
					Bytes: []byte(fmt.Sprintf("%s\n", k.HeadComment)),
				})
			}
			v := y.Content[i+1]

			toks = append(toks, yamlIntoTFTokens(k)...)
			toks = append(toks, &hclwrite.Token{
				Type:  hclsyntax.TokenEqual,
				Bytes: []byte{'='},
			},
			)
			toks = append(toks, yamlIntoTFTokens(v)...)
			toks = append(toks,
				&hclwrite.Token{
					Type:  hclsyntax.TokenComma,
					Bytes: []byte{','},
				},
				&hclwrite.Token{
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte{'\n'},
				},
			)
		}
		toks = append(toks, &hclwrite.Token{
			Type:  hclsyntax.TokenCBrace,
			Bytes: []byte{'}'},
		})
		return toks
	case yaml.SequenceNode:
		toks := []*hclwrite.Token{
			{
				Type:  hclsyntax.TokenOBrack,
				Bytes: []byte{'['},
			},
		}
		toks = append(toks, &hclwrite.Token{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		})
		for _, v := range y.Content {
			// TODO: comments
			toks = append(toks, yamlIntoTFTokens(v)...)
			toks = append(toks, &hclwrite.Token{
				Type:  hclsyntax.TokenComma,
				Bytes: []byte{','},
			})
			// TODO: newlines based on source
			toks = append(toks, &hclwrite.Token{
				Type:  hclsyntax.TokenNewline,
				Bytes: []byte{'\n'},
			})
		}
		toks = append(toks, &hclwrite.Token{
			Type:  hclsyntax.TokenCBrack,
			Bytes: []byte{']'},
		})
		return toks
	case yaml.ScalarNode:
		var ctyVal cty.Value
		switch y.Tag {
		case "!!str":
			// TODO: translate quote style, escape, etc?
			ctyVal = cty.StringVal(y.Value)
		case "!!bool":
			var b bool
			yaml.Unmarshal([]byte(y.Value), &b)
			ctyVal = cty.BoolVal(b)
		default:
			panic(fmt.Sprintf("[%d,%d] unhandled tag for scalar %v", y.Line, y.Column, y.Tag))
		}
		return hclwrite.TokensForValue(ctyVal)
	default:
		panic(fmt.Sprintf("[%d,%d] unhandled node kind %v", y.Line, y.Column, y.Kind))
	}
}

// yoinked from hclwrite
func escapeQuotedStringLit(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	buf := make([]byte, 0, len(s))
	for i, r := range s {
		switch r {
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		case '"':
			buf = append(buf, '\\', '"')
		case '\\':
			buf = append(buf, '\\', '\\')
		case '$', '%':
			buf = appendRune(buf, r)
			remain := s[i+1:]
			if len(remain) > 0 && remain[0] == '{' {
				// Double up our template introducer symbol to escape it.
				buf = appendRune(buf, r)
			}
		default:
			if !unicode.IsPrint(r) {
				var fmted string
				if r < 65536 {
					fmted = fmt.Sprintf("\\u%04x", r)
				} else {
					fmted = fmt.Sprintf("\\U%08x", r)
				}
				buf = append(buf, fmted...)
			} else {
				buf = appendRune(buf, r)
			}
		}
	}
	return buf
}

// yoinked from hclwrite
func appendRune(b []byte, r rune) []byte {
	l := utf8.RuneLen(r)
	for i := 0; i < l; i++ {
		b = append(b, 0) // make room at the end of our buffer
	}
	ch := b[len(b)-l:]
	utf8.EncodeRune(ch, r)
	return b
}

func yamlToTF(y *yaml.Node) *hclwrite.File {
	h := hclwrite.NewEmptyFile()
	h.Body().AppendUnstructuredTokens(yamlIntoTFTokens(y))
	terraformfmt.FormatBody(h.Body())
	return h
}

func main() {
	yb, _ := io.ReadAll(os.Stdin)

	y := yaml.Node{}
	yaml.Unmarshal(yb, &y)

	// TODO: also handle conversion of basic Terraform YAML templates with simple interpolation
	h := yamlToTF(&y)

	fmt.Println(string(h.Bytes()))
}
