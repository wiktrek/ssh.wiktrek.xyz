package main

import "time"

const (
	hostKeyPath = ".ssh/term_info_ed25519"

	errorColor = "#ff5f5f"

	workerCost            = 50
	workerPriceMultiplier = 1.2
	workerIncomeInterval  = 3 * time.Second
)
