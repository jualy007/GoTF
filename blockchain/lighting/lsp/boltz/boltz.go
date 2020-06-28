package boltz

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/lightningnetwork/lnd/input"
	"github.com/lightningnetwork/lnd/lnwallet/chainfee"

	"github.com/jualy007/GoTF/blockchain/btc"
)

const (
	getPairsEndpoint      = "/getpairs"
	getfeeestimation      = "/getfeeestimation"
	createSwapEndpoint    = "/createswap"
	swapStatusEndpoint    = "/swapstatus"
	claimWitnessInputSize = 1 + 1 + 8 + 73 + 1 + 32 + 1 + 100
	NormalSwaps           = "submarine"
	ReverseSwaps          = "reversesubmarine"
)

var apiURL = "https://boltz.exchange/api"

var chain = &chaincfg.MainNetParams

type BadRequestError string

type StatusInfo struct {
	Status      string `json:"status"`
	Transaction struct {
		ID  string `json:"id"`
		Hex string `json:"hex"`
		ETA int32  `json:"eta", omitempty`
	} `json:"transaction", omitempty`
}

type ReverseSwap struct {
	RSResponse
	Preimage string
	Key      string
}

type PairsInfo struct {
	Warnings []string `json:"warnings"`
	Pairs    map[string]struct {
		Rate   float64 `json:"rate"`
		Limits struct {
			Maximal         int64 `json:"maximal"`
			Minimal         int64 `json:"minimal"`
			MaximalZeroConf struct {
				BaseAsset  int64 `json:"baseAsset"`
				QuoteAsset int64 `json:"quoteAsset"`
			} `json:"maximalZeroConf"`
		}
		Fees struct {
			Percentage float64 `json:"percentage"`
			MinerFees  struct {
				BaseAsset struct {
					Normal  int64 `json:"normal"`
					Reverse struct {
						Lockup int64 `json:"lockup"`
						Claim  int64 `json:"claim"`
					} `json:"reverse"`
				} `json:"baseAsset"`
				QuoteAsset struct {
					Normal  int64 `json:"normal"`
					Reverse struct {
						Lockup int64 `json:"lockup"`
						Claim  int64 `json:"claim"`
					} `json:"reverse"`
				} `json:"quoteAsset"`
			} `json:"minerFees"`
		} `json:"fees"`
	} `json:"pairs"`
}

type BaseRequest struct {
	SwapType  string `json:"type"`
	PairId    string `json:"pairId"`
	OrderSide string `json:"orderSide"`
}

type BaseResponse struct {
	ID                 string `json:"id"`
	RedeemScript       string `json:"redeemScript"`
	TimeoutBlockHeight int64  `json:"timeoutBlockHeight"`
}

type SRequest struct {
	BaseRequest
	Invoice         string `json:"invoice"`
	RefundPublicKey string `json:"refundPublicKey"`
}

type SResponse struct {
	BaseResponse
	AcceptZeroConf bool   `json:"acceptZeroConf"`
	Address        string `json:"address"`
	ExpectedAmount int64  `json:"expectedAmount"`
	Bip21          string `json:"bip21"`
}

type RSRequest struct {
	BaseRequest
	InvoiceAmount  int64  `json:"invoiceAmount"`
	PreimageHash   string `json:"preimageHash"`
	ClaimPublicKey string `json:"claimPublicKey"`
}

type RSResponse struct {
	BaseResponse
	Invoice       string `json:"invoice"`
	OnchainAmount int64  `json:"onchainAmount"`
	LockupAddress string `json:"lockupAddress"`
}

type Fees struct {
	fee map[string]int32
}

func (e *BadRequestError) Error() string {
	return string(*e)
}

func TestNet() {
	apiURL = "https://testnet.boltz.exchange/api"
	chain = &chaincfg.TestNet3Params
}

func GetFees() (fees *Fees, err error) {
	//It is important to mention that if 0-conf wants to be used with normal swaps,
	//the lockup transaction has to have at least 80% of the recommended sat/vbyte value.
	resp, err := http.Get(apiURL + getfeeestimation)

	if err != nil {
		return nil, fmt.Errorf("getfeeestimation get %v: %w", apiURL+getfeeestimation, err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&fees)
	return fees, err
}

func GetSwapInfo(id string) (*http.Response, error) {
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(struct {
		ID string `json:"id"`
	}{ID: id})

	resp, err := http.Post(apiURL+swapStatusEndpoint, "application/json", buffer)
	if err != nil {
		err = fmt.Errorf("swapstatus post %v: %w", apiURL+swapStatusEndpoint, err)
		return nil, err
	}

	defer resp.Body.Close()

	return resp, nil
}

func GetReverseSwapInfo() (pairinfo *PairsInfo, err error) {
	resp, err := http.Get(apiURL + getPairsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("getpairs get %v: %w", apiURL+getPairsEndpoint, err)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&pairinfo)
	if err != nil {
		return nil, fmt.Errorf("json decode (status: %v): %w", resp.Status, err)
	}

	return pairinfo, err
}

func createReverseSwap(amt int64, preimage []byte, key *btcec.PrivateKey) (*RSResponse, error) {
	h := sha256.Sum256(preimage)
	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(struct {
		Type           string `json:"type"`
		PairID         string `json:"pairId"`
		OrderSide      string `json:"orderSide"`
		InvoiceAmount  int64  `json:"invoiceAmount"`
		PreimageHash   string `json:"preimageHash"`
		ClaimPublicKey string `json:"claimPublicKey"`
	}{
		Type:           "reversesubmarine",
		PairID:         "BTC/BTC",
		OrderSide:      "buy",
		InvoiceAmount:  amt,
		PreimageHash:   hex.EncodeToString(h[:]),
		ClaimPublicKey: hex.EncodeToString(key.PubKey().SerializeCompressed()),
	})
	if err != nil {
		return nil, fmt.Errorf("json encode %v, %v, %v: %w", amt, preimage, key, err)
	}

	resp, err := http.Post(apiURL+createSwapEndpoint, "application/json", buffer)
	if err != nil {
		return nil, fmt.Errorf("createswap post %v: %w", apiURL+createSwapEndpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		e := struct {
			Error string `json:"error"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&e)
		if err != nil {
			return nil, fmt.Errorf("json decode (status: %v): %w", resp.Status, err)
		}
		badRequestError := BadRequestError(e.Error)
		return nil, fmt.Errorf("createswap result (status: %v) %w", resp.Status, &badRequestError)
	}

	var rs boltzReverseSwap
	err = json.NewDecoder(resp.Body).Decode(&rs)
	if err != nil {
		return nil, fmt.Errorf("json decode (status ok): %w", err)
	}

	return &rs, nil
}

func checkReverseSwap(preimage []byte, key *btcec.PrivateKey, rs *RSResponse) error {
	script, err := hex.DecodeString(rs.RedeemScript)
	if err != nil {
		return fmt.Errorf("hex.DecodeString %v: %w", rs.RedeemScript, err)
	}
	dis, err := txscript.DisasmString(script)
	if err != nil {
		return fmt.Errorf("txscript.DisasmString %x: %w", script, err)
	}
	d := strings.Split(dis, " ")
	h := sha256.Sum256(preimage)

	s := fmt.Sprintf(
		"OP_SIZE 20 OP_EQUAL OP_IF OP_HASH160 %x OP_EQUALVERIFY %x OP_ELSE OP_DROP %s OP_CHECKLOCKTIMEVERIFY OP_DROP %s OP_ENDIF OP_CHECKSIG",
		input.Ripemd160H(h[:]),
		key.PubKey().SerializeCompressed(),
		btc.CheckHeight(rs.TimeoutBlockHeight, d[10]),
		d[13],
	)
	if s != dis {
		return fmt.Errorf("bad script")
	}
	a, err := addressWitnessScriptHash(script, chain)
	if err != nil {
		return fmt.Errorf("addressWitnessScriptHash %v: %w", script, err)
	}

	if rs.LockupAddress != a.String() {
		return fmt.Errorf("bad address: %v instead of %v", rs.LockupAddress, a.String())
	}
	return nil
}

func addressWitnessScriptHash(script []byte, net *chaincfg.Params) (*btcutil.AddressWitnessScriptHash, error) {
	witnessProg := sha256.Sum256(script)
	return btcutil.NewAddressWitnessScriptHash(witnessProg[:], net)
}

func getPreimage() []byte {
	preimage := make([]byte, 32)
	rand.Read(preimage)
	return preimage
}

func getPrivate() (*btcec.PrivateKey, error) {
	k, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, fmt.Errorf("btcec.NewPrivateKey: %w", err)
	}
	return k, nil
}

// NewReverseSwap begins the reverse submarine process.
func NewReverseSwap(amt btcutil.Amount) (*ReverseSwap, error) {
	preimage := getPreimage()

	key, err := getPrivate()
	if err != nil {
		return nil, fmt.Errorf("getPrivate: %w", err)
	}

	rs, err := createReverseSwap(int64(amt), preimage, key)
	if err != nil {
		return nil, fmt.Errorf("createReverseSwap amt:%v, preimage:%x, key:%x; %w", amt, preimage, key, err)
	}

	err = checkReverseSwap(preimage, key, rs)
	if err != nil {
		return nil, fmt.Errorf("checkReverseSwap preimage:%x, key:%x, %#v; %w", preimage, key, rs, err)
	}

	return &ReverseSwap{*rs, hex.EncodeToString(preimage), hex.EncodeToString(key.Serialize())}, nil
}

// CheckTransaction checks that the transaction corresponds to the adresss and amount
func CheckTransaction(transactionHex, lockupAddress string, amt int64) (string, error) {
	txSerialized, err := hex.DecodeString(transactionHex)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString(%v): %w", transactionHex, err)
	}
	tx, err := btcutil.NewTxFromBytes(txSerialized)
	if err != nil {
		return "", fmt.Errorf("btcutil.NewTxFromBytes(%x): %w", txSerialized, err)
	}
	var out *wire.OutPoint
	for i, txout := range tx.MsgTx().TxOut {
		class, addresses, requiredsigs, err := txscript.ExtractPkScriptAddrs(txout.PkScript, chain)
		if err != nil {
			return "", fmt.Errorf("txscript.ExtractPkScriptAddrs(%x) %w", txout.PkScript, err)
		}
		if class == txscript.WitnessV0ScriptHashTy && len(addresses) == 1 && addresses[0].EncodeAddress() == lockupAddress && requiredsigs == 1 {
			out = wire.NewOutPoint(tx.Hash(), uint32(i))
			if int64(amt) != txout.Value {
				return "", fmt.Errorf("bad amount: %v != %v", int64(amt), txout.Value)
			}
		}
	}
	if out == nil {
		return "", fmt.Errorf("lockupAddress: %v not found in the transaction: %v", lockupAddress, transactionHex)
	}
	return tx.Hash().String(), nil
}

// GetTransaction return the transaction after paying the ln invoice
func GetTransaction(id, lockupAddress string, amt int64) (status, txid, tx string, eta int, err error) {
	buffer := new(bytes.Buffer)
	err = json.NewEncoder(buffer).Encode(struct {
		ID string `json:"id"`
	}{ID: id})
	if err != nil {
		err = fmt.Errorf("json encode %v: %w", id, err)
		return
	}
	resp, err := http.Post(apiURL+swapStatusEndpoint, "application/json", buffer)
	if err != nil {
		err = fmt.Errorf("swapstatus post %v: %w", apiURL+swapStatusEndpoint, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		e := struct {
			Error string `json:"error"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&e)
		if err != nil {
			err = fmt.Errorf("json decode (status: %v): %w", resp.Status, err)
			return
		}
		badRequestError := BadRequestError(e.Error)
		err = fmt.Errorf("createswap result (status: %v) %w", resp.Status, &badRequestError)
		return
	}

	var ts transactionStatus
	err = json.NewDecoder(resp.Body).Decode(&ts)
	if err != nil {
		err = fmt.Errorf("json decode (status ok): %w", err)
		return
	}
	if ts.Status != "transaction.mempool" && ts.Status != "transaction.confirmed" {
		err = fmt.Errorf("transaction not in mempool or settled/canceled")
		return
	}

	if lockupAddress != "" {
		var calculatedTxid string
		calculatedTxid, err = CheckTransaction(ts.Transaction.Hex, lockupAddress, amt)
		if err != nil {
			err = fmt.Errorf("CheckTransaction(%v, %v, %v): %w)", ts.Transaction.Hex, lockupAddress, amt, err)
			return
		}
		if calculatedTxid != ts.Transaction.ID {
			err = fmt.Errorf("bad txid: %v != %v", ts.Transaction.ID, calculatedTxid)
			return
		}
	}
	status = ts.Status
	tx = ts.Transaction.Hex
	txid = ts.Transaction.ID
	eta = ts.Transaction.ETA
	return
}

//ClaimFees return the fees needed for the claimed transaction for a feePerKw
func ClaimFee(claimAddress string, feePerKw int64) (int64, error) {
	addr, err := btcutil.DecodeAddress(claimAddress, chain)
	if err != nil {
		return 0, fmt.Errorf("btcutil.DecodeAddress(%v) %w", addr, err)
	}
	claimScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return 0, fmt.Errorf("txscript.PayToAddrScript(%v): %w", addr.String(), err)
	}
	claimTx := wire.NewMsgTx(1)
	txIn := wire.NewTxIn(&wire.OutPoint{}, nil, nil)
	txIn.Sequence = 0
	claimTx.AddTxIn(txIn)
	txOut := wire.TxOut{PkScript: claimScript}
	claimTx.AddTxOut(&txOut)

	// Calcluate the weight and the fee
	weight := 4*claimTx.SerializeSizeStripped() + claimWitnessInputSize*len(claimTx.TxIn)
	fee := chainfee.SatPerKWeight(feePerKw).FeeForWeight(int64(weight))
	return int64(fee), nil
}

func claimTransaction(
	script []byte,
	amt btcutil.Amount,
	txout *wire.OutPoint,
	claimAddress btcutil.Address,
	preimage []byte,
	privateKey []byte,
	fees btcutil.Amount,
) ([]byte, error) {
	claimTx := wire.NewMsgTx(1)
	txIn := wire.NewTxIn(txout, nil, nil)
	txIn.Sequence = 0
	claimTx.AddTxIn(txIn)

	claimScript, err := txscript.PayToAddrScript(claimAddress)
	if err != nil {
		return nil, fmt.Errorf("txscript.PayToAddrScript(%v): %w", claimAddress.String(), err)
	}
	txOut := wire.TxOut{PkScript: claimScript}
	claimTx.AddTxOut(&txOut)

	// Adjust the amount in the txout
	claimTx.TxOut[0].Value = int64(amt - fees)

	sigHashes := txscript.NewTxSigHashes(claimTx)
	key, _ := btcec.PrivKeyFromBytes(btcec.S256(), privateKey)
	scriptSig, err := txscript.RawTxInWitnessSignature(claimTx, sigHashes, 0, int64(amt), script, txscript.SigHashAll, key)
	if err != nil {
		return nil, fmt.Errorf("txscript.RawTxInWitnessSignature: %w", err)
	}
	claimTx.TxIn[0].Witness = [][]byte{scriptSig, preimage, script}

	var rawTx bytes.Buffer
	err = claimTx.Serialize(&rawTx)
	if err != nil {
		return nil, fmt.Errorf("claimTx.Serialize %#v: %w", claimTx, err)
	}
	return rawTx.Bytes(), nil
}

// ClaimTransaction returns the claim transaction to broadcast after sending it also to boltz
func ClaimTransaction(
	redeemScript, transactionHex string,
	claimAddress string,
	preimage, key string,
	fees int64,
) (string, error) {
	txSerialized, err := hex.DecodeString(transactionHex)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString(%v): %w", transactionHex, err)
	}
	tx, err := btcutil.NewTxFromBytes(txSerialized)
	if err != nil {
		return "", fmt.Errorf("btcutil.NewTxFromBytes(%x): %w", txSerialized, err)
	}

	script, err := hex.DecodeString(redeemScript)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString(%v): %w", redeemScript, err)
	}
	lockupAddress, err := addressWitnessScriptHash(script, chain)
	if err != nil {
		return "", fmt.Errorf("addressWitnessScriptHash %v: %w", script, err)
	}
	var out *wire.OutPoint
	var amt btcutil.Amount
	for i, txout := range tx.MsgTx().TxOut {
		class, addresses, requiredsigs, err := txscript.ExtractPkScriptAddrs(txout.PkScript, chain)
		if err != nil {
			return "", fmt.Errorf("txscript.ExtractPkScriptAddrs(%x) %w", txout.PkScript, err)
		}
		if class == txscript.WitnessV0ScriptHashTy && requiredsigs == 1 &&
			len(addresses) == 1 && addresses[0].EncodeAddress() == lockupAddress.EncodeAddress() {
			out = wire.NewOutPoint(tx.Hash(), uint32(i))
			amt = btcutil.Amount(txout.Value)
		}
	}

	addr, err := btcutil.DecodeAddress(claimAddress, chain)
	if err != nil {
		return "", fmt.Errorf("btcutil.DecodeAddress(%v) %w", claimAddress, err)
	}

	preim, err := hex.DecodeString(preimage)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString(%v): %w", preimage, err)
	}
	privateKey, err := hex.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString(%v): %w", key, err)
	}

	ctx, err := claimTransaction(script, amt, out, addr, preim, privateKey, btcutil.Amount(fees))
	if err != nil {
		return "", fmt.Errorf("claimTransaction: %w", err)
	}
	ctxHex := hex.EncodeToString(ctx)
	//Ignore the result of broadcasting the transaction via boltz
	_, _ = broadcastTransaction(ctxHex)
	return ctxHex, nil
}
