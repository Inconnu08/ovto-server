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

var __1_initial_schema_down_sql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\xf0\x74\x53\x70\x8d\xf0\x0c\x0e\x09\x56\x28\x2d\x4e\x2d\x2a\x56\x70\x76\x0c\x76\x76\x74\x71\xb5\xe6\xc2\xaa\x26\xb9\x28\x35\x25\x35\xaf\x24\x33\x31\x87\x90\xca\xb4\xfc\xfc\x94\x82\xa2\xfc\xb2\xcc\x94\xd4\x22\x02\x4a\x8b\x52\x8b\x4b\x4a\x13\x8b\x12\xf3\x4a\x10\x0a\xb9\x00\x01\x00\x00\xff\xff\x9d\xe8\x2f\x12\xa3\x00\x00\x00")

func _1_initial_schema_down_sql() ([]byte, error) {
	return bindata_read(
		__1_initial_schema_down_sql,
		"1_initial_schema.down.sql",
	)
}

var __1_initial_schema_up_sql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x94\x41\x8f\x9b\x30\x10\x85\xef\xfc\x0a\x1f\x41\xea\xa1\xaa\x94\x53\x4f\x6e\xea\xa8\xa8\x84\xa6\xc6\x54\x4d\x2f\x68\x82\x9d\xae\x25\xe2\x89\x6c\x93\x28\xff\x7e\xb5\x01\x29\x84\x84\x25\x9b\xd5\x6a\x97\xeb\xbc\x87\x3d\xdf\x7b\x30\xe5\x8c\x0a\x46\x04\xfd\x96\x30\x12\xcf\x48\xfa\x4b\x10\xf6\x37\xce\x44\x46\x6a\xa7\xac\x0b\xc2\x80\x10\x42\xb4\x24\x67\x4f\xc6\x78\x4c\x13\x72\x94\xa7\x79\x92\x90\x05\x8f\xe7\x94\x2f\xc9\x4f\xb6\xfc\x74\x74\xa8\x0d\xe8\xaa\xe3\xf8\x43\xf9\xf4\x07\xe5\x27\x47\x9e\xc6\xbf\x73\xd6\x88\xd7\x75\x55\x19\xd8\xa8\x9e\x38\x9c\x7c\x8e\xae\x1b\x60\x07\x1e\xec\xc5\xdb\xdb\xa1\x94\x56\x39\xd7\x1f\x86\x5f\x26\x93\xa8\x51\x6c\x1f\xd0\xa8\xcb\xcb\x35\xc3\xd2\x2a\xf0\x4a\x16\xe0\x89\x88\xe7\x2c\x13\x74\xbe\x10\xff\x4e\xf7\xf8\xce\x66\x34\x4f\x04\x31\xb8\x0f\xa3\x20\xfa\x1a\x04\xcf\x40\x2c\xad\x92\xca\x78\x0d\xd5\xfd\x28\x9f\x82\x28\x3a\xb6\x38\x15\x27\x35\x67\x33\xc6\x59\x3a\x65\x6d\x60\xed\x82\xe0\xdc\x1e\xad\x1c\xa2\xff\x06\x9b\xae\x11\xe5\xd6\xe2\x4e\x4b\x65\x3f\x68\x6b\x06\x62\xef\x51\x79\x07\x76\x56\x39\x5f\x83\x05\xe3\xef\x26\xd7\x05\x71\x13\x0c\xdc\x9b\xb3\x5a\x0d\xb5\xaa\x9b\x6b\xfb\x7d\xad\xb0\xf6\x2f\x39\xaa\xc2\x12\xbc\x46\x33\x42\x54\xfb\xc3\x95\x0d\x7a\x2a\xb0\x0a\xc6\x55\x25\xd6\xc6\xdb\xc3\x58\xd2\x37\xf5\x01\xb7\xca\x68\xf3\xbf\xf0\xba\x21\x3c\x70\x62\x85\x6e\x5c\x05\x9b\x15\x38\x07\x12\x6d\x51\xa2\x54\xaf\xfc\xf1\x3c\x06\x00\x00\xff\xff\x40\x6a\x1c\xb5\xbd\x05\x00\x00")

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
