package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"go-chain/constcoe"
	"go-chain/utils"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxOutput struct {
	Value     int
	ToAddress []byte
}

type TxInput struct {
	TxID        []byte
	OutIdx      int
	FromAddress []byte
}

func (tx *Transaction) TxHash() []byte {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	utils.Handle(err)

	hash = sha256.Sum256(encoded.Bytes())
	return hash[:]
}

func (tx *Transaction) SetID() {
	tx.ID = tx.TxHash()
}

func BaseTx(toaddress []byte) *Transaction {
	txIn := TxInput{[]byte{}, -1, []byte{}}
	txOut := TxOutput{constcoe.InitCoin, toaddress}
	tx := Transaction{[]byte("This is the Base Transaction!"), []TxInput{txIn}, []TxOutput{txOut}}
	return &tx
}

func (tx *Transaction) IsBase() bool {
	return len(tx.Inputs) == 1 && tx.Inputs[0].OutIdx == -1
}

func (in *TxInput) FromAddressRight(address []byte) bool {
	return bytes.Equal(in.FromAddress, address)
}

func (out *TxOutput) ToAddressRight(address []byte) bool {
	return bytes.Equal(out.ToAddress, address)
}

func CreateTransaction(from, to []byte, amount int) (*Transaction, bool) {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := CreateBlockChain().FindSpendableOutputs(from, amount)
	if acc < amount {
		fmt.Println("Not enough coins!")
		return &Transaction{}, false
	}
	for txid, outidx := range validOutputs {
		txID, err := hex.DecodeString(txid)
		utils.Handle(err)
		input := TxInput{txID, outidx, from}
		inputs = append(inputs, input)
	}

	outputs = append(outputs, TxOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx, true
}
