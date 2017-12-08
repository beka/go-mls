package mls

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

var (
	aLeafData = [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
		[]byte("d"),
		[]byte("e"),
	}
)

func TestMerkleTree(t *testing.T) {
	aLeaves := make([]MerkleNode, len(aLeafData))
	aLeafNodes := make([]Node, len(aLeafData))
	for i, data := range aLeafData {
		aLeaves[i] = MerkleNode{merkleLeaf(data)}
		aLeafNodes[i] = aLeaves[i]
	}

	ab, _ := merkleNodeDefn.combine(aLeaves[0], aLeaves[1])
	cd, _ := merkleNodeDefn.combine(aLeaves[2], aLeaves[3])
	abcd, _ := merkleNodeDefn.combine(merkleNodeDefn.create(ab), merkleNodeDefn.create(cd))
	abcde, _ := merkleNodeDefn.combine(merkleNodeDefn.create(abcd), aLeaves[4])

	tree, err := newTreeFromLeaves(merkleNodeDefn, aLeafNodes)
	if err != nil {
		t.Fatalf("Error building tree: %v", err)
	}

	root, err := tree.Root()
	if err != nil {
		t.Fatalf("Error fetching tree root: %v", err)
	}

	rootData, ok := root.(MerkleNode)
	if !ok {
		t.Fatalf("Merkle tree root not of type MerkleNode")
	}

	if !merkleNodeDefn.valid(root) {
		t.Fatalf("Merkle tree root is not valid")
	}

	if !merkleNodeDefn.valid(root) {
		t.Fatalf("Merkle tree root is not equal to itself")
	}

	if !bytes.Equal(rootData.Value, abcde) {
		t.Fatalf("Incorrect Merkle tree root: %x != %x", rootData, abcde)
	}
}

func TestECNodeJSON(t *testing.T) {
	aData := []byte("data")
	aKey := ECNodeFromData(aData)

	kj, err := json.Marshal(aKey)
	if err != nil {
		t.Fatalf("Error marshaling ECNode: %v", err)
	}

	k2 := new(ECNode)
	err = json.Unmarshal(kj, k2)
	if err != nil {
		t.Fatalf("Error unmarshaling ECNode: %v", err)
	}

	if !ecdhNodeDefn.publicEqual(aKey, k2) {
		t.Fatalf("JSON round-trip failed: %v != %v", aKey, k2)
	}
}

func TestECDHTree(t *testing.T) {
	aLeaves := make([]*ECNode, len(aLeafData))
	aLeafNodes := make([]Node, len(aLeafData))
	for i, data := range aLeafData {
		aLeaves[i] = ECNodeFromData(data)
		aLeafNodes[i] = aLeaves[i]
	}

	ab := ECNodeFromData(aLeaves[0].PrivateKey.derive(aLeaves[1].PrivateKey.PublicKey))
	cd := ECNodeFromData(aLeaves[2].PrivateKey.derive(aLeaves[3].PrivateKey.PublicKey))
	abcd := ECNodeFromData(ab.PrivateKey.derive(cd.PrivateKey.PublicKey))
	abcde := ECNodeFromData(abcd.PrivateKey.derive(aLeaves[4].PrivateKey.PublicKey))

	tree, err := newTreeFromLeaves(ecdhNodeDefn, aLeafNodes)
	if err != nil {
		t.Fatalf("Error building tree: %v", err)
	}

	root, err := tree.Root()
	if err != nil {
		t.Fatalf("Error fetching tree root: %v", err)
	}

	rootData, ok := root.(*ECNode)
	if !ok {
		t.Fatalf("ECDH tree root not of type *ecdhKey")
	}

	if !ecdhNodeDefn.valid(root) {
		t.Fatalf("ECDH tree root is not valid")
	}

	if !ecdhNodeDefn.valid(root) {
		t.Fatalf("ECDH tree root is not equal to itself")
	}

	if !reflect.DeepEqual(rootData.PrivateKey.PublicKey, abcde.PrivateKey.PublicKey) {
		t.Fatalf("Incorrect ECDH tree root: %x != %x", rootData, abcde)
	}
}
