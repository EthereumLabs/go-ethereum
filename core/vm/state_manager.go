package vm

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type stateManagerFunction func(*EVM, *Contract, []byte) ([]byte, error)
type methodId [4]byte

var funcs = map[string]stateManagerFunction{
	"getStorage(address,bytes32)":            getStorage,
	"setStorage(address,bytes32,bytes32)":    setStorage,
	"getOvmContractNonce(address)":           getOvmContractNonce,
	"getCodeContractBytecode(address)":       getCodeContractBytecode,
	"getCodeContractHash(address)":           getCodeContractHash,
	"getCodeContractAddress(address)":        getCodeContractAddress,
	"associateCodeContract(address,address)": associateCodeContract,
	"incrementOvmContractNonce(address)":     incrementOvmContractNonce,
}
var methodIds map[[4]byte]stateManagerFunction
var executionMangerBytecode []byte

func init() {
	methodIds = make(map[[4]byte]stateManagerFunction, len(funcs))
	for methodSignature, f := range funcs {
		methodIds[MethodSignatureToMethodId(methodSignature)] = f
	}
}

func MethodSignatureToMethodId(methodSignature string) [4]byte {
	var methodId [4]byte
	copy(methodId[:], crypto.Keccak256([]byte(methodSignature)))
	return methodId
}

func callStateManager(input []byte, evm *EVM, contract *Contract) (ret []byte, err error) {
	var methodId [4]byte
	if len(input) == 0 {
		return nil, nil
	}
	copy(methodId[:], input[:4])
	ret, err = methodIds[methodId](evm, contract, input)
	return ret, err
}

func setStorage(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	address := common.BytesToAddress(input[4:36])
	key := common.BytesToHash(input[36:68])
	val := common.BytesToHash(input[68:100])
	fmt.Println("[State Mgr] Setting storage address:", hex.EncodeToString(address.Bytes()), "key:", hex.EncodeToString(key.Bytes()), "val:", hex.EncodeToString(val.Bytes()))
	evm.StateDB.SetState(address, key, val)
	return nil, nil
}

func getStorage(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	address := common.BytesToAddress(input[4:36])
	key := common.BytesToHash(input[36:68])
	val := evm.StateDB.GetState(address, key)
	fmt.Println("[State Mgr] Getting storage address:", hex.EncodeToString(address.Bytes()), "key:", hex.EncodeToString(key.Bytes()), "val:", hex.EncodeToString(val.Bytes()))
	return val.Bytes(), nil
}

func getCodeContractBytecode(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	address := common.BytesToAddress(input[4:36])
	code := evm.StateDB.GetCode(address)
	fmt.Println("[State Mgr] Getting Bytecode of address:", hex.EncodeToString(address.Bytes()), "Code:", hex.EncodeToString(code))
	return simpleAbiEncode(code), nil
}

func getCodeContractHash(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	address := common.BytesToAddress(input[4:36])
	codeHash := evm.StateDB.GetCodeHash(address)
	fmt.Println("[State Mgr] Getting Code Hash of address:", hex.EncodeToString(address.Bytes()), "Code hash:", hex.EncodeToString(codeHash.Bytes()))
	return codeHash.Bytes(), nil
}

func associateCodeContract(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	fmt.Println("[State Mgr] Associating code contract")
	return []byte{}, nil
}

func getCodeContractAddress(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	address := input[4:36]
	fmt.Println("[State Mgr] Getting code contract address:", hex.EncodeToString(address))
	return address, nil
}

func getOvmContractNonce(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	address := common.BytesToAddress(input[4:36])
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, evm.StateDB.GetNonce(address))
	val := append(make([]byte, 24), b[:]...)
	fmt.Println("[State Mgr] Getting nonce:", hex.EncodeToString(address.Bytes()))
	return val, nil
}

func incrementOvmContractNonce(evm *EVM, contract *Contract, input []byte) (ret []byte, err error) {
	address := common.BytesToAddress(input[4:36])
	oldNonce := evm.StateDB.GetNonce(address)
	evm.StateDB.SetNonce(address, oldNonce+1)
	fmt.Println("[State Mgr] Incrementing nonce:", hex.EncodeToString(address.Bytes()))
	return nil, nil
}

func simpleAbiEncode(bytes []byte) []byte {
	encodedCode := make([]byte, WORD_SIZE)
	binary.BigEndian.PutUint64(encodedCode[WORD_SIZE-8:], uint64(len(bytes)))
	padding := make([]byte, len(bytes)%WORD_SIZE)
	codeWithLength := append(append(encodedCode, bytes...), padding...)
	offset := make([]byte, WORD_SIZE)
	// Hardcode a 2 because we will only return dynamic bytes with a single element
	binary.BigEndian.PutUint64(offset[WORD_SIZE-8:], uint64(2))
	return append([]byte{0, 0}, append(offset, codeWithLength...)...)
}
