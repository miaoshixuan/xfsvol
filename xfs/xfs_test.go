package xfs_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cirocosta/xfsvol/xfs"
	"github.com/stretchr/testify/assert"
)

// setupTestFs takes a filesystem description as
// a variable and setups the desired structure under
// a temp directory.
func setupTestFs(base string, fs []string) (root string, err error) {
	var (
		extension string
		directory string
	)

	root, err = ioutil.TempDir(base, "")
	if err != nil {
		return
	}

	for _, element := range fs {
		element = filepath.Join(root, element)
		extension = filepath.Ext(element)
		directory = filepath.Dir(element)

		err = os.MkdirAll(directory, 0755)
		if err != nil && !os.IsExist(err) {
			return
		}

		if extension != "" {
			_, err = os.Create(element)
		} else {
			err = os.MkdirAll(element, 0755)
		}
		if err != nil {
			return
		}
	}

	return
}

// makeBigString creates a string filled with a single
// character that is `size` big.
func makeBigString(size int) (res string) {
	var buffer = make([]byte, size)
	for ndx := range buffer {
		buffer[ndx] = 'a'
	}

	res = string(buffer)
	return
}

func TestGetProjectId(t *testing.T) {
	var testCases = []struct {
		desc        string
		projectId   uint32
		fs          []string
		target      string
		shouldError bool
	}{
		{
			desc:        "fails if not a directory",
			fs:          []string{"/file.txt"},
			target:      "file.txt",
			shouldError: true,
		},
		{
			desc:        "fails if directory doesnt exist",
			fs:          []string{"/"},
			target:      "dir",
			shouldError: true,
		},
		{
			desc:        "returns zero if no attributes associated",
			fs:          []string{"/dir"},
			target:      "dir",
			shouldError: false,
			projectId:   0,
		},
	}

	var (
		err       error
		root      string
		projectId uint32
	)

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			root, err = setupTestFs("", tc.fs)
			assert.NoError(t, err)
			defer os.RemoveAll(root)

			projectId, err = xfs.GetProjectId(filepath.Join(root, tc.target))
			if tc.shouldError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.projectId, projectId)
		})
	}

}

func TestMakeBackingFsDev(t *testing.T) {
	var testCases = []struct {
		desc       string
		root       string
		file       string
		fs         []string
		shouldFail bool
	}{
		{
			desc:       "fails if root doesnt exist",
			root:       "/dir",
			file:       "dev",
			fs:         []string{},
			shouldFail: true,
		},
		{
			desc: "fails if root is not a dir",
			root: "/root.txt",
			file: "dev",
			fs: []string{
				"/root.txt",
			},
			shouldFail: true,
		},
		{
			desc: "fails if file name is too long",
			root: "/",
			file: makeBigString(1 << 20),
			fs: []string{
				"/",
			},
			shouldFail: true,
		},
		{
			desc: "succeeds if root exists and is a dir",
			root: "/dir",
			file: "dev",
			fs: []string{
				"/dir",
			},
			shouldFail: false,
		},
	}

	var (
		root string
		err  error
	)

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			root, err = setupTestFs("", tc.fs)
			assert.NoError(t, err)
			defer os.RemoveAll(root)

			err = xfs.MakeBackingFsDev(filepath.Join(root, tc.root), tc.file)
			if tc.shouldFail {
				assert.Error(t, err)
			}
		})
	}
}