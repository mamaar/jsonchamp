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
	values     []node
}

func (b *bitmasked) index(pos uint64) int {
	return bits.OnesCount64((b.valueMap | b.subMapsMap) & (pos - 1))
}

func (b *bitmasked) keys() []string {
	keys := make([]string, 0, len(b.values))

	for _, v := range b.values {
		switch value := v.(type) {
		case value:
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

		return b.values[valueIdx].get(key)
	}

	if b.subMapsMap&pos != 0 {
		subMapIndex := b.index(pos)

		subMap, ok := b.values[subMapIndex].(*bitmasked)
		if !ok {
			panic(fmt.Sprintf("submap not correct type: %s, %T", key.key, b.values[subMapIndex]))
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
			values: []node{
				b.mergeValueToSubNode(newLevel+1, keyA, valueA, keyB, valueB),
			},
		}
	}

	if posA < posB {
		return &bitmasked{
			level:      newLevel,
			valueMap:   posA | posB,
			subMapsMap: 0,
			values:     []node{value{key: keyA, value: valueA}, value{key: keyB, value: valueB}},
		}
	}

	return &bitmasked{
		level:      newLevel,
		valueMap:   posA | posB,
		subMapsMap: 0,
		values:     []node{value{key: keyB, value: valueB}, value{key: keyA, value: valueA}},
	}
}

func (b *bitmasked) set(key key, newValue any) node {
	pos := bitPosition(key.hash, b.level)

	// If there's already a leaf node at this position, we need to merge the values.
	valueExists := b.valueMap&pos != 0
	if valueExists {
		valueIdx := b.index(pos)

		existingValue, ok := b.values[valueIdx].(value)
		if !ok {
			panic(fmt.Sprintf("value not correct type: %s, %T", key.key, b.values[valueIdx]))
		}
		// If it's the same key, we can just update the value.
		if existingValue.key.hash == key.hash && existingValue.key.key == key.key {
			newValues := make([]node, len(b.values), len(b.values))
			copy(newValues, b.values)
			newValues[valueIdx] = value{key: key, value: newValue}

			return &bitmasked{
				level:      b.level,
				valueMap:   b.valueMap,
				subMapsMap: b.subMapsMap,
				values:     newValues,
			}
		}

		// If it's a different key, we need to handle the collision

		newValues := make([]node, len(b.values), len(b.values))
		copy(newValues, b.values)
		newValues[valueIdx] = b.mergeValueToSubNode(b.level+1, existingValue.key, existingValue.value, key, newValue)

		return &bitmasked{
			level:      b.level,
			valueMap:   b.valueMap ^ pos,
			subMapsMap: b.subMapsMap | pos,
			values:     newValues,
		}
	}

	nodeExists := b.subMapsMap&pos != 0
	if nodeExists {
		subNodeIndex := b.index(pos)
		subNode, ok := b.values[subNodeIndex].(*bitmasked)

		if !ok {
			panic(fmt.Sprintf("subnode not correct type: %s, %T", key.key, b.values[subNodeIndex]))
		}

		newValues := make([]node, len(b.values), len(b.values))
		copy(newValues, b.values)

		newValues[subNodeIndex] = subNode.set(key, newValue)

		return &bitmasked{
			level:      b.level,
			valueMap:   b.valueMap,
			subMapsMap: b.subMapsMap,
			values:     newValues,
		}
	}

	newValueIndex := b.index(pos)

	copiedValues := make([]node, len(b.values), len(b.values))
	copy(copiedValues, b.values)

	var before []node
	if len(copiedValues) > 0 {
		before = copiedValues[:newValueIndex]
	} else {
		before = []node{}
	}

	var after []node
	if newValueIndex < len(copiedValues) {
		after = copiedValues[newValueIndex:]
	} else {
		after = []node{}
	}

	var newValues = make([]node, 0, len(b.values)+1)
	newValues = append(newValues, before...)
	newValues = append(newValues, value{key: key, value: newValue})
	newValues = append(newValues, after...)

	return &bitmasked{
		valueMap:   b.valueMap | pos,
		subMapsMap: b.subMapsMap,
		level:      b.level,
		values:     newValues,
	}
}

func (b *bitmasked) copy() node {
	newValues := make([]node, len(b.values), len(b.values))

	for i, v := range b.values {
		newValues[i] = v.copy()
	}

	return &bitmasked{
		level:      b.level,
		valueMap:   b.valueMap,
		subMapsMap: b.subMapsMap,
		values:     newValues,
	}
}

// delete deletes a key from the map. If the key does not exist, it returns false.
func (b *bitmasked) delete(key key) (*bitmasked, bool) {
	if len(b.values) == 0 {
		return b, false
	}

	pos := bitPosition(key.hash, b.level)

	valueExists := b.valueMap&pos != 0
	if valueExists {
		valueIdx := b.index(pos)

		newMap := b.copy().(*bitmasked)
		newMap.valueMap = b.valueMap ^ pos
		newMap.values = append(b.values[:valueIdx], b.values[valueIdx+1:]...)

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
	subNode, ok := b.values[subNodeIndex].(*bitmasked)

	if !ok {
		panic(fmt.Sprintf("subnode not correct type: %s, %T", key.key, b.values[subNodeIndex]))
	}

	subNodeCopy, ok := subNode.copy().(*bitmasked)
	if !ok {
		panic("expected bitmasked")
	}

	subNodeCopy, ok = subNodeCopy.delete(key)

	// The last key in the subnode was deleted, so we remove the subnode.
	if len(subNodeCopy.values) == 0 {
		newMap, isBitmasked := b.copy().(*bitmasked)
		if !isBitmasked {
			panic("expected bitmasked")
		}

		newMap.subMapsMap = b.subMapsMap ^ pos
		newMap.values = append(b.values[:subNodeIndex], b.values[subNodeIndex+1:]...)

		return newMap, ok
	}

	newB, ok := b.copy().(*bitmasked)
	if !ok {
		panic("expected bitmasked")
	}

	newB.values[subNodeIndex] = subNodeCopy

	return newB, ok
}

var _ node = &bitmasked{
	level:      0,
	valueMap:   0,
	subMapsMap: 0,
	values:     nil,
}
