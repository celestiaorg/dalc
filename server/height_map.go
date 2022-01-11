package server

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// HeightMap keeps track of which optimint blocks where posted to which celestia
// blocks
type HeightMap struct {
	Heights map[string]int64
	mut     *sync.RWMutex
}

// Search returns the celestia height of which the optimint blocks were stored,
// if it exists
func (hm *HeightMap) Search(optimintHeight int64) (celestiaHeight int64, has bool) {
	strOptiHeight := fmt.Sprintf("%d", optimintHeight)
	hm.mut.RLock()
	defer hm.mut.RUnlock()
	celestiaHeight, has = hm.Heights[strOptiHeight]
	return
}

// Save records the mapping between optimint block height and the celestia block
// at which the optimint block exists
func (hm *HeightMap) Save(optimintHeight, celestiaHeight int64) error {
	hm.mut.Lock()
	defer hm.mut.Unlock()

	strOptiHeight := fmt.Sprintf("%d", optimintHeight)

	if cHeight, has := hm.Heights[strOptiHeight]; has {
		return fmt.Errorf(
			"optimint block at height %s already saved on celestia at celestia height %d",
			strOptiHeight,
			cHeight,
		)
	}
	hm.Heights[strOptiHeight] = celestiaHeight
	return nil
}

// Encode marshals the HeightMap using a json format
func (hm *HeightMap) Encode(w io.Writer) error {
	return json.NewEncoder(w).Encode(hm.Heights)
}

// DecodeHeightMap unmarshals a json formatted HeightMap
func DecodeHeightMap(r io.Reader) (HeightMap, error) {
	heights := make(map[string]int64)
	err := json.NewDecoder(r).Decode(&heights)
	if err != nil {
		return HeightMap{}, err
	}
	return HeightMap{
		Heights: heights,
		mut:     &sync.RWMutex{},
	}, nil
}

func HeightMapFromFile(path string) (HeightMap, error) {
	file, err := os.Open(path)
	if err != nil {
		return HeightMap{}, err
	}
	return DecodeHeightMap(file)
}

func (hm *HeightMap) SaveToFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	return hm.Encode(file)
}
