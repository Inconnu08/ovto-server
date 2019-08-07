package migrations

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

var __1_initial_schema_down_sql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\xf0\x74\x53\x70\x8d\xf0\x0c\x0e\x09\x56\x28\x2d\x4e\x2d\x2a\x56\x70\x76\x0c\x76\x76\x74\x71\xb5\xe6\xc2\xaa\x26\xb9\x28\x35\x25\x35\xaf\x24\x33\x31\x87\x90\xca\xb4\xfc\xfc\x94\x82\xa2\xfc\xb2\xcc\x94\xd4\x22\x02\x4a\x8b\x52\x8b\x4b\x4a\x13\x8b\x12\xf3\x4a\x08\x28\x4c\xcc\x4d\x4a\x2c\x2e\x4e\x4c\xc9\x47\x32\x11\x10\x00\x00\xff\xff\xe3\xba\xc0\xd2\xcb\x00\x00\x00")

func _1_initial_schema_down_sql() ([]byte, error) {
	return bindata_read(
		__1_initial_schema_down_sql,
		"1_initial_schema.down.sql",
	)
}

var __1_initial_schema_up_sql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xd4\x55\xdd\x6e\xa3\x3c\x10\xbd\xcf\x53\x58\xbd\x22\x52\xbf\x4f\xe9\x4a\xbd\xaa\x76\x25\x9a\xba\x2a\x2a\x21\x5d\x02\xab\x76\x6f\xd0\x04\x3b\xd4\x2a\xd8\x68\x6c\x52\xe5\xed\x57\x0d\xa4\xfc\x04\x9a\x6c\xd5\xfd\xe3\x76\xce\x98\xf1\x99\x73\x8e\xa7\x3e\xb5\x03\x4a\xe8\x7d\x40\xbd\x85\x33\xf7\x48\x9e\xc4\xb8\xc9\x8d\xba\x18\xed\x95\x9c\x6b\xe2\xcd\x03\x42\xef\x9d\x45\xb0\x20\x27\x45\x21\xd8\x7f\x4a\xeb\xfc\xe4\x62\xb4\x03\x07\xf6\xa5\x4b\x3b\xc0\x42\x73\xd4\x23\x6b\x44\x08\x21\x82\x91\xd6\xb7\xa0\xbe\x63\xbb\x64\x0b\xf7\x42\xd7\x25\x77\xbe\x33\xb3\xfd\x07\x72\x4b\x1f\x4e\xb7\x1d\x3c\x03\x91\x36\x3a\xbe\xd9\xfe\xf4\xc6\xf6\xeb\x8e\xd0\x73\xbe\x86\xb4\x04\xaf\x8a\x34\x95\x90\xf1\x0e\xd8\x3a\x9f\x8c\xfb\x1b\x60\x0d\x06\x70\xef\xf4\xaa\xc8\x18\x72\xad\xbb\x45\xeb\xd3\xf9\xf9\xb8\x44\xe4\x8f\x4a\xf2\xfd\xe1\xca\x62\x8c\x1c\x0c\x67\x11\x98\xb2\x18\x38\x33\xba\x08\xec\xd9\x5d\xf0\xbd\x1e\xe6\x8a\x5e\xdb\xa1\x1b\x10\xa9\x9e\xad\xf1\x68\xfc\x36\x93\x31\x72\xc6\xa5\x11\x90\xbe\x9f\xcf\x97\x6d\x44\x8d\x36\xc7\x0b\x6a\xb4\x4f\xaf\xa9\x4f\xbd\x29\xad\xb6\x56\xdd\x12\xb4\x7e\x56\xc8\x86\x56\xf0\xab\xae\xbb\x52\x8a\xe5\xa8\xd6\x82\x71\xfc\x7b\xf4\xf3\xe6\xe6\xfb\x8f\xfd\x53\x04\x42\xb6\x04\xad\x81\xa9\x7f\x9b\xbe\xe5\x13\xe8\xc7\x21\x97\xa1\x8a\x9f\xb8\x19\x28\x1e\x47\xfc\x0a\x62\xbe\x54\xea\xe9\xc0\x7a\x84\xd9\x90\x37\x46\xae\x32\x03\x39\x1c\x81\xea\x4f\x96\x0e\x0a\xf9\x8a\x23\x42\x1a\xc5\x8a\xbd\x90\xb5\x06\x8c\x1f\x01\xfb\x49\xfa\x78\xf5\x20\xd7\x06\x0a\x04\x69\x06\xd4\x13\x86\xce\x55\xaf\x74\x5e\xff\x93\x70\x19\x21\x48\xa6\xb2\xe8\xe5\xb5\xb0\xaa\xd8\x34\xc2\xa4\x3d\xdb\x1f\x8e\x69\xf5\x2c\x5b\xa9\x35\x14\x5a\xcd\xc4\x38\x26\xe0\x97\xaa\x30\x43\xc2\x4a\x55\x0c\x46\x28\xf9\xfb\x54\x11\xab\x42\x1a\xdc\x1c\x40\x1d\xb2\x4e\xc5\x58\xce\xa5\x90\x49\x64\x44\x69\xc7\x81\x3f\xa6\x4a\x1f\x46\xd5\x31\x52\x2a\xb1\x45\xd3\x1a\x4c\x84\x3c\x89\xa4\xea\xe1\x10\xc1\x08\x99\xd4\xb3\x5e\xd1\xa9\x33\xb3\x5d\xeb\xec\xf4\x6c\xfc\x2a\x92\xc9\xff\x13\x32\xbd\xa1\xd3\x5b\x62\x55\xf8\x2f\x9f\xc9\xa4\x92\x0a\xc4\x46\xac\x1b\xd7\xbd\x9c\xcf\x5d\x6a\x7b\xfb\x92\x36\x58\xf0\xfa\x52\x3c\xd2\x06\x4c\xa1\x8f\x6d\xf9\x70\xf3\xe4\x1c\x33\xa1\xb5\x50\x72\xc0\x3c\x8e\x17\xec\x5c\xbe\xb3\x59\x29\xf0\xad\xab\x1a\x9a\xae\xeb\xbb\xc0\x6b\x79\xa7\xe5\x85\xee\x89\x43\x0a\xd9\xc2\x9a\x86\xb5\x04\x3b\x6d\x4f\xf2\x13\xf1\x10\x25\x90\xa6\x1c\x37\xef\x7e\x64\x7a\x29\xe8\xf3\x77\x97\x0b\x91\x41\x72\xd8\x0b\xef\xda\xee\x8f\x00\x00\x00\xff\xff\xec\x65\xd8\xd9\x19\x0b\x00\x00")

func _1_initial_schema_up_sql() ([]byte, error) {
	return bindata_read(
		__1_initial_schema_up_sql,
		"1_initial_schema.up.sql",
	)
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() ([]byte, error){
	"1_initial_schema.down.sql": _1_initial_schema_down_sql,
	"1_initial_schema.up.sql": _1_initial_schema_up_sql,
}
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() ([]byte, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"1_initial_schema.down.sql": &_bintree_t{_1_initial_schema_down_sql, map[string]*_bintree_t{
	}},
	"1_initial_schema.up.sql": &_bintree_t{_1_initial_schema_up_sql, map[string]*_bintree_t{
	}},
}}
