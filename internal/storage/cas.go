package storage

import (
	"os"
	"path/filepath"
)

type CAS struct {
	Root string // e.g. ".glazel/cas"
}

func NewCAS(root string) *CAS { return &CAS{Root: root} }

func (c *CAS) ObjPath(hash string) string {
	return filepath.Join(c.Root, "obj", hash+".o")
}

func (c *CAS) EnsureDirs() error {
	return os.MkdirAll(filepath.Join(c.Root, "obj"), 0755)
}

func (c *CAS) HasObj(hash string) bool {
	_, err := os.Stat(c.ObjPath(hash))
	return err == nil
}

func (c *CAS) PutObj(hash string, b []byte) (string, error) {
	if err := c.EnsureDirs(); err != nil {
		return "", err
	}
	p := c.ObjPath(hash)
	// Write once; overwrite is fine for now
	if err := os.WriteFile(p, b, 0644); err != nil {
		return "", err
	}
	return p, nil
}

func (c *CAS) GetObj(hash string) ([]byte, error) {
	return os.ReadFile(c.ObjPath(hash))
}
