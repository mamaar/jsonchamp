package jsonchamp

import (
	"fmt"
	"math/bits"
)

const (
	bitPartitionSize = 6
	maskLevel0       = 0b11111100 << 56
	maskLevel10      = 0b0000_1111
	maxTreeDepth     = 10
	numKeyBits       = 64
)

func bitPosition(hash uint64, level uint8) uint64 {
	return 1 << partition(hash, level)
}

func partitionMask(level uint8) uint64 {
	return maskLevel0 >> (bitPartitionSize * level)
}

func partition(hash uint64, level uint8) uint64 {
	// At the max level, we only need the last 4 bits.
	// Because that's what's left after the previous levels.
	if level == maxTreeDepth {
		return hash & maskLevel10
	}

	startBit := numKeyBits - (level+1)*bitPartitionSize

	return (hash & partitionMask(level)) >> uint64(startBit)
}

type bitmasked struct {
	level      uint8
	valueMap   uint64
	subMapsMap uint64
	values     *cowSlice
}

func (b *bitmasked) index(pos uint64) int {
	return bits.OnesCount64((b.valueMap | b.subMapsMap) & (pos - 1))
}

func (b *bitmasked) keys() []string {
	keys := make([]string, 0, b.values.Len())

	for _, v := range b.values.Values() {
		switch value := v.(type) {
		case *value:
			keys = append(keys, value.key.key)
		case *collision:
			for _, v := range value.values {
				keys = append(keys, v.key.key)
			}
		case *bitmasked:
			keys = append(keys, value.keys()...)
		}
	}

	return keys
}

// Get implements node.
func (b *bitmasked) get(key key) (any, bool) {
	pos := bitPosition(key.hash, b.level)

	if b.valueMap&pos != 0 {
		valueIdx := b.index(pos)

		val := b.values.Get(valueIdx)

		return val.get(key)
	}

	if b.subMapsMap&pos != 0 {
		subMapIndex := b.index(pos)

		subMapNode := b.values.Get(subMapIndex)

		subMap, ok := subMapNode.(*bitmasked)
		if !ok {
			panic(fmt.Sprintf("submap not correct type: %s, %T", key.key, subMapNode))
		}

		return subMap.get(key)
	}

	return nil, false
}

func (b *bitmasked) mergeValueToSubNode(newLevel uint8, keyA key, valueA any, keyB key, valueB any) node {
	if b.level >= maxTreeDepth {
		panic("Max level reached")
	}

	posA := bitPosition(keyA.hash, newLevel)
	posB := bitPosition(keyB.hash, newLevel)

	// Collides on next level
	if posA == posB {
		return &bitmasked{
			level:      newLevel,
			valueMap:   0,
			subMapsMap: posA,
			values: newCowSliceWithItems(
				b.mergeValueToSubNode(newLevel+1, keyA, valueA, keyB, valueB),
			),
		}
	}

	if posA < posB {
		return &bitmasked{
			level:      newLevel,
			valueMap:   posA | posB,
			subMapsMap: 0,
			values: newCowSliceWithItems(
				&value{key: keyA, value: valueA},
				&value{key: keyB, value: valueB},
			),
		}
	}

	return &bitmasked{
		level:      newLevel,
		valueMap:   posA | posB,
		subMapsMap: 0,
		values: newCowSliceWithItems(
			&value{key: keyB, value: valueB},
			&value{key: keyA, value: valueA},
		),
	}
}

func (b *bitmasked) set(key key, newValue any) node {
	currentSubNode := b

	pos := bitPosition(key.hash, currentSubNode.level)

	valueExists := currentSubNode.valueMap&pos != 0
	subNodeExists := currentSubNode.subMapsMap&pos != 0

	valueIdx := currentSubNode.index(pos)

	if valueExists && subNodeExists {
		panic(fmt.Sprintf("both value and subnode exist at the same position: %s", key.key))
	}

	var indexedNode node
	if currentSubNode.values.Len() > valueIdx {
		indexedNode = currentSubNode.values.Get(valueIdx)
	}

	switch {
	// The hash partition exists and is a sub node.
	// We will recursively set the value in the sub node.
	case subNodeExists:
		subNode, ok := indexedNode.(*bitmasked)
		if !ok {
			panic(fmt.Sprintf("subnode not correct type: %s, %T", key.key, indexedNode))
		}

		return &bitmasked{
			level:      currentSubNode.level,
			valueMap:   currentSubNode.valueMap,
			subMapsMap: currentSubNode.subMapsMap,
			values:     currentSubNode.values.Set(valueIdx, subNode.set(key, newValue)),
		}

	// The leaf node exists.
	// This branch will set the new value if the key exists.
	// If the hash collides, it will create a new sub node.
	case valueExists:
		existingValue, ok := indexedNode.(*value)
		if !ok {
			panic(fmt.Sprintf("value not correct type: %s, %T", key.key, indexedNode))
		}

		if existingValue.key.hash == key.hash && existingValue.key.key == key.key {
			return &bitmasked{
				level:      currentSubNode.level,
				valueMap:   currentSubNode.valueMap,
				subMapsMap: currentSubNode.subMapsMap,
				values:     currentSubNode.values.Set(valueIdx, &value{key: key, value: newValue}),
			}
		}

		return &bitmasked{
			level:      currentSubNode.level,
			valueMap:   currentSubNode.valueMap ^ pos,
			subMapsMap: currentSubNode.subMapsMap | pos,
			values: currentSubNode.values.Set(valueIdx,
				currentSubNode.mergeValueToSubNode(
					currentSubNode.level+1,
					existingValue.key,
					existingValue.value,
					key,
					newValue,
				),
			),
		}

	// The hash partition does not exist.
	// We will create a new value in the current node.
	default:
		return &bitmasked{
			valueMap:   currentSubNode.valueMap | pos,
			subMapsMap: currentSubNode.subMapsMap,
			level:      currentSubNode.level,
			values:     currentSubNode.values.Insert(valueIdx, &value{key: key, value: newValue}),
		}
	}

}

func (b *bitmasked) copy() node {
	return &bitmasked{
		level:      b.level,
		valueMap:   b.valueMap,
		subMapsMap: b.subMapsMap,
		values:     b.values.Share(),
	}
}

// delete deletes a key from the map. If the key does not exist, it returns false.
func (b *bitmasked) delete(key key) (*bitmasked, bool) {
	if b.values.Len() == 0 {
		return b, false
	}

	pos := bitPosition(key.hash, b.level)

	valueExists := b.valueMap&pos != 0
	if valueExists {
		valueIdx := b.index(pos)

		newMap := b.copy().(*bitmasked)
		newMap.valueMap = b.valueMap ^ pos
		newMap.values = newMap.values.Delete(valueIdx)

		return newMap, true
	}

	nodeExists := b.subMapsMap&pos != 0
	if !nodeExists {
		newB, ok := b.copy().(*bitmasked)
		if !ok {
			panic("expected bitmasked")
		}

		return newB, false
	}

	subNodeIndex := b.index(pos)
	subNode, ok := b.values.Get(subNodeIndex).(*bitmasked)

	if !ok {
		panic(fmt.Sprintf("subnode not correct type: %s, %T", key.key, subNode))
	}

	subNodeCopy, ok := subNode.copy().(*bitmasked)
	if !ok {
		panic("expected bitmasked")
	}

	subNodeCopy, ok = subNodeCopy.delete(key)

	// The last key in the subnode was deleted, so we remove the subnode.
	if subNodeCopy.values.Len() == 0 {
		newMap, isBitmasked := b.copy().(*bitmasked)
		if !isBitmasked {
			panic("expected bitmasked")
		}

		newMap.subMapsMap = b.subMapsMap ^ pos
		newMap.values = newMap.values.Delete(subNodeIndex)

		return newMap, ok
	}

	newB, ok := b.copy().(*bitmasked)
	if !ok {
		panic("expected bitmasked")
	}

	newB.values = newB.values.Set(subNodeIndex, subNodeCopy)

	return newB, ok
}

var _ node = &bitmasked{
	level:      0,
	valueMap:   0,
	subMapsMap: 0,
	values:     nil,
}
