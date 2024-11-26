package xmldom_test

import (
	"github.com/rtenhove/go-xmldom"
	"strings"
	"testing"
)

func TestParseNamespaces(t *testing.T) {
	testCases := []struct {
		inputXML     string
		expectedAttr map[string]string
	}{
		{
			inputXML: `<root xmlns:xlink="http://www.w3.org/1999/xlink" xlink:href="https://github.com/subchen/go-xmldom"></root>`,
			expectedAttr: map[string]string{
				"xmlns:xlink": "http://www.w3.org/1999/xlink",
				"xlink:href":  "https://github.com/subchen/go-xmldom",
			},
		},
		{
			inputXML: `<root xml:lang="en" xsi:type="string" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"></root>`,
			expectedAttr: map[string]string{
				"xml:lang":  "en",
				"xsi:type":  "string",
				"xmlns:xsi": "http://www.w3.org/2001/XMLSchema-instance",
			},
		},
	}

	for _, testCase := range testCases {
		r := strings.NewReader(testCase.inputXML)
		doc, err := xmldom.Parse(r)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		root := doc.Root
		attributes := root.Attributes

		attrMap := make(map[string]string)
		for _, attr := range attributes {
			attrMap[attr.Name] = attr.Value
		}

		for expectedName, expectedValue := range testCase.expectedAttr {
			if value, exists := attrMap[expectedName]; !exists {
				t.Errorf("Attribute %s was expected but not found", expectedName)
			} else if value != expectedValue {
				t.Errorf("Attribute %s has unexpected value. Got: %s, Want: %s", expectedName, value, expectedValue)
			}
		}
	}
}

func TestSvgParse(t *testing.T) {
	root := xmldom.Must(xmldom.ParseFile("test.svg")).Root

	imagesNodes := root.FindByName("image")
	if len(imagesNodes) < 4 {
		t.Fatalf("No images")
	}
}

func TestSvgAttrNamespace(t *testing.T) {
	root := xmldom.Must(xmldom.ParseFile("test.svg")).Root
	uses := root.FindByName("use")

	var contains bool
	for _, a := range root.Attributes {
		if a.Name == "xmlns:xlink" {
			contains = true
		}
	}

	if !contains {
		t.Fatalf("Expect root to contain xmlns:xlink attribute")
	}

	for _, u := range uses {
		for _, a := range u.Attributes {
			if a.Name != "xlink:href" {
				t.Fatalf("Expect use tag to contain xlink:href attribute but got %s", a.Name)
			}
		}
	}
}

func TestParserDefaultsToTrimmingCharData(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<foo> bar </foo>`
	dp := xmldom.NewDOMParser()

	doc := xmldom.Must(dp.ParseXML(xml))

	if "bar" != doc.Root.Text {
		t.Fatalf("Expect xml to contain 'bar' but got '%s'", doc.Root.Text)
	}
}

func TestParserWithPreserveWhitespaceDoesNotTrimCharData(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<foo> bar </foo>`
	dp := xmldom.NewDOMParser().PreserveWhitespace(true)

	doc := xmldom.Must(dp.ParseXML(xml))

	if " bar " != doc.Root.Text {
		t.Fatalf("Expect xml to contain ' bar ' but got '%s'", doc.Root.Text)
	}
}
