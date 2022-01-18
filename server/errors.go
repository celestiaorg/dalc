package server

import "fmt"

type PreExistingBlockMappingError struct {
	OptimintBlockHeight, CelestiaBlockHeight int64
}

func (pe PreExistingBlockMappingError) Error() string {
	return fmt.Sprintf(
		"optimint block at height %d is already associated with a celestia block height %d",
		pe.OptimintBlockHeight,
		pe.CelestiaBlockHeight,
	)
}

type NoAssociatedBlockError struct {
	OptimintBlockHeight int64
}

func (ne NoAssociatedBlockError) Error() string {
	return fmt.Sprintf(
		"No assocated celestia block for optimint block of height %d",
		ne.OptimintBlockHeight,
	)
}
