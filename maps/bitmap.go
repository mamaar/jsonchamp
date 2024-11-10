package maps

import (
	"fmt"
	"math/bits"
)

const (
	MaskLevel0 = (0b11111100 << 56)
)

func bitPosition(hash uint64, level uint8) uint64 {
	return 1 << partition(hash, level)
}

func partitionMask(level uint8) uint64 {
	return MaskLevel0 >> (BitPartitionSize * level)
}

func partition(hash uint64, level uint8) uint64 {
	if level == 10 {
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

func (b *bitmasked) Keys() []string {
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
			keys = append(keys, value.Keys()...)
		}
	}
	return keys
}

// Get implements node.
func (b *bitmasked) Get(key Key) (any, bool) {
	pos := bitPosition(key.hash, b.level)

	if b.valueMap&pos != 0 {
		valueIdx := b.index(pos)
		return b.values[valueIdx].Get(key)
	}
	if b.subMapsMap&pos != 0 {
		subMapIndex := b.index(pos)
		subMap := b.values[subMapIndex].(*bitmasked)
		return subMap.Get(key)
	}

	return nil, false
}

func (b *bitmasked) mergeValueToSubNode(newLevel uint8, keyA Key, valueA any, keyB Key, valueB any) node {
	if b.level >= 10 {
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

func (b *bitmasked) Set(key Key, newValue any) node {
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
			newB := b.copy()
			newB.values[valueIdx] = &value{key: key, value: newValue}
			return newB
		}

		// If it's a different key, we need to handle the collision
		newMap := b.copy()
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
		newB := b.copy()
		newB.values[subNodeIndex] = subNode.Set(key, newValue)
		return newB
	}

	newB := b.copy()

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

func (b *bitmasked) copy() *bitmasked {
	newValues := make([]node, len(b.values))
	copy(newValues, b.values)
	return &bitmasked{
		level:      b.level,
		valueMap:   b.valueMap,
		subMapsMap: b.subMapsMap,
		values:     newValues,
	}
}

// Delete deletes a key from the map. If the key does not exist, it returns false.
func (b *bitmasked) Delete(key Key) (*bitmasked, bool) {
	if len(b.values) == 0 {
		return b, false
	}
	pos := bitPosition(key.hash, b.level)

	valueExists := b.valueMap&pos != 0
	if valueExists {
		valueIdx := b.index(pos)

		newMap := b.copy()
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
		subNodeCopy := subNode.copy()
		subNodeCopy, ok = subNodeCopy.Delete(key)

		// The last key in the subnode was deleted, so we remove the subnode.
		if len(subNodeCopy.values) == 0 {
			newMap := b.copy()
			newMap.subMapsMap = b.subMapsMap ^ pos
			newMap.values = append(b.values[:subNodeIndex], b.values[subNodeIndex+1:]...)
			return newMap, ok
		}

		newB := b.copy()
		newB.values[subNodeIndex] = subNodeCopy
		return newB, ok
	}

	return b, false
}

var _ node = &bitmasked{}
