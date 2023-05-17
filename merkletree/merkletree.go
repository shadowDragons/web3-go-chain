package merkletree

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"go-chain/transaction"
	"go-chain/utils"
)

type MerkleTree struct {
	RootNode *MerkleTreeNode
}

type MerkleTreeNode struct {
	LeftNode  *MerkleTreeNode
	RightNode *MerkleTreeNode
	Data      []byte
}

func CreateNode(left, right *MerkleTreeNode, data []byte) *MerkleTreeNode {
	tmpNode := MerkleTreeNode{}
	if left == nil && right == nil {
		tmpNode.Data = data
	} else {
		catenateHash := append(left.Data, right.Data...)
		hash := sha256.Sum256(catenateHash)
		tmpNode.Data = hash[:]
	}

	tmpNode.LeftNode = left
	tmpNode.RightNode = right

	return &tmpNode
}

func CrateMerkleTree(txs []*transaction.Transaction) *MerkleTree {
	txsLen := len(txs)

	if txsLen%2 != 0 {
		txs = append(txs, txs[txsLen-1])
	}

	var nodePool []*MerkleTreeNode
	for _, tx := range txs {
		nodePool = append(nodePool, CreateNode(nil, nil, tx.ID))
	}

	for len(nodePool) > 1 {
		var tempNodePool []*MerkleTreeNode

		poolLen := len(nodePool)
		if poolLen%2 != 0 {
			tempNodePool = append(tempNodePool, nodePool[poolLen-1])
		}

		for i := 0; i < poolLen/2; i++ {
			tempNodePool = append(tempNodePool, CreateNode(nodePool[2*i], nodePool[2*i+1], nil))
		}

		nodePool = tempNodePool
	}

	return &MerkleTree{nodePool[0]}
}

func (mn *MerkleTreeNode) Find(data []byte, route []int, hashroute [][]byte) (bool, []int, [][]byte) {
	findFlag := false

	if bytes.Equal(mn.Data, data) {
		findFlag = true
		return findFlag, route, hashroute
	} else {
		if mn.LeftNode != nil {
			route_t := append(route, 0)
			hashroute_t := append(hashroute, mn.RightNode.Data)
			findFlag, route_t, hashroute_t = mn.LeftNode.Find(data, route_t, hashroute_t)
			if findFlag {
				return findFlag, route_t, hashroute_t
			} else {
				if mn.RightNode != nil {
					route_t = append(route, 1)
					hashroute_t = append(hashroute, mn.LeftNode.Data)
					findFlag, route_t, hashroute_t = mn.RightNode.Find(data, route_t, hashroute_t)
					if findFlag {
						return findFlag, route_t, hashroute_t
					} else {
						return findFlag, route, hashroute
					}

				}
			}
		} else {
			return findFlag, route, hashroute
		}
	}
	return findFlag, route, hashroute
}

func (mt *MerkleTree) BackValidationRoute(txid []byte) ([]int, [][]byte, bool) {
	ok, route, hashroute := mt.RootNode.Find(txid, []int{}, [][]byte{})
	return route, hashroute, ok
}

func SimplePaymentValidation(txid, mtroothash []byte, route []int, hashroute [][]byte) bool {
	routeLen := len(route)
	var tempHash []byte
	tempHash = txid

	for i := routeLen - 1; i >= 0; i-- {
		if route[i] == 0 {
			catenateHash := append(tempHash, hashroute[i]...)
			hash := sha256.Sum256(catenateHash)
			tempHash = hash[:]
		} else if route[i] == 1 {
			catenateHash := append(hashroute[i], tempHash...)
			hash := sha256.Sum256(catenateHash)
			tempHash = hash[:]
		} else {
			utils.Handle(errors.New("error in validation route"))
		}
	}
	return bytes.Equal(tempHash, mtroothash)
}
