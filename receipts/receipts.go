// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package receipts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type Receipt struct {
	Entry                  *JSON
	MerkleBranch           []*primitives.MerkleNode
	EntryBlockKeyMR        *primitives.Hash
	DirectoryBlockKeyMR    *primitives.Hash
	BitcoinTransactionHash *primitives.Hash
	BitcoinBlockHash       *primitives.Hash
}

func (e *Receipt) Validate() error {
	if e.Entry == nil {
		return fmt.Errorf("Receipt has no entry")
	}
	if e.MerkleBranch == nil {

	}
	entryHash, err := primitives.NewShaHashFromStr(e.Entry.Key)
	//TODO: validate entry hashes into EntryHash

	if err != nil {
		return err
	}
	var left interfaces.IHash
	var right interfaces.IHash
	var currentEntry interfaces.IHash
	currentEntry = entryHash
	eBlockFound := false
	dBlockFound := false
	for i, node := range e.MerkleBranch {
		if node.Left == nil {
			left = currentEntry
			right = node.Right
		} else {
			left = node.Left
			if node.Right == nil {
				right = currentEntry
			} else {
				right = node.Right
			}
		}
		if node.Right == nil {
			return fmt.Errorf("Node %v/%v has two nil sides", i, len(e.MerkleBranch))
		}
		if left.IsSameAs(currentEntry) == false && left.IsSameAs(currentEntry) {
			return fmt.Errorf("Entry %v not found in node %v/%v", currentEntry, i, len(e.MerkleBranch))
		}
		top := primitives.HashMerkleBranches(left, right)
		if node.Top != nil {
			if top.IsSameAs(node.Top) == false {
				return fmt.Errorf("Derived top %v is not the same as saved top in node %v/%v", top, i, len(e.MerkleBranch))
			}
		}
		if top.IsSameAs(e.EntryBlockKeyMR) == true {
			eBlockFound = true
		}
		if top.IsSameAs(e.DirectoryBlockKeyMR) == true {
			dBlockFound = true
		}
		currentEntry = top
	}

	if eBlockFound == false {
		return fmt.Errorf("EntryBlockKeyMR not found in branch")
	}

	if dBlockFound == false {
		return fmt.Errorf("DirectoryBlockKeyMR not found in branch")
	}

	return nil
}

func (e *Receipt) IsSameAs(r *Receipt) bool {
	if e.Entry == nil {
		if r.Entry != nil {
			return false
		}
	}
	if e.Entry.IsSameAs(r.Entry) == false {
		return false
	}

	//...

	return true
}

func (e *Receipt) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Receipt) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Receipt) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *Receipt) String() string {
	str, _ := e.JSONString()
	return str
}

func (e *Receipt) DecodeString(str string) error {
	jsonByte := []byte(str)
	err := json.Unmarshal(jsonByte, e)
	if err != nil {
		return err
	}
	return nil
}

func DecodeReceiptString(str string) (*Receipt, error) {
	receipt := new(Receipt)
	err := receipt.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return receipt, err
}

type JSON struct {
	Raw  string `json:",omitempty"`
	Key  string `json:",omitempty"`
	Json string `json:",omitempty"`
}

func (e *JSON) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *JSON) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *JSON) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *JSON) String() string {
	str, _ := e.JSONString()
	return str
}

func (e *JSON) IsSameAs(r *JSON) bool {
	if r == nil {
		return false
	}
	if e.Raw != r.Raw {
		return false
	}
	if e.Key != r.Key {
		return false
	}
	if e.Json != r.Json {
		return false
	}
	return true
}

func CreateFullReceipt(dbo interfaces.DBOverlay, entryID interfaces.IHash) (*Receipt, error) {
	receipt := new(Receipt)
	receipt.Entry = new(JSON)
	receipt.Entry.Key = entryID.String()

	//EBlock

	hash, err := dbo.LoadIncludedIn(entryID)
	if err != nil {
		return nil, err
	}

	if hash == nil {
		return nil, fmt.Errorf("Block containing entry not found")
	}

	eBlock, err := dbo.FetchEBlockByKeyMR(hash)
	if err != nil {
		return nil, err
	}

	if eBlock == nil {
		return nil, fmt.Errorf("EBlock not found")
	}

	hash = eBlock.DatabasePrimaryIndex()
	receipt.EntryBlockKeyMR = hash.(*primitives.Hash)

	entries := eBlock.GetEntryHashes()
	fmt.Printf("eBlock entries - %v\n\n", entries)
	branch := primitives.BuildMerkleBranchForEntryHash(entries, entryID, true)
	blockNode := new(primitives.MerkleNode)
	left, err := eBlock.HeaderHash()
	if err != nil {
		return nil, err
	}
	blockNode.Left = left.(*primitives.Hash)
	blockNode.Right = eBlock.BodyKeyMR().(*primitives.Hash)
	blockNode.Top = hash.(*primitives.Hash)
	fmt.Printf("eBlock blockNode - %v\n\n", blockNode)
	branch = append(branch, blockNode)
	receipt.MerkleBranch = append(receipt.MerkleBranch, branch...)

	str, _ := eBlock.JSONString()

	fmt.Printf("eBlock - %v\n\n", str)

	//DBlock

	hash, err = dbo.LoadIncludedIn(hash)
	if err != nil {
		return nil, err
	}

	if hash == nil {
		return nil, fmt.Errorf("Block containing EBlock not found")
	}

	dBlock, err := dbo.FetchDBlockByKeyMR(hash)
	if err != nil {
		return nil, err
	}

	if dBlock == nil {
		return nil, fmt.Errorf("DBlock not found")
	}

	str, _ = dBlock.JSONString()

	fmt.Printf("dBlock - %v\n\n", str)

	entries = dBlock.GetEntryHashesForBranch()
	fmt.Printf("dBlock entries - %v\n\n", entries)

	merkleTree := primitives.BuildMerkleTreeStore(entries)
	fmt.Printf("dBlock merkleTree - %v\n\n", merkleTree)

	branch = primitives.BuildMerkleBranchForEntryHash(entries, receipt.EntryBlockKeyMR, true)
	blockNode = new(primitives.MerkleNode)
	left, err = dBlock.HeaderHash()
	if err != nil {
		return nil, err
	}
	blockNode.Left = left.(*primitives.Hash)
	blockNode.Right = dBlock.BodyKeyMR().(*primitives.Hash)
	blockNode.Top = hash.(*primitives.Hash)
	fmt.Printf("dBlock blockNode - %v\n\n", blockNode)
	branch = append(branch, blockNode)
	receipt.MerkleBranch = append(receipt.MerkleBranch, branch...)

	//DirBlockInfo

	hash = dBlock.DatabasePrimaryIndex()
	receipt.DirectoryBlockKeyMR = hash.(*primitives.Hash)

	dirBlockInfo, err := dbo.FetchDirBlockInfoByKeyMR(hash)
	if err != nil {
		return nil, err
	}

	if dirBlockInfo == nil {
		return nil, fmt.Errorf("dirBlockInfo not found")
	}
	dbi := dirBlockInfo.(*dbInfo.DirBlockInfo)

	receipt.BitcoinTransactionHash = dbi.BTCTxHash.(*primitives.Hash)
	receipt.BitcoinBlockHash = dbi.BTCBlockHash.(*primitives.Hash)

	return receipt, nil
}

func CreateMinimalReceipt(dbo interfaces.DBOverlay, entryID interfaces.IHash) (*Receipt, error) {
	return nil, nil
}

func VerifyFullReceipt(dbo interfaces.DBOverlay, receiptStr string) error {
	receipt, err := DecodeReceiptString(receiptStr)
	if err != nil {
		return err
	}

	err = receipt.Validate()
	if err != nil {
		return err
	}

	return nil
}

func VerifyMinimalReceipt(dbo interfaces.DBOverlay, receiptStr string) error {
	return nil
}
