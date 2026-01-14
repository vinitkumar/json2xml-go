package json2xml_test

import (
	"fmt"

	json2xml "github.com/vinitkumar/json2xml-go"
)

func Example() {
	data := map[string]any{
		"name":   "John",
		"age":    30,
		"active": true,
	}

	converter := json2xml.New(data)
	xml, err := converter.WithPretty(false).WithAttrType(false).ToXMLString()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(xml)
	// Output: <?xml version="1.0" encoding="UTF-8" ?><all><active>true</active><age>30</age><name>John</name></all>
}

func Example_withoutRoot() {
	data := map[string]any{"key": "value"}

	xml, err := json2xml.New(data).
		WithRoot(false).
		WithPretty(false).
		WithAttrType(false).
		ToXMLString()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(xml)
	// Output: <key>value</key>
}

func Example_xpathFormat() {
	data := map[string]any{"name": "Alice", "age": 25}

	xml, err := json2xml.New(data).
		WithXPathFormat(true).
		WithPretty(false).
		ToXMLString()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(xml)
	// Output: <?xml version="1.0" encoding="UTF-8" ?><map xmlns="http://www.w3.org/2005/xpath-functions"><number key="age">25</number><string key="name">Alice</string></map>
}

func Example_readFromString() {
	jsonStr := `{"login":"mojombo","id":1}`
	data, err := json2xml.ReadFromString(jsonStr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	xml, err := json2xml.New(data).
		WithPretty(false).
		WithAttrType(false).
		ToXMLString()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(xml)
	// Output: <?xml version="1.0" encoding="UTF-8" ?><all><id>1</id><login>mojombo</login></all>
}

func Example_listConversion() {
	data := map[string]any{
		"colors": []any{"red", "green", "blue"},
	}

	xml, err := json2xml.New(data).
		WithPretty(false).
		WithAttrType(false).
		ToXMLString()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(xml)
	// Output: <?xml version="1.0" encoding="UTF-8" ?><all><colors><item>red</item><item>green</item><item>blue</item></colors></all>
}

func Example_noItemWrap() {
	data := map[string]any{
		"bike": []any{"blue", "green"},
	}

	xml, err := json2xml.New(data).
		WithPretty(false).
		WithAttrType(false).
		WithItemWrap(false).
		WithRoot(false).
		ToXMLString()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(xml)
	// Output: <bike>blue</bike><bike>green</bike>
}
