// Code generated by go-bindata. DO NOT EDIT.
// sources:
// sql/20180519220506_add_events_table.down.sql
// sql/20180519220506_add_events_table.up.sql
package migrations

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var __20180519220506_add_events_tableDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\xf0\xf4\x73\x71\x8d\x50\xf0\x74\x53\x70\x8d\xf0\x0c\x0e\x09\x56\x48\x2d\x4b\xcd\x2b\x29\x8e\x2f\x28\x4d\xca\xc9\x4c\x8e\xcf\x4e\xad\x8c\xcf\x4c\xa9\x50\x70\x76\x0c\x76\x76\x74\x71\xb5\xe6\xc2\xa7\xa7\x28\x35\x39\xbf\x28\x25\x35\x25\x3e\xb1\x24\xbe\xb4\x38\xb5\x28\xbe\x34\x33\x85\x78\xdd\xd8\x75\x40\xb4\x84\x38\x3a\xf9\xb8\x62\x68\x81\xab\x02\x04\x00\x00\xff\xff\xbd\xc1\x7e\x32\xc9\x00\x00\x00")

func _20180519220506_add_events_tableDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__20180519220506_add_events_tableDownSql,
		"20180519220506_add_events_table.down.sql",
	)
}

func _20180519220506_add_events_tableDownSql() (*asset, error) {
	bytes, err := _20180519220506_add_events_tableDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "20180519220506_add_events_table.down.sql", size: 201, mode: os.FileMode(420), modTime: time.Unix(1526915439, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __20180519220506_add_events_tableUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x90\xcd\x4a\xc4\x30\x14\x46\xf7\x79\x8a\x6f\x39\x05\xdf\x60\x56\x19\xe7\x0e\x06\xd3\x74\x48\xee\x30\xad\x9b\x50\x27\x59\x04\x45\xa5\x3f\xa2\x6f\x2f\x2d\xd6\x16\xea\xa2\xcb\x70\xf8\x0e\x27\xf7\xde\x92\x64\x02\xcb\x83\x26\xa8\x13\x4c\xc1\xa0\x52\x39\x76\x88\x9f\xf1\xad\x6b\xb1\x13\x40\x0a\x70\x64\x95\xd4\x38\x5b\x95\x4b\x5b\xe1\x91\xaa\x3b\x01\x7c\xf4\xcf\xaf\xe9\xe6\x5f\xe2\x37\x98\x4a\x1e\xe7\xe6\xa2\xf5\xc0\xfa\x36\x36\xbe\x4f\x61\x4d\x9a\x78\x7b\x6f\x42\x0c\xbe\xee\xc0\x2a\x27\xc7\x32\x3f\xe3\xaa\xf8\x61\x7c\xe2\xa9\x30\x84\x23\x9d\xe4\x45\x0f\xc3\xeb\x2e\x1b\x56\xa1\xee\x6a\x1c\x2a\x26\x29\xb2\xbd\x10\xbf\xe5\xca\x1c\xa9\xfc\xb7\xdc\xcf\x71\x3e\x85\x2f\x01\x14\xe6\xef\x53\x33\xdb\xe6\x5a\x24\xaf\x65\x0b\xb8\xcd\x36\x9d\x66\xad\x9a\x48\xb6\xff\x09\x00\x00\xff\xff\xb2\x38\x66\xb6\x9a\x01\x00\x00")

func _20180519220506_add_events_tableUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__20180519220506_add_events_tableUpSql,
		"20180519220506_add_events_table.up.sql",
	)
}

func _20180519220506_add_events_tableUpSql() (*asset, error) {
	bytes, err := _20180519220506_add_events_tableUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "20180519220506_add_events_table.up.sql", size: 410, mode: os.FileMode(420), modTime: time.Unix(1526917025, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
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
var _bindata = map[string]func() (*asset, error){
	"20180519220506_add_events_table.down.sql": _20180519220506_add_events_tableDownSql,
	"20180519220506_add_events_table.up.sql": _20180519220506_add_events_tableUpSql,
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
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"20180519220506_add_events_table.down.sql": &bintree{_20180519220506_add_events_tableDownSql, map[string]*bintree{}},
	"20180519220506_add_events_table.up.sql": &bintree{_20180519220506_add_events_tableUpSql, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

