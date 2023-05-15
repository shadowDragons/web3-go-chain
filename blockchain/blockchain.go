package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"go-chain/constcoe"
	"go-chain/utils"
	"io/ioutil"
	"os"
)

type BlockChain struct {
	Blocks   []*Block
	LastHash []byte
}

// blockchain.go
func (bc *BlockChain) AddBlock(newBlock *Block) {
	bc.Blocks = append(bc.Blocks, newBlock)
}

func InitBlockChain(address []byte) {
	blockchain := BlockChain{}
	blockchain.Blocks = append(blockchain.Blocks, GenesisBlock(address))
	blockchain.SaveFile()
}

func CreateBlockChain() *BlockChain {
	blockchain := BlockChain{}
	blockchain.LoadFile()
	return &blockchain
}

// blockchain.go
func (bc *BlockChain) FindUnspentTransactions(address []byte) []Transaction {
	var unSpentTxs []Transaction
	spentTxs := make(map[string][]int) // can't use type []byte as key value
	for idx := len(bc.Blocks) - 1; idx >= 0; idx-- {
		block := bc.Blocks[idx]
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		IterOutputs:
			for outIdx, out := range tx.Outputs {
				if spentTxs[txID] != nil {
					for _, spentOut := range spentTxs[txID] {
						if spentOut == outIdx {
							continue IterOutputs
						}
					}
				}

				if out.ToAddressRight(address) {
					unSpentTxs = append(unSpentTxs, *tx)
				}
			}
			if !tx.IsBase() {
				for _, in := range tx.Inputs {
					if in.FromAddressRight(address) {
						inTxID := hex.EncodeToString(in.TxID)
						spentTxs[inTxID] = append(spentTxs[inTxID], in.OutIdx)
					}
				}
			}
		}
	}
	return unSpentTxs
}

// blockchain.go
func (bc *BlockChain) FindUTXOs(address []byte) (int, map[string]int) {
	unspentOuts := make(map[string]int)
	unspentTxs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Outputs {
			if out.ToAddressRight(address) {
				accumulated += out.Value
				unspentOuts[txID] = outIdx
				continue Work // one transaction can only have one output referred to adderss
			}
		}
	}
	return accumulated, unspentOuts
}

// blockchain.go
func (bc *BlockChain) FindSpendableOutputs(address []byte, amount int) (int, map[string]int) {
	unspentOuts := make(map[string]int)
	unspentTxs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Outputs {
			if out.ToAddressRight(address) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = outIdx
				if accumulated >= amount {
					break Work
				}
				continue Work // one transaction can only have one output referred to adderss
			}
		}
	}
	return accumulated, unspentOuts
}

func (bc *BlockChain) RunMine() {
	transactionPool := CreateTransactionPool()
	//In the near future, we'll have to validate the transactions first here.
	candidateBlock := CreateBlock(bc.LastHash, transactionPool.PubTx) //PoW has been done here.
	if candidateBlock.ValidatePoW() {
		bc.AddBlock(candidateBlock)
		bc.SaveFile()
		err := RemoveTransactionPoolFile()
		utils.Handle(err)
		return
	} else {
		fmt.Println("Block has invalid nonce.")
		return
	}
}

func (bc *BlockChain) SaveFile() {
	var content bytes.Buffer
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(bc)
	utils.Handle(err)
	err = ioutil.WriteFile(constcoe.BlockchainFile, content.Bytes(), 0644)
	utils.Handle(err)
}

func (bc *BlockChain) LoadFile() error {
	if !utils.FileExists(constcoe.BlockchainFile) {
		return nil
	}

	fileContent, err := ioutil.ReadFile(constcoe.BlockchainFile)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(bytes.NewBuffer(fileContent))
	err = decoder.Decode(&bc)

	if err != nil {
		return err
	}

	return nil
}

func RemoveBlockchainFile() error {
	err := os.Remove(constcoe.BlockchainFile)
	return err
}
