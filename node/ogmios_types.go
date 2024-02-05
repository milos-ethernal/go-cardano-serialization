package node

type ogmiosNode struct {
	address string
}

type queryLedgerStateUtxoParams struct {
	Addresses []string `json:"addresses"`
}

type queryLedgerStateUtxo struct {
	Jsonrpc string                     `json:"jsonrpc"`
	Method  string                     `json:"method"`
	Params  queryLedgerStateUtxoParams `json:"params"`
	Id      interface{}                `json:"id"`
}

type queryLedgerStateUtxoResponseResultTransaction struct {
	Id string `json:"id"`
}

type queryLedgerStateUtxoResponseResultValueAda struct {
	Lovelace uint `json:"lovelace"`
}

type queryLedgerStateUtxoResponseResultValue struct {
	Ada queryLedgerStateUtxoResponseResultValueAda `json:"ada"`
}

type queryLedgerStateUtxoResponseResult struct {
	Transaction queryLedgerStateUtxoResponseResultTransaction `json:"transaction"`
	Index       uint                                          `json:"index"`
	Address     string                                        `json:"address"`
	Value       queryLedgerStateUtxoResponseResultValue       `json:"value"`
}

type queryLedgerStateUtxoResponse struct {
	Jsonrpc string                               `json:"jsonrpc"`
	Method  string                               `json:"method"`
	Result  []queryLedgerStateUtxoResponseResult `json:"result"`
	Id      interface{}                          `json:"id"`
}

// Define the types for the protocol parameters response
type queryLedgerStateProtocolParametersResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Result  struct {
		MinFeeCoefficient uint `json:"minFeeCoefficient"`
		MinFeeConstant    struct {
			Ada struct {
				Lovelace uint `json:"lovelace"`
			} `json:"ada"`
		} `json:"minFeeConstant"`
		MaxBlockBodySize struct {
			Bytes uint `json:"bytes"`
		} `json:"maxBlockBodySize"`
		MaxBlockHeaderSize struct {
			Bytes uint `json:"bytes"`
		} `json:"maxBlockHeaderSize"`
		MaxTransactionSize struct {
			Bytes uint `json:"bytes"`
		} `json:"maxTransactionSize"`
		StakeCredentialDeposit struct {
			Ada struct {
				Lovelace uint `json:"lovelace"`
			} `json:"ada"`
		} `json:"stakeCredentialDeposit"`
		StakePoolDeposit struct {
			Ada struct {
				Lovelace uint `json:"lovelace"`
			} `json:"ada"`
		} `json:"stakePoolDeposit"`
		StakePoolRetirementEpochBound uint   `json:"stakePoolRetirementEpochBound"`
		DesiredNumberOfStakePools     uint   `json:"desiredNumberOfStakePools"`
		StakePoolPledgeInfluence      string `json:"stakePoolPledgeInfluence"`
		MonetaryExpansion             string `json:"monetaryExpansion"`
		TreasuryExpansion             string `json:"treasuryExpansion"`
		MinStakePoolCost              struct {
			Ada struct {
				Lovelace uint `json:"lovelace"`
			} `json:"ada"`
		} `json:"minStakePoolCost"`
		MinUtxoDepositConstant struct {
			Ada struct {
				Lovelace uint `json:"lovelace"`
			} `json:"ada"`
		} `json:"minUtxoDepositConstant"`
		MinUtxoDepositCoefficient uint              `json:"minUtxoDepositCoefficient"`
		PlutusCostModels          map[string][]uint `json:"plutusCostModels"`
		ScriptExecutionPrices     struct {
			Memory string `json:"memory"`
			CPU    string `json:"cpu"`
		} `json:"scriptExecutionPrices"`
		MaxExecutionUnitsPerTransaction struct {
			Memory uint `json:"memory"`
			CPU    uint `json:"cpu"`
		} `json:"maxExecutionUnitsPerTransaction"`
		MaxExecutionUnitsPerBlock struct {
			Memory uint `json:"memory"`
			CPU    uint `json:"cpu"`
		} `json:"maxExecutionUnitsPerBlock"`
		MaxValueSize struct {
			Bytes uint `json:"bytes"`
		} `json:"maxValueSize"`
		CollateralPercentage uint `json:"collateralPercentage"`
		MaxCollateralInputs  uint `json:"maxCollateralInputs"`
		Version              struct {
			Major uint `json:"major"`
			Minor uint `json:"minor"`
		} `json:"version"`
	} `json:"result"`
	Id interface{} `json:"id"`
}

type queryLedgerStateTipResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Result  struct {
		Slot uint   `json:"slot"`
		ID   string `json:"id"`
	} `json:"result"`
	Id interface{} `json:"id"`
}

type queryLedgerStateProtocolParameters queryLedgerState
type queryLedgerStateTip queryLedgerState

type queryLedgerState struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Id      interface{} `json:"id"`
}

type submitTransactionParamsTransaction struct {
	CBOR string `json:"cbor"`
}

type submitTransactionParams struct {
	Transaction submitTransactionParamsTransaction `json:"transaction"`
}

type submitTransaction struct {
	Jsonrpc string                  `json:"jsonrpc"`
	Method  string                  `json:"method"`
	Params  submitTransactionParams `json:"params"`
	Id      interface{}             `json:"id"`
}

// Define the types for the submit transaction response
type submitTransactionResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Result  struct {
		Transaction struct {
			ID string `json:"id"`
		} `json:"transaction"`
	} `json:"result"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			MissingScripts []string `json:"missingScripts"`
		} `json:"data"`
	} `json:"error"`
	Id interface{} `json:"id"`
}
