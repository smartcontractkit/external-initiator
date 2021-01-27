// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package keeper_registry_contract

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// KeeperRegistryContractABI is the input ABI used to generate the binding from.
const KeeperRegistryContractABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"link\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"linkEthFeed\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"fastGasFeed\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"paymentPremiumPPB\",\"type\":\"uint32\"},{\"internalType\":\"uint24\",\"name\":\"blockCountPerTurn\",\"type\":\"uint24\"},{\"internalType\":\"uint32\",\"name\":\"checkGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"uint24\",\"name\":\"stalenessSeconds\",\"type\":\"uint24\"},{\"internalType\":\"int256\",\"name\":\"fallbackGasPrice\",\"type\":\"int256\"},{\"internalType\":\"int256\",\"name\":\"fallbackLinkPrice\",\"type\":\"int256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"paymentPremiumPPB\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint24\",\"name\":\"blockCountPerTurn\",\"type\":\"uint24\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"checkGasLimit\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint24\",\"name\":\"stalenessSeconds\",\"type\":\"uint24\"},{\"indexed\":false,\"internalType\":\"int256\",\"name\":\"fallbackGasPrice\",\"type\":\"int256\"},{\"indexed\":false,\"internalType\":\"int256\",\"name\":\"fallbackLinkPrice\",\"type\":\"int256\"}],\"name\":\"ConfigSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint96\",\"name\":\"amount\",\"type\":\"uint96\"}],\"name\":\"FundsAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"FundsWithdrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"keepers\",\"type\":\"address[]\"},{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"payees\",\"type\":\"address[]\"}],\"name\":\"KeepersUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"OwnershipTransferRequested\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"keeper\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"PayeeshipTransferRequested\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"keeper\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"PayeeshipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"keeper\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"payee\",\"type\":\"address\"}],\"name\":\"PaymentWithdrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"atBlockHeight\",\"type\":\"uint64\"}],\"name\":\"UpkeepCanceled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint96\",\"name\":\"payment\",\"type\":\"uint96\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"performData\",\"type\":\"bytes\"}],\"name\":\"UpkeepPerformed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"executeGas\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"admin\",\"type\":\"address\"}],\"name\":\"UpkeepRegistered\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"FAST_GAS_FEED\",\"outputs\":[{\"internalType\":\"contractAggregatorV3Interface\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"LINK\",\"outputs\":[{\"internalType\":\"contractLinkTokenInterface\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"LINK_ETH_FEED\",\"outputs\":[{\"internalType\":\"contractAggregatorV3Interface\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"keeper\",\"type\":\"address\"}],\"name\":\"acceptPayeeship\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint96\",\"name\":\"amount\",\"type\":\"uint96\"}],\"name\":\"addFunds\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"cancelUpkeep\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"}],\"name\":\"checkUpkeep\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"performData\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"maxLinkPayment\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"int256\",\"name\":\"gasWei\",\"type\":\"int256\"},{\"internalType\":\"int256\",\"name\":\"linkEth\",\"type\":\"int256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCanceledUpkeepList\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getConfig\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"paymentPremiumPPB\",\"type\":\"uint32\"},{\"internalType\":\"uint24\",\"name\":\"blockCountPerTurn\",\"type\":\"uint24\"},{\"internalType\":\"uint32\",\"name\":\"checkGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"uint24\",\"name\":\"stalenessSeconds\",\"type\":\"uint24\"},{\"internalType\":\"int256\",\"name\":\"fallbackGasPrice\",\"type\":\"int256\"},{\"internalType\":\"int256\",\"name\":\"fallbackLinkPrice\",\"type\":\"int256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"query\",\"type\":\"address\"}],\"name\":\"getKeeperInfo\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"payee\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint96\",\"name\":\"balance\",\"type\":\"uint96\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getKeeperList\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"getUpkeep\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"executeGas\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"checkData\",\"type\":\"bytes\"},{\"internalType\":\"uint96\",\"name\":\"balance\",\"type\":\"uint96\"},{\"internalType\":\"address\",\"name\":\"lastKeeper\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"admin\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"maxValidBlocknumber\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getUpkeepCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"onTokenTransfer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"performData\",\"type\":\"bytes\"}],\"name\":\"performUpkeep\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"recoverFunds\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"gasLimit\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"admin\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"checkData\",\"type\":\"bytes\"}],\"name\":\"registerUpkeep\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"paymentPremiumPPB\",\"type\":\"uint32\"},{\"internalType\":\"uint24\",\"name\":\"blockCountPerTurn\",\"type\":\"uint24\"},{\"internalType\":\"uint32\",\"name\":\"checkGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"uint24\",\"name\":\"stalenessSeconds\",\"type\":\"uint24\"},{\"internalType\":\"int256\",\"name\":\"fallbackGasPrice\",\"type\":\"int256\"},{\"internalType\":\"int256\",\"name\":\"fallbackLinkPrice\",\"type\":\"int256\"}],\"name\":\"setConfig\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"keepers\",\"type\":\"address[]\"},{\"internalType\":\"address[]\",\"name\":\"payees\",\"type\":\"address[]\"}],\"name\":\"setKeepers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"keeper\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"proposed\",\"type\":\"address\"}],\"name\":\"transferPayeeship\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"withdrawFunds\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"withdrawPayment\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// KeeperRegistryContract is an auto generated Go binding around an Ethereum contract.
type KeeperRegistryContract struct {
	KeeperRegistryContractCaller     // Read-only binding to the contract
	KeeperRegistryContractTransactor // Write-only binding to the contract
	KeeperRegistryContractFilterer   // Log filterer for contract events
}

// KeeperRegistryContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type KeeperRegistryContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeeperRegistryContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type KeeperRegistryContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeeperRegistryContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type KeeperRegistryContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeeperRegistryContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type KeeperRegistryContractSession struct {
	Contract     *KeeperRegistryContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts           // Call options to use throughout this session
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// KeeperRegistryContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type KeeperRegistryContractCallerSession struct {
	Contract *KeeperRegistryContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                 // Call options to use throughout this session
}

// KeeperRegistryContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type KeeperRegistryContractTransactorSession struct {
	Contract     *KeeperRegistryContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// KeeperRegistryContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type KeeperRegistryContractRaw struct {
	Contract *KeeperRegistryContract // Generic contract binding to access the raw methods on
}

// KeeperRegistryContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type KeeperRegistryContractCallerRaw struct {
	Contract *KeeperRegistryContractCaller // Generic read-only contract binding to access the raw methods on
}

// KeeperRegistryContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type KeeperRegistryContractTransactorRaw struct {
	Contract *KeeperRegistryContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewKeeperRegistryContract creates a new instance of KeeperRegistryContract, bound to a specific deployed contract.
func NewKeeperRegistryContract(address common.Address, backend bind.ContractBackend) (*KeeperRegistryContract, error) {
	contract, err := bindKeeperRegistryContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContract{KeeperRegistryContractCaller: KeeperRegistryContractCaller{contract: contract}, KeeperRegistryContractTransactor: KeeperRegistryContractTransactor{contract: contract}, KeeperRegistryContractFilterer: KeeperRegistryContractFilterer{contract: contract}}, nil
}

// NewKeeperRegistryContractCaller creates a new read-only instance of KeeperRegistryContract, bound to a specific deployed contract.
func NewKeeperRegistryContractCaller(address common.Address, caller bind.ContractCaller) (*KeeperRegistryContractCaller, error) {
	contract, err := bindKeeperRegistryContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractCaller{contract: contract}, nil
}

// NewKeeperRegistryContractTransactor creates a new write-only instance of KeeperRegistryContract, bound to a specific deployed contract.
func NewKeeperRegistryContractTransactor(address common.Address, transactor bind.ContractTransactor) (*KeeperRegistryContractTransactor, error) {
	contract, err := bindKeeperRegistryContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractTransactor{contract: contract}, nil
}

// NewKeeperRegistryContractFilterer creates a new log filterer instance of KeeperRegistryContract, bound to a specific deployed contract.
func NewKeeperRegistryContractFilterer(address common.Address, filterer bind.ContractFilterer) (*KeeperRegistryContractFilterer, error) {
	contract, err := bindKeeperRegistryContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractFilterer{contract: contract}, nil
}

// bindKeeperRegistryContract binds a generic wrapper to an already deployed contract.
func bindKeeperRegistryContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(KeeperRegistryContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_KeeperRegistryContract *KeeperRegistryContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _KeeperRegistryContract.Contract.KeeperRegistryContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_KeeperRegistryContract *KeeperRegistryContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.KeeperRegistryContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_KeeperRegistryContract *KeeperRegistryContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.KeeperRegistryContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_KeeperRegistryContract *KeeperRegistryContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _KeeperRegistryContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_KeeperRegistryContract *KeeperRegistryContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_KeeperRegistryContract *KeeperRegistryContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.contract.Transact(opts, method, params...)
}

// FASTGASFEED is a free data retrieval call binding the contract method 0x4584a419.
//
// Solidity: function FAST_GAS_FEED() view returns(address)
func (_KeeperRegistryContract *KeeperRegistryContractCaller) FASTGASFEED(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _KeeperRegistryContract.contract.Call(opts, &out, "FAST_GAS_FEED")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// FASTGASFEED is a free data retrieval call binding the contract method 0x4584a419.
//
// Solidity: function FAST_GAS_FEED() view returns(address)
func (_KeeperRegistryContract *KeeperRegistryContractSession) FASTGASFEED() (common.Address, error) {
	return _KeeperRegistryContract.Contract.FASTGASFEED(&_KeeperRegistryContract.CallOpts)
}

// FASTGASFEED is a free data retrieval call binding the contract method 0x4584a419.
//
// Solidity: function FAST_GAS_FEED() view returns(address)
func (_KeeperRegistryContract *KeeperRegistryContractCallerSession) FASTGASFEED() (common.Address, error) {
	return _KeeperRegistryContract.Contract.FASTGASFEED(&_KeeperRegistryContract.CallOpts)
}

// LINK is a free data retrieval call binding the contract method 0x1b6b6d23.
//
// Solidity: function LINK() view returns(address)
func (_KeeperRegistryContract *KeeperRegistryContractCaller) LINK(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _KeeperRegistryContract.contract.Call(opts, &out, "LINK")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// LINK is a free data retrieval call binding the contract method 0x1b6b6d23.
//
// Solidity: function LINK() view returns(address)
func (_KeeperRegistryContract *KeeperRegistryContractSession) LINK() (common.Address, error) {
	return _KeeperRegistryContract.Contract.LINK(&_KeeperRegistryContract.CallOpts)
}

// LINK is a free data retrieval call binding the contract method 0x1b6b6d23.
//
// Solidity: function LINK() view returns(address)
func (_KeeperRegistryContract *KeeperRegistryContractCallerSession) LINK() (common.Address, error) {
	return _KeeperRegistryContract.Contract.LINK(&_KeeperRegistryContract.CallOpts)
}

// LINKETHFEED is a free data retrieval call binding the contract method 0xad178361.
//
// Solidity: function LINK_ETH_FEED() view returns(address)
func (_KeeperRegistryContract *KeeperRegistryContractCaller) LINKETHFEED(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _KeeperRegistryContract.contract.Call(opts, &out, "LINK_ETH_FEED")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// LINKETHFEED is a free data retrieval call binding the contract method 0xad178361.
//
// Solidity: function LINK_ETH_FEED() view returns(address)
func (_KeeperRegistryContract *KeeperRegistryContractSession) LINKETHFEED() (common.Address, error) {
	return _KeeperRegistryContract.Contract.LINKETHFEED(&_KeeperRegistryContract.CallOpts)
}

// LINKETHFEED is a free data retrieval call binding the contract method 0xad178361.
//
// Solidity: function LINK_ETH_FEED() view returns(address)
func (_KeeperRegistryContract *KeeperRegistryContractCallerSession) LINKETHFEED() (common.Address, error) {
	return _KeeperRegistryContract.Contract.LINKETHFEED(&_KeeperRegistryContract.CallOpts)
}

// GetCanceledUpkeepList is a free data retrieval call binding the contract method 0x2cb6864d.
//
// Solidity: function getCanceledUpkeepList() view returns(uint256[])
func (_KeeperRegistryContract *KeeperRegistryContractCaller) GetCanceledUpkeepList(opts *bind.CallOpts) ([]*big.Int, error) {
	var out []interface{}
	err := _KeeperRegistryContract.contract.Call(opts, &out, "getCanceledUpkeepList")

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetCanceledUpkeepList is a free data retrieval call binding the contract method 0x2cb6864d.
//
// Solidity: function getCanceledUpkeepList() view returns(uint256[])
func (_KeeperRegistryContract *KeeperRegistryContractSession) GetCanceledUpkeepList() ([]*big.Int, error) {
	return _KeeperRegistryContract.Contract.GetCanceledUpkeepList(&_KeeperRegistryContract.CallOpts)
}

// GetCanceledUpkeepList is a free data retrieval call binding the contract method 0x2cb6864d.
//
// Solidity: function getCanceledUpkeepList() view returns(uint256[])
func (_KeeperRegistryContract *KeeperRegistryContractCallerSession) GetCanceledUpkeepList() ([]*big.Int, error) {
	return _KeeperRegistryContract.Contract.GetCanceledUpkeepList(&_KeeperRegistryContract.CallOpts)
}

// GetConfig is a free data retrieval call binding the contract method 0xc3f909d4.
//
// Solidity: function getConfig() view returns(uint32 paymentPremiumPPB, uint24 blockCountPerTurn, uint32 checkGasLimit, uint24 stalenessSeconds, int256 fallbackGasPrice, int256 fallbackLinkPrice)
func (_KeeperRegistryContract *KeeperRegistryContractCaller) GetConfig(opts *bind.CallOpts) (struct {
	PaymentPremiumPPB uint32
	BlockCountPerTurn *big.Int
	CheckGasLimit     uint32
	StalenessSeconds  *big.Int
	FallbackGasPrice  *big.Int
	FallbackLinkPrice *big.Int
}, error) {
	var out []interface{}
	err := _KeeperRegistryContract.contract.Call(opts, &out, "getConfig")

	outstruct := new(struct {
		PaymentPremiumPPB uint32
		BlockCountPerTurn *big.Int
		CheckGasLimit     uint32
		StalenessSeconds  *big.Int
		FallbackGasPrice  *big.Int
		FallbackLinkPrice *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.PaymentPremiumPPB = out[0].(uint32)
	outstruct.BlockCountPerTurn = out[1].(*big.Int)
	outstruct.CheckGasLimit = out[2].(uint32)
	outstruct.StalenessSeconds = out[3].(*big.Int)
	outstruct.FallbackGasPrice = out[4].(*big.Int)
	outstruct.FallbackLinkPrice = out[5].(*big.Int)

	return *outstruct, err

}

// GetConfig is a free data retrieval call binding the contract method 0xc3f909d4.
//
// Solidity: function getConfig() view returns(uint32 paymentPremiumPPB, uint24 blockCountPerTurn, uint32 checkGasLimit, uint24 stalenessSeconds, int256 fallbackGasPrice, int256 fallbackLinkPrice)
func (_KeeperRegistryContract *KeeperRegistryContractSession) GetConfig() (struct {
	PaymentPremiumPPB uint32
	BlockCountPerTurn *big.Int
	CheckGasLimit     uint32
	StalenessSeconds  *big.Int
	FallbackGasPrice  *big.Int
	FallbackLinkPrice *big.Int
}, error) {
	return _KeeperRegistryContract.Contract.GetConfig(&_KeeperRegistryContract.CallOpts)
}

// GetConfig is a free data retrieval call binding the contract method 0xc3f909d4.
//
// Solidity: function getConfig() view returns(uint32 paymentPremiumPPB, uint24 blockCountPerTurn, uint32 checkGasLimit, uint24 stalenessSeconds, int256 fallbackGasPrice, int256 fallbackLinkPrice)
func (_KeeperRegistryContract *KeeperRegistryContractCallerSession) GetConfig() (struct {
	PaymentPremiumPPB uint32
	BlockCountPerTurn *big.Int
	CheckGasLimit     uint32
	StalenessSeconds  *big.Int
	FallbackGasPrice  *big.Int
	FallbackLinkPrice *big.Int
}, error) {
	return _KeeperRegistryContract.Contract.GetConfig(&_KeeperRegistryContract.CallOpts)
}

// GetKeeperInfo is a free data retrieval call binding the contract method 0x1e12b8a5.
//
// Solidity: function getKeeperInfo(address query) view returns(address payee, bool active, uint96 balance)
func (_KeeperRegistryContract *KeeperRegistryContractCaller) GetKeeperInfo(opts *bind.CallOpts, query common.Address) (struct {
	Payee   common.Address
	Active  bool
	Balance *big.Int
}, error) {
	var out []interface{}
	err := _KeeperRegistryContract.contract.Call(opts, &out, "getKeeperInfo", query)

	outstruct := new(struct {
		Payee   common.Address
		Active  bool
		Balance *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Payee = out[0].(common.Address)
	outstruct.Active = out[1].(bool)
	outstruct.Balance = out[2].(*big.Int)

	return *outstruct, err

}

// GetKeeperInfo is a free data retrieval call binding the contract method 0x1e12b8a5.
//
// Solidity: function getKeeperInfo(address query) view returns(address payee, bool active, uint96 balance)
func (_KeeperRegistryContract *KeeperRegistryContractSession) GetKeeperInfo(query common.Address) (struct {
	Payee   common.Address
	Active  bool
	Balance *big.Int
}, error) {
	return _KeeperRegistryContract.Contract.GetKeeperInfo(&_KeeperRegistryContract.CallOpts, query)
}

// GetKeeperInfo is a free data retrieval call binding the contract method 0x1e12b8a5.
//
// Solidity: function getKeeperInfo(address query) view returns(address payee, bool active, uint96 balance)
func (_KeeperRegistryContract *KeeperRegistryContractCallerSession) GetKeeperInfo(query common.Address) (struct {
	Payee   common.Address
	Active  bool
	Balance *big.Int
}, error) {
	return _KeeperRegistryContract.Contract.GetKeeperInfo(&_KeeperRegistryContract.CallOpts, query)
}

// GetKeeperList is a free data retrieval call binding the contract method 0x15a126ea.
//
// Solidity: function getKeeperList() view returns(address[])
func (_KeeperRegistryContract *KeeperRegistryContractCaller) GetKeeperList(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _KeeperRegistryContract.contract.Call(opts, &out, "getKeeperList")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetKeeperList is a free data retrieval call binding the contract method 0x15a126ea.
//
// Solidity: function getKeeperList() view returns(address[])
func (_KeeperRegistryContract *KeeperRegistryContractSession) GetKeeperList() ([]common.Address, error) {
	return _KeeperRegistryContract.Contract.GetKeeperList(&_KeeperRegistryContract.CallOpts)
}

// GetKeeperList is a free data retrieval call binding the contract method 0x15a126ea.
//
// Solidity: function getKeeperList() view returns(address[])
func (_KeeperRegistryContract *KeeperRegistryContractCallerSession) GetKeeperList() ([]common.Address, error) {
	return _KeeperRegistryContract.Contract.GetKeeperList(&_KeeperRegistryContract.CallOpts)
}

// GetUpkeep is a free data retrieval call binding the contract method 0xc7c3a19a.
//
// Solidity: function getUpkeep(uint256 id) view returns(address target, uint32 executeGas, bytes checkData, uint96 balance, address lastKeeper, address admin, uint64 maxValidBlocknumber)
func (_KeeperRegistryContract *KeeperRegistryContractCaller) GetUpkeep(opts *bind.CallOpts, id *big.Int) (struct {
	Target              common.Address
	ExecuteGas          uint32
	CheckData           []byte
	Balance             *big.Int
	LastKeeper          common.Address
	Admin               common.Address
	MaxValidBlocknumber uint64
}, error) {
	var out []interface{}
	err := _KeeperRegistryContract.contract.Call(opts, &out, "getUpkeep", id)

	outstruct := new(struct {
		Target              common.Address
		ExecuteGas          uint32
		CheckData           []byte
		Balance             *big.Int
		LastKeeper          common.Address
		Admin               common.Address
		MaxValidBlocknumber uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Target = out[0].(common.Address)
	outstruct.ExecuteGas = out[1].(uint32)
	outstruct.CheckData = out[2].([]byte)
	outstruct.Balance = out[3].(*big.Int)
	outstruct.LastKeeper = out[4].(common.Address)
	outstruct.Admin = out[5].(common.Address)
	outstruct.MaxValidBlocknumber = out[6].(uint64)

	return *outstruct, err

}

// GetUpkeep is a free data retrieval call binding the contract method 0xc7c3a19a.
//
// Solidity: function getUpkeep(uint256 id) view returns(address target, uint32 executeGas, bytes checkData, uint96 balance, address lastKeeper, address admin, uint64 maxValidBlocknumber)
func (_KeeperRegistryContract *KeeperRegistryContractSession) GetUpkeep(id *big.Int) (struct {
	Target              common.Address
	ExecuteGas          uint32
	CheckData           []byte
	Balance             *big.Int
	LastKeeper          common.Address
	Admin               common.Address
	MaxValidBlocknumber uint64
}, error) {
	return _KeeperRegistryContract.Contract.GetUpkeep(&_KeeperRegistryContract.CallOpts, id)
}

// GetUpkeep is a free data retrieval call binding the contract method 0xc7c3a19a.
//
// Solidity: function getUpkeep(uint256 id) view returns(address target, uint32 executeGas, bytes checkData, uint96 balance, address lastKeeper, address admin, uint64 maxValidBlocknumber)
func (_KeeperRegistryContract *KeeperRegistryContractCallerSession) GetUpkeep(id *big.Int) (struct {
	Target              common.Address
	ExecuteGas          uint32
	CheckData           []byte
	Balance             *big.Int
	LastKeeper          common.Address
	Admin               common.Address
	MaxValidBlocknumber uint64
}, error) {
	return _KeeperRegistryContract.Contract.GetUpkeep(&_KeeperRegistryContract.CallOpts, id)
}

// GetUpkeepCount is a free data retrieval call binding the contract method 0xfecf27c9.
//
// Solidity: function getUpkeepCount() view returns(uint256)
func (_KeeperRegistryContract *KeeperRegistryContractCaller) GetUpkeepCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _KeeperRegistryContract.contract.Call(opts, &out, "getUpkeepCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUpkeepCount is a free data retrieval call binding the contract method 0xfecf27c9.
//
// Solidity: function getUpkeepCount() view returns(uint256)
func (_KeeperRegistryContract *KeeperRegistryContractSession) GetUpkeepCount() (*big.Int, error) {
	return _KeeperRegistryContract.Contract.GetUpkeepCount(&_KeeperRegistryContract.CallOpts)
}

// GetUpkeepCount is a free data retrieval call binding the contract method 0xfecf27c9.
//
// Solidity: function getUpkeepCount() view returns(uint256)
func (_KeeperRegistryContract *KeeperRegistryContractCallerSession) GetUpkeepCount() (*big.Int, error) {
	return _KeeperRegistryContract.Contract.GetUpkeepCount(&_KeeperRegistryContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_KeeperRegistryContract *KeeperRegistryContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _KeeperRegistryContract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_KeeperRegistryContract *KeeperRegistryContractSession) Owner() (common.Address, error) {
	return _KeeperRegistryContract.Contract.Owner(&_KeeperRegistryContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_KeeperRegistryContract *KeeperRegistryContractCallerSession) Owner() (common.Address, error) {
	return _KeeperRegistryContract.Contract.Owner(&_KeeperRegistryContract.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_KeeperRegistryContract *KeeperRegistryContractSession) AcceptOwnership() (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.AcceptOwnership(&_KeeperRegistryContract.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.AcceptOwnership(&_KeeperRegistryContract.TransactOpts)
}

// AcceptPayeeship is a paid mutator transaction binding the contract method 0xb121e147.
//
// Solidity: function acceptPayeeship(address keeper) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) AcceptPayeeship(opts *bind.TransactOpts, keeper common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "acceptPayeeship", keeper)
}

// AcceptPayeeship is a paid mutator transaction binding the contract method 0xb121e147.
//
// Solidity: function acceptPayeeship(address keeper) returns()
func (_KeeperRegistryContract *KeeperRegistryContractSession) AcceptPayeeship(keeper common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.AcceptPayeeship(&_KeeperRegistryContract.TransactOpts, keeper)
}

// AcceptPayeeship is a paid mutator transaction binding the contract method 0xb121e147.
//
// Solidity: function acceptPayeeship(address keeper) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) AcceptPayeeship(keeper common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.AcceptPayeeship(&_KeeperRegistryContract.TransactOpts, keeper)
}

// AddFunds is a paid mutator transaction binding the contract method 0x948108f7.
//
// Solidity: function addFunds(uint256 id, uint96 amount) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) AddFunds(opts *bind.TransactOpts, id *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "addFunds", id, amount)
}

// AddFunds is a paid mutator transaction binding the contract method 0x948108f7.
//
// Solidity: function addFunds(uint256 id, uint96 amount) returns()
func (_KeeperRegistryContract *KeeperRegistryContractSession) AddFunds(id *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.AddFunds(&_KeeperRegistryContract.TransactOpts, id, amount)
}

// AddFunds is a paid mutator transaction binding the contract method 0x948108f7.
//
// Solidity: function addFunds(uint256 id, uint96 amount) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) AddFunds(id *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.AddFunds(&_KeeperRegistryContract.TransactOpts, id, amount)
}

// CancelUpkeep is a paid mutator transaction binding the contract method 0xc8048022.
//
// Solidity: function cancelUpkeep(uint256 id) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) CancelUpkeep(opts *bind.TransactOpts, id *big.Int) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "cancelUpkeep", id)
}

// CancelUpkeep is a paid mutator transaction binding the contract method 0xc8048022.
//
// Solidity: function cancelUpkeep(uint256 id) returns()
func (_KeeperRegistryContract *KeeperRegistryContractSession) CancelUpkeep(id *big.Int) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.CancelUpkeep(&_KeeperRegistryContract.TransactOpts, id)
}

// CancelUpkeep is a paid mutator transaction binding the contract method 0xc8048022.
//
// Solidity: function cancelUpkeep(uint256 id) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) CancelUpkeep(id *big.Int) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.CancelUpkeep(&_KeeperRegistryContract.TransactOpts, id)
}

// CheckUpkeep is a paid mutator transaction binding the contract method 0xc41b813a.
//
// Solidity: function checkUpkeep(uint256 id, address from) returns(bytes performData, uint256 maxLinkPayment, uint256 gasLimit, int256 gasWei, int256 linkEth)
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) CheckUpkeep(opts *bind.TransactOpts, id *big.Int, from common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "checkUpkeep", id, from)
}

// CheckUpkeep is a paid mutator transaction binding the contract method 0xc41b813a.
//
// Solidity: function checkUpkeep(uint256 id, address from) returns(bytes performData, uint256 maxLinkPayment, uint256 gasLimit, int256 gasWei, int256 linkEth)
func (_KeeperRegistryContract *KeeperRegistryContractSession) CheckUpkeep(id *big.Int, from common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.CheckUpkeep(&_KeeperRegistryContract.TransactOpts, id, from)
}

// CheckUpkeep is a paid mutator transaction binding the contract method 0xc41b813a.
//
// Solidity: function checkUpkeep(uint256 id, address from) returns(bytes performData, uint256 maxLinkPayment, uint256 gasLimit, int256 gasWei, int256 linkEth)
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) CheckUpkeep(id *big.Int, from common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.CheckUpkeep(&_KeeperRegistryContract.TransactOpts, id, from)
}

// OnTokenTransfer is a paid mutator transaction binding the contract method 0xa4c0ed36.
//
// Solidity: function onTokenTransfer(address sender, uint256 amount, bytes data) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) OnTokenTransfer(opts *bind.TransactOpts, sender common.Address, amount *big.Int, data []byte) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "onTokenTransfer", sender, amount, data)
}

// OnTokenTransfer is a paid mutator transaction binding the contract method 0xa4c0ed36.
//
// Solidity: function onTokenTransfer(address sender, uint256 amount, bytes data) returns()
func (_KeeperRegistryContract *KeeperRegistryContractSession) OnTokenTransfer(sender common.Address, amount *big.Int, data []byte) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.OnTokenTransfer(&_KeeperRegistryContract.TransactOpts, sender, amount, data)
}

// OnTokenTransfer is a paid mutator transaction binding the contract method 0xa4c0ed36.
//
// Solidity: function onTokenTransfer(address sender, uint256 amount, bytes data) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) OnTokenTransfer(sender common.Address, amount *big.Int, data []byte) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.OnTokenTransfer(&_KeeperRegistryContract.TransactOpts, sender, amount, data)
}

// PerformUpkeep is a paid mutator transaction binding the contract method 0x7bbaf1ea.
//
// Solidity: function performUpkeep(uint256 id, bytes performData) returns(bool success)
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) PerformUpkeep(opts *bind.TransactOpts, id *big.Int, performData []byte) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "performUpkeep", id, performData)
}

// PerformUpkeep is a paid mutator transaction binding the contract method 0x7bbaf1ea.
//
// Solidity: function performUpkeep(uint256 id, bytes performData) returns(bool success)
func (_KeeperRegistryContract *KeeperRegistryContractSession) PerformUpkeep(id *big.Int, performData []byte) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.PerformUpkeep(&_KeeperRegistryContract.TransactOpts, id, performData)
}

// PerformUpkeep is a paid mutator transaction binding the contract method 0x7bbaf1ea.
//
// Solidity: function performUpkeep(uint256 id, bytes performData) returns(bool success)
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) PerformUpkeep(id *big.Int, performData []byte) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.PerformUpkeep(&_KeeperRegistryContract.TransactOpts, id, performData)
}

// RecoverFunds is a paid mutator transaction binding the contract method 0xb79550be.
//
// Solidity: function recoverFunds() returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) RecoverFunds(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "recoverFunds")
}

// RecoverFunds is a paid mutator transaction binding the contract method 0xb79550be.
//
// Solidity: function recoverFunds() returns()
func (_KeeperRegistryContract *KeeperRegistryContractSession) RecoverFunds() (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.RecoverFunds(&_KeeperRegistryContract.TransactOpts)
}

// RecoverFunds is a paid mutator transaction binding the contract method 0xb79550be.
//
// Solidity: function recoverFunds() returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) RecoverFunds() (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.RecoverFunds(&_KeeperRegistryContract.TransactOpts)
}

// RegisterUpkeep is a paid mutator transaction binding the contract method 0xda5c6741.
//
// Solidity: function registerUpkeep(address target, uint32 gasLimit, address admin, bytes checkData) returns(uint256 id)
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) RegisterUpkeep(opts *bind.TransactOpts, target common.Address, gasLimit uint32, admin common.Address, checkData []byte) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "registerUpkeep", target, gasLimit, admin, checkData)
}

// RegisterUpkeep is a paid mutator transaction binding the contract method 0xda5c6741.
//
// Solidity: function registerUpkeep(address target, uint32 gasLimit, address admin, bytes checkData) returns(uint256 id)
func (_KeeperRegistryContract *KeeperRegistryContractSession) RegisterUpkeep(target common.Address, gasLimit uint32, admin common.Address, checkData []byte) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.RegisterUpkeep(&_KeeperRegistryContract.TransactOpts, target, gasLimit, admin, checkData)
}

// RegisterUpkeep is a paid mutator transaction binding the contract method 0xda5c6741.
//
// Solidity: function registerUpkeep(address target, uint32 gasLimit, address admin, bytes checkData) returns(uint256 id)
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) RegisterUpkeep(target common.Address, gasLimit uint32, admin common.Address, checkData []byte) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.RegisterUpkeep(&_KeeperRegistryContract.TransactOpts, target, gasLimit, admin, checkData)
}

// SetConfig is a paid mutator transaction binding the contract method 0x10833057.
//
// Solidity: function setConfig(uint32 paymentPremiumPPB, uint24 blockCountPerTurn, uint32 checkGasLimit, uint24 stalenessSeconds, int256 fallbackGasPrice, int256 fallbackLinkPrice) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) SetConfig(opts *bind.TransactOpts, paymentPremiumPPB uint32, blockCountPerTurn *big.Int, checkGasLimit uint32, stalenessSeconds *big.Int, fallbackGasPrice *big.Int, fallbackLinkPrice *big.Int) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "setConfig", paymentPremiumPPB, blockCountPerTurn, checkGasLimit, stalenessSeconds, fallbackGasPrice, fallbackLinkPrice)
}

// SetConfig is a paid mutator transaction binding the contract method 0x10833057.
//
// Solidity: function setConfig(uint32 paymentPremiumPPB, uint24 blockCountPerTurn, uint32 checkGasLimit, uint24 stalenessSeconds, int256 fallbackGasPrice, int256 fallbackLinkPrice) returns()
func (_KeeperRegistryContract *KeeperRegistryContractSession) SetConfig(paymentPremiumPPB uint32, blockCountPerTurn *big.Int, checkGasLimit uint32, stalenessSeconds *big.Int, fallbackGasPrice *big.Int, fallbackLinkPrice *big.Int) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.SetConfig(&_KeeperRegistryContract.TransactOpts, paymentPremiumPPB, blockCountPerTurn, checkGasLimit, stalenessSeconds, fallbackGasPrice, fallbackLinkPrice)
}

// SetConfig is a paid mutator transaction binding the contract method 0x10833057.
//
// Solidity: function setConfig(uint32 paymentPremiumPPB, uint24 blockCountPerTurn, uint32 checkGasLimit, uint24 stalenessSeconds, int256 fallbackGasPrice, int256 fallbackLinkPrice) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) SetConfig(paymentPremiumPPB uint32, blockCountPerTurn *big.Int, checkGasLimit uint32, stalenessSeconds *big.Int, fallbackGasPrice *big.Int, fallbackLinkPrice *big.Int) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.SetConfig(&_KeeperRegistryContract.TransactOpts, paymentPremiumPPB, blockCountPerTurn, checkGasLimit, stalenessSeconds, fallbackGasPrice, fallbackLinkPrice)
}

// SetKeepers is a paid mutator transaction binding the contract method 0xb7fdb436.
//
// Solidity: function setKeepers(address[] keepers, address[] payees) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) SetKeepers(opts *bind.TransactOpts, keepers []common.Address, payees []common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "setKeepers", keepers, payees)
}

// SetKeepers is a paid mutator transaction binding the contract method 0xb7fdb436.
//
// Solidity: function setKeepers(address[] keepers, address[] payees) returns()
func (_KeeperRegistryContract *KeeperRegistryContractSession) SetKeepers(keepers []common.Address, payees []common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.SetKeepers(&_KeeperRegistryContract.TransactOpts, keepers, payees)
}

// SetKeepers is a paid mutator transaction binding the contract method 0xb7fdb436.
//
// Solidity: function setKeepers(address[] keepers, address[] payees) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) SetKeepers(keepers []common.Address, payees []common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.SetKeepers(&_KeeperRegistryContract.TransactOpts, keepers, payees)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _to) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) TransferOwnership(opts *bind.TransactOpts, _to common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "transferOwnership", _to)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _to) returns()
func (_KeeperRegistryContract *KeeperRegistryContractSession) TransferOwnership(_to common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.TransferOwnership(&_KeeperRegistryContract.TransactOpts, _to)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _to) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) TransferOwnership(_to common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.TransferOwnership(&_KeeperRegistryContract.TransactOpts, _to)
}

// TransferPayeeship is a paid mutator transaction binding the contract method 0xeb5dcd6c.
//
// Solidity: function transferPayeeship(address keeper, address proposed) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) TransferPayeeship(opts *bind.TransactOpts, keeper common.Address, proposed common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "transferPayeeship", keeper, proposed)
}

// TransferPayeeship is a paid mutator transaction binding the contract method 0xeb5dcd6c.
//
// Solidity: function transferPayeeship(address keeper, address proposed) returns()
func (_KeeperRegistryContract *KeeperRegistryContractSession) TransferPayeeship(keeper common.Address, proposed common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.TransferPayeeship(&_KeeperRegistryContract.TransactOpts, keeper, proposed)
}

// TransferPayeeship is a paid mutator transaction binding the contract method 0xeb5dcd6c.
//
// Solidity: function transferPayeeship(address keeper, address proposed) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) TransferPayeeship(keeper common.Address, proposed common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.TransferPayeeship(&_KeeperRegistryContract.TransactOpts, keeper, proposed)
}

// WithdrawFunds is a paid mutator transaction binding the contract method 0x744bfe61.
//
// Solidity: function withdrawFunds(uint256 id, address to) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) WithdrawFunds(opts *bind.TransactOpts, id *big.Int, to common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "withdrawFunds", id, to)
}

// WithdrawFunds is a paid mutator transaction binding the contract method 0x744bfe61.
//
// Solidity: function withdrawFunds(uint256 id, address to) returns()
func (_KeeperRegistryContract *KeeperRegistryContractSession) WithdrawFunds(id *big.Int, to common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.WithdrawFunds(&_KeeperRegistryContract.TransactOpts, id, to)
}

// WithdrawFunds is a paid mutator transaction binding the contract method 0x744bfe61.
//
// Solidity: function withdrawFunds(uint256 id, address to) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) WithdrawFunds(id *big.Int, to common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.WithdrawFunds(&_KeeperRegistryContract.TransactOpts, id, to)
}

// WithdrawPayment is a paid mutator transaction binding the contract method 0xa710b221.
//
// Solidity: function withdrawPayment(address from, address to) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactor) WithdrawPayment(opts *bind.TransactOpts, from common.Address, to common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.contract.Transact(opts, "withdrawPayment", from, to)
}

// WithdrawPayment is a paid mutator transaction binding the contract method 0xa710b221.
//
// Solidity: function withdrawPayment(address from, address to) returns()
func (_KeeperRegistryContract *KeeperRegistryContractSession) WithdrawPayment(from common.Address, to common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.WithdrawPayment(&_KeeperRegistryContract.TransactOpts, from, to)
}

// WithdrawPayment is a paid mutator transaction binding the contract method 0xa710b221.
//
// Solidity: function withdrawPayment(address from, address to) returns()
func (_KeeperRegistryContract *KeeperRegistryContractTransactorSession) WithdrawPayment(from common.Address, to common.Address) (*types.Transaction, error) {
	return _KeeperRegistryContract.Contract.WithdrawPayment(&_KeeperRegistryContract.TransactOpts, from, to)
}

// KeeperRegistryContractConfigSetIterator is returned from FilterConfigSet and is used to iterate over the raw logs and unpacked data for ConfigSet events raised by the KeeperRegistryContract contract.
type KeeperRegistryContractConfigSetIterator struct {
	Event *KeeperRegistryContractConfigSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeeperRegistryContractConfigSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperRegistryContractConfigSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeeperRegistryContractConfigSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeeperRegistryContractConfigSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperRegistryContractConfigSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperRegistryContractConfigSet represents a ConfigSet event raised by the KeeperRegistryContract contract.
type KeeperRegistryContractConfigSet struct {
	PaymentPremiumPPB uint32
	BlockCountPerTurn *big.Int
	CheckGasLimit     uint32
	StalenessSeconds  *big.Int
	FallbackGasPrice  *big.Int
	FallbackLinkPrice *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterConfigSet is a free log retrieval operation binding the contract event 0xf68c98604b3cbc1a909e0df75315ce475fbc89bae7a6ab88b53573897f58d0d7.
//
// Solidity: event ConfigSet(uint32 paymentPremiumPPB, uint24 blockCountPerTurn, uint32 checkGasLimit, uint24 stalenessSeconds, int256 fallbackGasPrice, int256 fallbackLinkPrice)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) FilterConfigSet(opts *bind.FilterOpts) (*KeeperRegistryContractConfigSetIterator, error) {

	logs, sub, err := _KeeperRegistryContract.contract.FilterLogs(opts, "ConfigSet")
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractConfigSetIterator{contract: _KeeperRegistryContract.contract, event: "ConfigSet", logs: logs, sub: sub}, nil
}

// WatchConfigSet is a free log subscription operation binding the contract event 0xf68c98604b3cbc1a909e0df75315ce475fbc89bae7a6ab88b53573897f58d0d7.
//
// Solidity: event ConfigSet(uint32 paymentPremiumPPB, uint24 blockCountPerTurn, uint32 checkGasLimit, uint24 stalenessSeconds, int256 fallbackGasPrice, int256 fallbackLinkPrice)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) WatchConfigSet(opts *bind.WatchOpts, sink chan<- *KeeperRegistryContractConfigSet) (event.Subscription, error) {

	logs, sub, err := _KeeperRegistryContract.contract.WatchLogs(opts, "ConfigSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperRegistryContractConfigSet)
				if err := _KeeperRegistryContract.contract.UnpackLog(event, "ConfigSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseConfigSet is a log parse operation binding the contract event 0xf68c98604b3cbc1a909e0df75315ce475fbc89bae7a6ab88b53573897f58d0d7.
//
// Solidity: event ConfigSet(uint32 paymentPremiumPPB, uint24 blockCountPerTurn, uint32 checkGasLimit, uint24 stalenessSeconds, int256 fallbackGasPrice, int256 fallbackLinkPrice)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) ParseConfigSet(log types.Log) (*KeeperRegistryContractConfigSet, error) {
	event := new(KeeperRegistryContractConfigSet)
	if err := _KeeperRegistryContract.contract.UnpackLog(event, "ConfigSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeeperRegistryContractFundsAddedIterator is returned from FilterFundsAdded and is used to iterate over the raw logs and unpacked data for FundsAdded events raised by the KeeperRegistryContract contract.
type KeeperRegistryContractFundsAddedIterator struct {
	Event *KeeperRegistryContractFundsAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeeperRegistryContractFundsAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperRegistryContractFundsAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeeperRegistryContractFundsAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeeperRegistryContractFundsAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperRegistryContractFundsAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperRegistryContractFundsAdded represents a FundsAdded event raised by the KeeperRegistryContract contract.
type KeeperRegistryContractFundsAdded struct {
	Id     *big.Int
	From   common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterFundsAdded is a free log retrieval operation binding the contract event 0xafd24114486da8ebfc32f3626dada8863652e187461aa74d4bfa734891506203.
//
// Solidity: event FundsAdded(uint256 indexed id, address indexed from, uint96 amount)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) FilterFundsAdded(opts *bind.FilterOpts, id []*big.Int, from []common.Address) (*KeeperRegistryContractFundsAddedIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.FilterLogs(opts, "FundsAdded", idRule, fromRule)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractFundsAddedIterator{contract: _KeeperRegistryContract.contract, event: "FundsAdded", logs: logs, sub: sub}, nil
}

// WatchFundsAdded is a free log subscription operation binding the contract event 0xafd24114486da8ebfc32f3626dada8863652e187461aa74d4bfa734891506203.
//
// Solidity: event FundsAdded(uint256 indexed id, address indexed from, uint96 amount)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) WatchFundsAdded(opts *bind.WatchOpts, sink chan<- *KeeperRegistryContractFundsAdded, id []*big.Int, from []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.WatchLogs(opts, "FundsAdded", idRule, fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperRegistryContractFundsAdded)
				if err := _KeeperRegistryContract.contract.UnpackLog(event, "FundsAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFundsAdded is a log parse operation binding the contract event 0xafd24114486da8ebfc32f3626dada8863652e187461aa74d4bfa734891506203.
//
// Solidity: event FundsAdded(uint256 indexed id, address indexed from, uint96 amount)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) ParseFundsAdded(log types.Log) (*KeeperRegistryContractFundsAdded, error) {
	event := new(KeeperRegistryContractFundsAdded)
	if err := _KeeperRegistryContract.contract.UnpackLog(event, "FundsAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeeperRegistryContractFundsWithdrawnIterator is returned from FilterFundsWithdrawn and is used to iterate over the raw logs and unpacked data for FundsWithdrawn events raised by the KeeperRegistryContract contract.
type KeeperRegistryContractFundsWithdrawnIterator struct {
	Event *KeeperRegistryContractFundsWithdrawn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeeperRegistryContractFundsWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperRegistryContractFundsWithdrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeeperRegistryContractFundsWithdrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeeperRegistryContractFundsWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperRegistryContractFundsWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperRegistryContractFundsWithdrawn represents a FundsWithdrawn event raised by the KeeperRegistryContract contract.
type KeeperRegistryContractFundsWithdrawn struct {
	Id     *big.Int
	Amount *big.Int
	To     common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterFundsWithdrawn is a free log retrieval operation binding the contract event 0xf3b5906e5672f3e524854103bcafbbdba80dbdfeca2c35e116127b1060a68318.
//
// Solidity: event FundsWithdrawn(uint256 indexed id, uint256 amount, address to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) FilterFundsWithdrawn(opts *bind.FilterOpts, id []*big.Int) (*KeeperRegistryContractFundsWithdrawnIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.FilterLogs(opts, "FundsWithdrawn", idRule)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractFundsWithdrawnIterator{contract: _KeeperRegistryContract.contract, event: "FundsWithdrawn", logs: logs, sub: sub}, nil
}

// WatchFundsWithdrawn is a free log subscription operation binding the contract event 0xf3b5906e5672f3e524854103bcafbbdba80dbdfeca2c35e116127b1060a68318.
//
// Solidity: event FundsWithdrawn(uint256 indexed id, uint256 amount, address to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) WatchFundsWithdrawn(opts *bind.WatchOpts, sink chan<- *KeeperRegistryContractFundsWithdrawn, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.WatchLogs(opts, "FundsWithdrawn", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperRegistryContractFundsWithdrawn)
				if err := _KeeperRegistryContract.contract.UnpackLog(event, "FundsWithdrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFundsWithdrawn is a log parse operation binding the contract event 0xf3b5906e5672f3e524854103bcafbbdba80dbdfeca2c35e116127b1060a68318.
//
// Solidity: event FundsWithdrawn(uint256 indexed id, uint256 amount, address to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) ParseFundsWithdrawn(log types.Log) (*KeeperRegistryContractFundsWithdrawn, error) {
	event := new(KeeperRegistryContractFundsWithdrawn)
	if err := _KeeperRegistryContract.contract.UnpackLog(event, "FundsWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeeperRegistryContractKeepersUpdatedIterator is returned from FilterKeepersUpdated and is used to iterate over the raw logs and unpacked data for KeepersUpdated events raised by the KeeperRegistryContract contract.
type KeeperRegistryContractKeepersUpdatedIterator struct {
	Event *KeeperRegistryContractKeepersUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeeperRegistryContractKeepersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperRegistryContractKeepersUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeeperRegistryContractKeepersUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeeperRegistryContractKeepersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperRegistryContractKeepersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperRegistryContractKeepersUpdated represents a KeepersUpdated event raised by the KeeperRegistryContract contract.
type KeeperRegistryContractKeepersUpdated struct {
	Keepers []common.Address
	Payees  []common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterKeepersUpdated is a free log retrieval operation binding the contract event 0x056264c94f28bb06c99d13f0446eb96c67c215d8d707bce2655a98ddf1c0b71f.
//
// Solidity: event KeepersUpdated(address[] keepers, address[] payees)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) FilterKeepersUpdated(opts *bind.FilterOpts) (*KeeperRegistryContractKeepersUpdatedIterator, error) {

	logs, sub, err := _KeeperRegistryContract.contract.FilterLogs(opts, "KeepersUpdated")
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractKeepersUpdatedIterator{contract: _KeeperRegistryContract.contract, event: "KeepersUpdated", logs: logs, sub: sub}, nil
}

// WatchKeepersUpdated is a free log subscription operation binding the contract event 0x056264c94f28bb06c99d13f0446eb96c67c215d8d707bce2655a98ddf1c0b71f.
//
// Solidity: event KeepersUpdated(address[] keepers, address[] payees)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) WatchKeepersUpdated(opts *bind.WatchOpts, sink chan<- *KeeperRegistryContractKeepersUpdated) (event.Subscription, error) {

	logs, sub, err := _KeeperRegistryContract.contract.WatchLogs(opts, "KeepersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperRegistryContractKeepersUpdated)
				if err := _KeeperRegistryContract.contract.UnpackLog(event, "KeepersUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseKeepersUpdated is a log parse operation binding the contract event 0x056264c94f28bb06c99d13f0446eb96c67c215d8d707bce2655a98ddf1c0b71f.
//
// Solidity: event KeepersUpdated(address[] keepers, address[] payees)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) ParseKeepersUpdated(log types.Log) (*KeeperRegistryContractKeepersUpdated, error) {
	event := new(KeeperRegistryContractKeepersUpdated)
	if err := _KeeperRegistryContract.contract.UnpackLog(event, "KeepersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeeperRegistryContractOwnershipTransferRequestedIterator is returned from FilterOwnershipTransferRequested and is used to iterate over the raw logs and unpacked data for OwnershipTransferRequested events raised by the KeeperRegistryContract contract.
type KeeperRegistryContractOwnershipTransferRequestedIterator struct {
	Event *KeeperRegistryContractOwnershipTransferRequested // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeeperRegistryContractOwnershipTransferRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperRegistryContractOwnershipTransferRequested)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeeperRegistryContractOwnershipTransferRequested)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeeperRegistryContractOwnershipTransferRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperRegistryContractOwnershipTransferRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperRegistryContractOwnershipTransferRequested represents a OwnershipTransferRequested event raised by the KeeperRegistryContract contract.
type KeeperRegistryContractOwnershipTransferRequested struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferRequested is a free log retrieval operation binding the contract event 0xed8889f560326eb138920d842192f0eb3dd22b4f139c87a2c57538e05bae1278.
//
// Solidity: event OwnershipTransferRequested(address indexed from, address indexed to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) FilterOwnershipTransferRequested(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*KeeperRegistryContractOwnershipTransferRequestedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.FilterLogs(opts, "OwnershipTransferRequested", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractOwnershipTransferRequestedIterator{contract: _KeeperRegistryContract.contract, event: "OwnershipTransferRequested", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferRequested is a free log subscription operation binding the contract event 0xed8889f560326eb138920d842192f0eb3dd22b4f139c87a2c57538e05bae1278.
//
// Solidity: event OwnershipTransferRequested(address indexed from, address indexed to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) WatchOwnershipTransferRequested(opts *bind.WatchOpts, sink chan<- *KeeperRegistryContractOwnershipTransferRequested, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.WatchLogs(opts, "OwnershipTransferRequested", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperRegistryContractOwnershipTransferRequested)
				if err := _KeeperRegistryContract.contract.UnpackLog(event, "OwnershipTransferRequested", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferRequested is a log parse operation binding the contract event 0xed8889f560326eb138920d842192f0eb3dd22b4f139c87a2c57538e05bae1278.
//
// Solidity: event OwnershipTransferRequested(address indexed from, address indexed to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) ParseOwnershipTransferRequested(log types.Log) (*KeeperRegistryContractOwnershipTransferRequested, error) {
	event := new(KeeperRegistryContractOwnershipTransferRequested)
	if err := _KeeperRegistryContract.contract.UnpackLog(event, "OwnershipTransferRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeeperRegistryContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the KeeperRegistryContract contract.
type KeeperRegistryContractOwnershipTransferredIterator struct {
	Event *KeeperRegistryContractOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeeperRegistryContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperRegistryContractOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeeperRegistryContractOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeeperRegistryContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperRegistryContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperRegistryContractOwnershipTransferred represents a OwnershipTransferred event raised by the KeeperRegistryContract contract.
type KeeperRegistryContractOwnershipTransferred struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed from, address indexed to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*KeeperRegistryContractOwnershipTransferredIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.FilterLogs(opts, "OwnershipTransferred", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractOwnershipTransferredIterator{contract: _KeeperRegistryContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed from, address indexed to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *KeeperRegistryContractOwnershipTransferred, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.WatchLogs(opts, "OwnershipTransferred", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperRegistryContractOwnershipTransferred)
				if err := _KeeperRegistryContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed from, address indexed to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) ParseOwnershipTransferred(log types.Log) (*KeeperRegistryContractOwnershipTransferred, error) {
	event := new(KeeperRegistryContractOwnershipTransferred)
	if err := _KeeperRegistryContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeeperRegistryContractPayeeshipTransferRequestedIterator is returned from FilterPayeeshipTransferRequested and is used to iterate over the raw logs and unpacked data for PayeeshipTransferRequested events raised by the KeeperRegistryContract contract.
type KeeperRegistryContractPayeeshipTransferRequestedIterator struct {
	Event *KeeperRegistryContractPayeeshipTransferRequested // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeeperRegistryContractPayeeshipTransferRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperRegistryContractPayeeshipTransferRequested)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeeperRegistryContractPayeeshipTransferRequested)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeeperRegistryContractPayeeshipTransferRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperRegistryContractPayeeshipTransferRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperRegistryContractPayeeshipTransferRequested represents a PayeeshipTransferRequested event raised by the KeeperRegistryContract contract.
type KeeperRegistryContractPayeeshipTransferRequested struct {
	Keeper common.Address
	From   common.Address
	To     common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterPayeeshipTransferRequested is a free log retrieval operation binding the contract event 0x84f7c7c80bb8ed2279b4aab5f61cd05e6374073d38f46d7f32de8c30e9e38367.
//
// Solidity: event PayeeshipTransferRequested(address indexed keeper, address indexed from, address indexed to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) FilterPayeeshipTransferRequested(opts *bind.FilterOpts, keeper []common.Address, from []common.Address, to []common.Address) (*KeeperRegistryContractPayeeshipTransferRequestedIterator, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.FilterLogs(opts, "PayeeshipTransferRequested", keeperRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractPayeeshipTransferRequestedIterator{contract: _KeeperRegistryContract.contract, event: "PayeeshipTransferRequested", logs: logs, sub: sub}, nil
}

// WatchPayeeshipTransferRequested is a free log subscription operation binding the contract event 0x84f7c7c80bb8ed2279b4aab5f61cd05e6374073d38f46d7f32de8c30e9e38367.
//
// Solidity: event PayeeshipTransferRequested(address indexed keeper, address indexed from, address indexed to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) WatchPayeeshipTransferRequested(opts *bind.WatchOpts, sink chan<- *KeeperRegistryContractPayeeshipTransferRequested, keeper []common.Address, from []common.Address, to []common.Address) (event.Subscription, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.WatchLogs(opts, "PayeeshipTransferRequested", keeperRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperRegistryContractPayeeshipTransferRequested)
				if err := _KeeperRegistryContract.contract.UnpackLog(event, "PayeeshipTransferRequested", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePayeeshipTransferRequested is a log parse operation binding the contract event 0x84f7c7c80bb8ed2279b4aab5f61cd05e6374073d38f46d7f32de8c30e9e38367.
//
// Solidity: event PayeeshipTransferRequested(address indexed keeper, address indexed from, address indexed to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) ParsePayeeshipTransferRequested(log types.Log) (*KeeperRegistryContractPayeeshipTransferRequested, error) {
	event := new(KeeperRegistryContractPayeeshipTransferRequested)
	if err := _KeeperRegistryContract.contract.UnpackLog(event, "PayeeshipTransferRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeeperRegistryContractPayeeshipTransferredIterator is returned from FilterPayeeshipTransferred and is used to iterate over the raw logs and unpacked data for PayeeshipTransferred events raised by the KeeperRegistryContract contract.
type KeeperRegistryContractPayeeshipTransferredIterator struct {
	Event *KeeperRegistryContractPayeeshipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeeperRegistryContractPayeeshipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperRegistryContractPayeeshipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeeperRegistryContractPayeeshipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeeperRegistryContractPayeeshipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperRegistryContractPayeeshipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperRegistryContractPayeeshipTransferred represents a PayeeshipTransferred event raised by the KeeperRegistryContract contract.
type KeeperRegistryContractPayeeshipTransferred struct {
	Keeper common.Address
	From   common.Address
	To     common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterPayeeshipTransferred is a free log retrieval operation binding the contract event 0x78af32efdcad432315431e9b03d27e6cd98fb79c405fdc5af7c1714d9c0f75b3.
//
// Solidity: event PayeeshipTransferred(address indexed keeper, address indexed from, address indexed to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) FilterPayeeshipTransferred(opts *bind.FilterOpts, keeper []common.Address, from []common.Address, to []common.Address) (*KeeperRegistryContractPayeeshipTransferredIterator, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.FilterLogs(opts, "PayeeshipTransferred", keeperRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractPayeeshipTransferredIterator{contract: _KeeperRegistryContract.contract, event: "PayeeshipTransferred", logs: logs, sub: sub}, nil
}

// WatchPayeeshipTransferred is a free log subscription operation binding the contract event 0x78af32efdcad432315431e9b03d27e6cd98fb79c405fdc5af7c1714d9c0f75b3.
//
// Solidity: event PayeeshipTransferred(address indexed keeper, address indexed from, address indexed to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) WatchPayeeshipTransferred(opts *bind.WatchOpts, sink chan<- *KeeperRegistryContractPayeeshipTransferred, keeper []common.Address, from []common.Address, to []common.Address) (event.Subscription, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.WatchLogs(opts, "PayeeshipTransferred", keeperRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperRegistryContractPayeeshipTransferred)
				if err := _KeeperRegistryContract.contract.UnpackLog(event, "PayeeshipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePayeeshipTransferred is a log parse operation binding the contract event 0x78af32efdcad432315431e9b03d27e6cd98fb79c405fdc5af7c1714d9c0f75b3.
//
// Solidity: event PayeeshipTransferred(address indexed keeper, address indexed from, address indexed to)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) ParsePayeeshipTransferred(log types.Log) (*KeeperRegistryContractPayeeshipTransferred, error) {
	event := new(KeeperRegistryContractPayeeshipTransferred)
	if err := _KeeperRegistryContract.contract.UnpackLog(event, "PayeeshipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeeperRegistryContractPaymentWithdrawnIterator is returned from FilterPaymentWithdrawn and is used to iterate over the raw logs and unpacked data for PaymentWithdrawn events raised by the KeeperRegistryContract contract.
type KeeperRegistryContractPaymentWithdrawnIterator struct {
	Event *KeeperRegistryContractPaymentWithdrawn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeeperRegistryContractPaymentWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperRegistryContractPaymentWithdrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeeperRegistryContractPaymentWithdrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeeperRegistryContractPaymentWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperRegistryContractPaymentWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperRegistryContractPaymentWithdrawn represents a PaymentWithdrawn event raised by the KeeperRegistryContract contract.
type KeeperRegistryContractPaymentWithdrawn struct {
	Keeper common.Address
	Amount *big.Int
	To     common.Address
	Payee  common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterPaymentWithdrawn is a free log retrieval operation binding the contract event 0x9819093176a1851202c7bcfa46845809b4e47c261866550e94ed3775d2f40698.
//
// Solidity: event PaymentWithdrawn(address indexed keeper, uint256 indexed amount, address indexed to, address payee)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) FilterPaymentWithdrawn(opts *bind.FilterOpts, keeper []common.Address, amount []*big.Int, to []common.Address) (*KeeperRegistryContractPaymentWithdrawnIterator, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}
	var amountRule []interface{}
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.FilterLogs(opts, "PaymentWithdrawn", keeperRule, amountRule, toRule)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractPaymentWithdrawnIterator{contract: _KeeperRegistryContract.contract, event: "PaymentWithdrawn", logs: logs, sub: sub}, nil
}

// WatchPaymentWithdrawn is a free log subscription operation binding the contract event 0x9819093176a1851202c7bcfa46845809b4e47c261866550e94ed3775d2f40698.
//
// Solidity: event PaymentWithdrawn(address indexed keeper, uint256 indexed amount, address indexed to, address payee)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) WatchPaymentWithdrawn(opts *bind.WatchOpts, sink chan<- *KeeperRegistryContractPaymentWithdrawn, keeper []common.Address, amount []*big.Int, to []common.Address) (event.Subscription, error) {

	var keeperRule []interface{}
	for _, keeperItem := range keeper {
		keeperRule = append(keeperRule, keeperItem)
	}
	var amountRule []interface{}
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.WatchLogs(opts, "PaymentWithdrawn", keeperRule, amountRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperRegistryContractPaymentWithdrawn)
				if err := _KeeperRegistryContract.contract.UnpackLog(event, "PaymentWithdrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePaymentWithdrawn is a log parse operation binding the contract event 0x9819093176a1851202c7bcfa46845809b4e47c261866550e94ed3775d2f40698.
//
// Solidity: event PaymentWithdrawn(address indexed keeper, uint256 indexed amount, address indexed to, address payee)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) ParsePaymentWithdrawn(log types.Log) (*KeeperRegistryContractPaymentWithdrawn, error) {
	event := new(KeeperRegistryContractPaymentWithdrawn)
	if err := _KeeperRegistryContract.contract.UnpackLog(event, "PaymentWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeeperRegistryContractUpkeepCanceledIterator is returned from FilterUpkeepCanceled and is used to iterate over the raw logs and unpacked data for UpkeepCanceled events raised by the KeeperRegistryContract contract.
type KeeperRegistryContractUpkeepCanceledIterator struct {
	Event *KeeperRegistryContractUpkeepCanceled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeeperRegistryContractUpkeepCanceledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperRegistryContractUpkeepCanceled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeeperRegistryContractUpkeepCanceled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeeperRegistryContractUpkeepCanceledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperRegistryContractUpkeepCanceledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperRegistryContractUpkeepCanceled represents a UpkeepCanceled event raised by the KeeperRegistryContract contract.
type KeeperRegistryContractUpkeepCanceled struct {
	Id            *big.Int
	AtBlockHeight uint64
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterUpkeepCanceled is a free log retrieval operation binding the contract event 0x91cb3bb75cfbd718bbfccc56b7f53d92d7048ef4ca39a3b7b7c6d4af1f791181.
//
// Solidity: event UpkeepCanceled(uint256 indexed id, uint64 indexed atBlockHeight)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) FilterUpkeepCanceled(opts *bind.FilterOpts, id []*big.Int, atBlockHeight []uint64) (*KeeperRegistryContractUpkeepCanceledIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var atBlockHeightRule []interface{}
	for _, atBlockHeightItem := range atBlockHeight {
		atBlockHeightRule = append(atBlockHeightRule, atBlockHeightItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.FilterLogs(opts, "UpkeepCanceled", idRule, atBlockHeightRule)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractUpkeepCanceledIterator{contract: _KeeperRegistryContract.contract, event: "UpkeepCanceled", logs: logs, sub: sub}, nil
}

// WatchUpkeepCanceled is a free log subscription operation binding the contract event 0x91cb3bb75cfbd718bbfccc56b7f53d92d7048ef4ca39a3b7b7c6d4af1f791181.
//
// Solidity: event UpkeepCanceled(uint256 indexed id, uint64 indexed atBlockHeight)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) WatchUpkeepCanceled(opts *bind.WatchOpts, sink chan<- *KeeperRegistryContractUpkeepCanceled, id []*big.Int, atBlockHeight []uint64) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var atBlockHeightRule []interface{}
	for _, atBlockHeightItem := range atBlockHeight {
		atBlockHeightRule = append(atBlockHeightRule, atBlockHeightItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.WatchLogs(opts, "UpkeepCanceled", idRule, atBlockHeightRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperRegistryContractUpkeepCanceled)
				if err := _KeeperRegistryContract.contract.UnpackLog(event, "UpkeepCanceled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpkeepCanceled is a log parse operation binding the contract event 0x91cb3bb75cfbd718bbfccc56b7f53d92d7048ef4ca39a3b7b7c6d4af1f791181.
//
// Solidity: event UpkeepCanceled(uint256 indexed id, uint64 indexed atBlockHeight)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) ParseUpkeepCanceled(log types.Log) (*KeeperRegistryContractUpkeepCanceled, error) {
	event := new(KeeperRegistryContractUpkeepCanceled)
	if err := _KeeperRegistryContract.contract.UnpackLog(event, "UpkeepCanceled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeeperRegistryContractUpkeepPerformedIterator is returned from FilterUpkeepPerformed and is used to iterate over the raw logs and unpacked data for UpkeepPerformed events raised by the KeeperRegistryContract contract.
type KeeperRegistryContractUpkeepPerformedIterator struct {
	Event *KeeperRegistryContractUpkeepPerformed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeeperRegistryContractUpkeepPerformedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperRegistryContractUpkeepPerformed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeeperRegistryContractUpkeepPerformed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeeperRegistryContractUpkeepPerformedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperRegistryContractUpkeepPerformedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperRegistryContractUpkeepPerformed represents a UpkeepPerformed event raised by the KeeperRegistryContract contract.
type KeeperRegistryContractUpkeepPerformed struct {
	Id          *big.Int
	Success     bool
	From        common.Address
	Payment     *big.Int
	PerformData []byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUpkeepPerformed is a free log retrieval operation binding the contract event 0xcaacad83e47cc45c280d487ec84184eee2fa3b54ebaa393bda7549f13da228f6.
//
// Solidity: event UpkeepPerformed(uint256 indexed id, bool indexed success, address indexed from, uint96 payment, bytes performData)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) FilterUpkeepPerformed(opts *bind.FilterOpts, id []*big.Int, success []bool, from []common.Address) (*KeeperRegistryContractUpkeepPerformedIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var successRule []interface{}
	for _, successItem := range success {
		successRule = append(successRule, successItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.FilterLogs(opts, "UpkeepPerformed", idRule, successRule, fromRule)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractUpkeepPerformedIterator{contract: _KeeperRegistryContract.contract, event: "UpkeepPerformed", logs: logs, sub: sub}, nil
}

// WatchUpkeepPerformed is a free log subscription operation binding the contract event 0xcaacad83e47cc45c280d487ec84184eee2fa3b54ebaa393bda7549f13da228f6.
//
// Solidity: event UpkeepPerformed(uint256 indexed id, bool indexed success, address indexed from, uint96 payment, bytes performData)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) WatchUpkeepPerformed(opts *bind.WatchOpts, sink chan<- *KeeperRegistryContractUpkeepPerformed, id []*big.Int, success []bool, from []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var successRule []interface{}
	for _, successItem := range success {
		successRule = append(successRule, successItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.WatchLogs(opts, "UpkeepPerformed", idRule, successRule, fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperRegistryContractUpkeepPerformed)
				if err := _KeeperRegistryContract.contract.UnpackLog(event, "UpkeepPerformed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpkeepPerformed is a log parse operation binding the contract event 0xcaacad83e47cc45c280d487ec84184eee2fa3b54ebaa393bda7549f13da228f6.
//
// Solidity: event UpkeepPerformed(uint256 indexed id, bool indexed success, address indexed from, uint96 payment, bytes performData)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) ParseUpkeepPerformed(log types.Log) (*KeeperRegistryContractUpkeepPerformed, error) {
	event := new(KeeperRegistryContractUpkeepPerformed)
	if err := _KeeperRegistryContract.contract.UnpackLog(event, "UpkeepPerformed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeeperRegistryContractUpkeepRegisteredIterator is returned from FilterUpkeepRegistered and is used to iterate over the raw logs and unpacked data for UpkeepRegistered events raised by the KeeperRegistryContract contract.
type KeeperRegistryContractUpkeepRegisteredIterator struct {
	Event *KeeperRegistryContractUpkeepRegistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeeperRegistryContractUpkeepRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperRegistryContractUpkeepRegistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeeperRegistryContractUpkeepRegistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeeperRegistryContractUpkeepRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperRegistryContractUpkeepRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperRegistryContractUpkeepRegistered represents a UpkeepRegistered event raised by the KeeperRegistryContract contract.
type KeeperRegistryContractUpkeepRegistered struct {
	Id         *big.Int
	ExecuteGas uint32
	Admin      common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterUpkeepRegistered is a free log retrieval operation binding the contract event 0xbae366358c023f887e791d7a62f2e4316f1026bd77f6fb49501a917b3bc5d012.
//
// Solidity: event UpkeepRegistered(uint256 indexed id, uint32 executeGas, address admin)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) FilterUpkeepRegistered(opts *bind.FilterOpts, id []*big.Int) (*KeeperRegistryContractUpkeepRegisteredIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.FilterLogs(opts, "UpkeepRegistered", idRule)
	if err != nil {
		return nil, err
	}
	return &KeeperRegistryContractUpkeepRegisteredIterator{contract: _KeeperRegistryContract.contract, event: "UpkeepRegistered", logs: logs, sub: sub}, nil
}

// WatchUpkeepRegistered is a free log subscription operation binding the contract event 0xbae366358c023f887e791d7a62f2e4316f1026bd77f6fb49501a917b3bc5d012.
//
// Solidity: event UpkeepRegistered(uint256 indexed id, uint32 executeGas, address admin)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) WatchUpkeepRegistered(opts *bind.WatchOpts, sink chan<- *KeeperRegistryContractUpkeepRegistered, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _KeeperRegistryContract.contract.WatchLogs(opts, "UpkeepRegistered", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperRegistryContractUpkeepRegistered)
				if err := _KeeperRegistryContract.contract.UnpackLog(event, "UpkeepRegistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpkeepRegistered is a log parse operation binding the contract event 0xbae366358c023f887e791d7a62f2e4316f1026bd77f6fb49501a917b3bc5d012.
//
// Solidity: event UpkeepRegistered(uint256 indexed id, uint32 executeGas, address admin)
func (_KeeperRegistryContract *KeeperRegistryContractFilterer) ParseUpkeepRegistered(log types.Log) (*KeeperRegistryContractUpkeepRegistered, error) {
	event := new(KeeperRegistryContractUpkeepRegistered)
	if err := _KeeperRegistryContract.contract.UnpackLog(event, "UpkeepRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
