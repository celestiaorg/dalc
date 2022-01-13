package server

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"sync"
)

const HeightMapFileName = "height_map.json"

// HeightMapper keeps track of which optimint blocks where posted to which celestia
// blocks
type HeightMapper struct {
	Heights map[string]int64
	mut     *sync.RWMutex
}

// Search returns the celestia height of which the optimint blocks were stored,
// if it exists
func (hm *HeightMapper) Search(optimintHeight int64) (celestiaHeight int64, has bool) {
	strOptiHeight := strconv.FormatInt(optimintHeight, 10)
	hm.mut.RLock()
	defer hm.mut.RUnlock()
	celestiaHeight, has = hm.Heights[strOptiHeight]
	return
}

// Save records the mapping between optimint block height and the celestia block
// at which the optimint block exists
func (hm *HeightMapper) Save(optimintHeight, celestiaHeight int64) error {
	hm.mut.Lock()
	defer hm.mut.Unlock()

	strOptiHeight := strconv.FormatInt(optimintHeight, 10)

	if cHeight, has := hm.Heights[strOptiHeight]; has {
		return PreExistingBlockMappingError{
			CelestiaBlockHeight: cHeight,
			OptimintBlockHeight: optimintHeight,
		}
	}
	hm.Heights[strOptiHeight] = celestiaHeight
	return nil
}

// Encode marshals the HeightMapper using a json format
func (hm *HeightMapper) Encode(w io.Writer) error {
	return json.NewEncoder(w).Encode(hm.Heights)
}

// DecodeHeightMapper unmarshals a json formatted HeightMapper
func DecodeHeightMapper(r io.Reader) (HeightMapper, error) {
	heights := make(map[string]int64)
	err := json.NewDecoder(r).Decode(&heights)
	if err != nil {
		return HeightMapper{}, err
	}
	return HeightMapper{
		Heights: heights,
		mut:     &sync.RWMutex{},
	}, nil
}

func HeightMapperFromFile(path string) (HeightMapper, error) {
	file, err := os.Open(path)
	if err != nil {
		return HeightMapper{}, err
	}
	return DecodeHeightMapper(file)
}

func (hm *HeightMapper) SaveToFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	return hm.Encode(file)
}
