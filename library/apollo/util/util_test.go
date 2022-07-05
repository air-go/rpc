package util

import (
	"encoding/xml"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestExtractConf(t *testing.T) {
	convey.Convey("TestExtractConf", t, func() {
		convey.Convey("ext empty", func() {
			var c struct {
				A string `json:"a"`
				B string `json:"b"`
			}
			err := ExtractConf("conf", `{"a": "a"}`, &c)

			assert.NotNil(t, err)
		})
		convey.Convey("ext json", func() {
			var c struct {
				A string `json:"a"`
				B string `json:"b"`
			}
			err := ExtractConf("conf.json", `{"a": "a"}`, &c)

			assert.Nil(t, err)
			assert.Equal(t, "a", c.A)
		})
		convey.Convey("ext toml", func() {
			var c struct {
				A string `toml:"a"`
				B string `toml:"b"`
			}
			data := `A = "a"` + "\n" + `B = "b"` + "\n"
			err := ExtractConf("conf.txt", data, &c)

			assert.Nil(t, err)
			assert.Equal(t, "a", c.A)
			assert.Equal(t, "b", c.B)
		})
		convey.Convey("ext yaml", func() {
			var c struct {
				A string `yaml:"a"`
				B string `yaml:"b"`
			}
			data := `a: a` + "\n" + `b: b` + "\n"
			err := ExtractConf("conf.yaml", data, &c)

			assert.Nil(t, err)
			assert.Equal(t, "a", c.A)
			assert.Equal(t, "b", c.B)
		})
		convey.Convey("ext xml", func() {
			type Book struct {
				XMLName xml.Name `xml:"book"`
				Name    string   `xml:"name,attr"`
				Author  string   `xml:"author"`
				Time    string   `xml:"time"`
				Types   []string `xml:"types>type"`
				Test    string   `xml:",any"`
			}

			type Books struct {
				XMLName xml.Name `xml:"books"`
				Nums    int      `xml:"nums,attr"`
				Book    []Book   `xml:"book"`
			}

			c := &Books{}

			data := `<?xml version="1.0" encoding="utf-8"?>
            <books nums="2">
                <book name="思想">
                    <author>小张</author>
                    <time>2018年1月20日</time>
                    <types>
                        <type>教育</type>
                        <type>历史</type>
                    </types>
                    <test>我是多余的</test>
                </book>
                <book name="政治">
                    <author>小王</author>
                    <time>2018年1月20日</time>
                    <types>
                        <type>科学</type>
                        <type>人文</type>
                    </types>
                    <test>我是多余的</test>
                </book>
            </books>`
			err := ExtractConf("conf.xml", data, c)

			assert.Nil(t, err)
			assert.Equal(t, &Books{
				XMLName: xml.Name{
					Local: "books",
				},
				Nums: 2,
				Book: []Book{
					{
						XMLName: xml.Name{
							Local: "book",
						},
						Name:   "思想",
						Author: "小张",
						Time:   "2018年1月20日",
						Types: []string{
							"教育",
							"历史",
						},
						Test: "我是多余的",
					},
					{
						XMLName: xml.Name{Local: "book"},
						Name:    "政治",
						Author:  "小王",
						Time:    "2018年1月20日",
						Types: []string{
							"科学",
							"人文",
						},
						Test: "我是多余的",
					},
				},
			}, c)
		})
		convey.Convey("ext error", func() {
			var c struct {
				A string `json:"a"`
				B string `json:"b"`
			}
			err := ExtractConf("conf.abc", `{"a": "a"}`, &c)

			assert.NotNil(t, err)
		})
	})
}
