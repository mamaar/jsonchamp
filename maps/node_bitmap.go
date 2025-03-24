package maps

import (
	"fmt"
	"math/bits"
)

const (
	BitPartitionSize = 6
	MaskLevel0       = 0b11111100 << 56

	maxTreeDepth = 10
)

func bitPosition(hash uint64, level uint8) uint64 {
	return 1 << partition(hash, level)
}

func partitionMask(level uint8) uint64 {
	return MaskLevel0 >> (BitPartitionSize * level)
}

func partition(hash uint64, level uint8) uint64 {
	// At the max level, we only need the last 4 bits.
	// Because that's what's left after the previous levels.
	if level == maxTreeDepth {
		return hash & 0b0000_1111
	}
	startBit := 64 - (level+1)*BitPartitionSize
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
func (b *bitmasked) get(key Key) (any, bool) {
	pos := bitPosition(key.hash, b.level)

	if b.valueMap&pos != 0 {
		valueIdx := b.index(pos)
		return b.values[valueIdx].get(key)
	}
	if b.subMapsMap&pos != 0 {
		subMapIndex := b.index(pos)
		subMap := b.values[subMapIndex].(*bitmasked)
		return subMap.get(key)
	}

	return nil, false
}

func (b *bitmasked) mergeValueToSubNode(newLevel uint8, keyA Key, valueA any, keyB Key, valueB any) node {
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
			values:     []node{&value{key: keyA, value: valueA}, &value{key: keyB, value: valueB}},
		}
	} else {
		return &bitmasked{
			level:      newLevel,
			valueMap:   posA | posB,
			subMapsMap: 0,
			values:     []node{&value{key: keyB, value: valueB}, &value{key: keyA, value: valueA}},
		}
	}
}

func (b *bitmasked) set(key Key, newValue any) node {
	pos := bitPosition(key.hash, b.level)

	valueExists := b.valueMap&pos != 0
	if valueExists {
		valueIdx := b.index(pos)

		existingValue, ok := b.values[valueIdx].(*value)
		if !ok {
			panic(fmt.Sprintf("value not correct type: %s, %T", key.key, b.values[valueIdx]))
		}
		// If it's the same key, we can just update the value.
		if existingValue.key.hash == key.hash && existingValue.key.key == key.key {
			newB := b.copy().(*bitmasked)
			newB.values[valueIdx] = &value{key: key, value: newValue}
			return newB
		}

		// If it's a different key, we need to handle the collision
		newMap := b.copy().(*bitmasked)
		newMap.valueMap = b.valueMap ^ pos
		newMap.subMapsMap = b.subMapsMap | pos
		newMap.values[valueIdx] = b.mergeValueToSubNode(b.level+1, existingValue.key, existingValue.value, key, newValue)
		return newMap
	}

	nodeExists := b.subMapsMap&pos != 0
	if nodeExists {
		subNodeIndex := b.index(pos)
		subNode, ok := b.values[subNodeIndex].(*bitmasked)
		if !ok {
			panic(fmt.Sprintf("subnode not correct type: %s, %T", key.key, b.values[subNodeIndex]))
		}
		newB := b.copy().(*bitmasked)
		newB.values[subNodeIndex] = subNode.set(key, newValue)
		return newB
	}

	newB := b.copy().(*bitmasked)

	newValueIndex := b.index(pos)

	var before []node
	if len(newB.values) > 0 {
		before = newB.values[:newValueIndex]
	} else {
		before = []node{}
	}

	var after []node
	if newValueIndex < len(newB.values) {
		after = newB.values[newValueIndex:]
	} else {
		after = []node{}
	}
	var newValues = make([]node, 0, len(b.values)+1)
	newValues = append(newValues, before...)
	newValues = append(newValues, &value{key: key, value: newValue})
	newValues = append(newValues, after...)
	newB.valueMap |= pos
	newB.values = newValues
	return newB
}

func (b *bitmasked) copy() node {
	newValues := make([]node, len(b.values))

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
func (b *bitmasked) delete(key Key) (*bitmasked, bool) {
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
	if nodeExists {
		subNodeIndex := b.index(pos)
		subNode, ok := b.values[subNodeIndex].(*bitmasked)
		if !ok {
			panic(fmt.Sprintf("subnode not correct type: %s, %T", key.key, b.values[subNodeIndex]))
		}
		subNodeCopy := subNode.copy().(*bitmasked)
		subNodeCopy, ok = subNodeCopy.delete(key)

		// The last key in the subnode was deleted, so we remove the subnode.
		if len(subNodeCopy.values) == 0 {
			newMap := b.copy().(*bitmasked)
			newMap.subMapsMap = b.subMapsMap ^ pos
			newMap.values = append(b.values[:subNodeIndex], b.values[subNodeIndex+1:]...)
			return newMap, ok
		}

		newB := b.copy().(*bitmasked)
		newB.values[subNodeIndex] = subNodeCopy
		return newB, ok
	}

	return b, false
}

var _ node = &bitmasked{}
