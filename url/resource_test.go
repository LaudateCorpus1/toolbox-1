package url_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/url"
	"os"
	"path"
	"strings"
	"testing"
)

func TestNewResource(t *testing.T) {

	{
		var resource = url.NewResource("https://raw.githubusercontent.com/viant/toolbox/master/LICENSE.txt")
		assert.EqualValues(t, resource.ParsedURL.String(), "https://raw.githubusercontent.com/viant/toolbox/master/LICENSE.txt")
		data, err := resource.Download()
		assert.Nil(t, err)
		assert.NotNil(t, data)
	}

	{
		var resource = url.NewResource("https://raw.githubusercontent.com/viant//toolbox//master/LICENSE.txt")
		assert.Equal(t, "https://raw.githubusercontent.com/viant/toolbox/master/LICENSE.txt", resource.URL)
	}
	{
		var resource = url.NewResource("./../test")
		assert.True(t, strings.HasSuffix(resource.DirectoryPath(), "/toolbox/test"))

	}
	{
		var resource = url.NewResource("../test")
		assert.True(t, strings.HasSuffix(resource.DirectoryPath(), "/toolbox/test"))

	}
}

func TestNew_CredentialURL(t *testing.T) {

	{
		var resource = url.NewResource("https://raw.githubusercontent.com:80/viant/toolbox/master/LICENSE.txt?check=1&p=2")
		var URL = resource.CredentialURL("smith", "123")
		assert.EqualValues(t, "https://smith:123@raw.githubusercontent.com:80/viant/toolbox/master/LICENSE.txt?check=1&p=2", URL)
	}

	{
		var resource = url.NewResource("https://raw.githubusercontent.com:80/viant/toolbox/master/LICENSE.txt")
		var URL = resource.CredentialURL("smith", "")
		assert.EqualValues(t, "https://smith@raw.githubusercontent.com:80/viant/toolbox/master/LICENSE.txt", URL)
	}

}

func TestNew_DirectoryPath(t *testing.T) {
	{
		var resource = url.NewResource("https://raw.githubusercontent.com:80/viant/toolbox/master/LICENSE.txt")
		assert.EqualValues(t, "/viant/toolbox/master", resource.DirectoryPath())
	}
	{
		var resource = url.NewResource("https://raw.githubusercontent.com:80/viant/toolbox/master/avc")
		assert.EqualValues(t, "/viant/toolbox/master/avc", resource.DirectoryPath())
	}
	{
		var resource = url.NewResource("hter")
		assert.True(t, strings.HasSuffix(resource.DirectoryPath(), "hter"))
	}
}

func TestResource_YamlDecode(t *testing.T) {
	var filename = path.Join(os.Getenv("TMPDIR"), "resource.yaml")
	toolbox.RemoveFileIfExist(filename)
	defer toolbox.RemoveFileIfExist(filename)
	var aMap = map[string]interface{}{
		"a": 1,
		"b": "123",
		"c": []int{1, 3, 6},
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	fmt.Printf("%v\n", filename)
	if assert.Nil(t, err) {
		err = toolbox.NewYamlEncoderFactory().Create(file).Encode(aMap)
		assert.Nil(t, err)
	}

	var resource = url.NewResource(filename)
	assert.EqualValues(t, resource.ParsedURL.String(), toolbox.FileSchema+filename)

	var resourceData = make(map[string]interface{})
	err = resource.YAMLDecode(&resourceData)
	assert.Nil(t, err)
	assert.EqualValues(t, resourceData["a"], 1)
	assert.EqualValues(t, resourceData["b"], "123")

}

func TestResource_JsonDecode(t *testing.T) {
	var filename = path.Join(os.Getenv("TMPDIR"), "resource.json")
	toolbox.RemoveFileIfExist(filename)
	defer toolbox.RemoveFileIfExist(filename)
	var aMap = map[string]interface{}{
		"a": 1,
		"b": "123",
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	fmt.Printf("%v\n", filename)
	if assert.Nil(t, err) {
		err = toolbox.NewJSONEncoderFactory().Create(file).Encode(aMap)
		assert.Nil(t, err)
	}

	var resource = url.NewResource(filename)
	assert.EqualValues(t, resource.ParsedURL.String(), toolbox.FileSchema+filename)

	var resourceData = make(map[string]interface{})
	err = resource.Decode(&resourceData)
	assert.Nil(t, err)

	assert.EqualValues(t, resourceData["a"], 1)
	assert.EqualValues(t, resourceData["b"], "123")

}
