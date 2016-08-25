package vault

import (
	"encoding/hex"
	"fmt"

	"github.com/hashicorp/vault/helper/logformat"
	"github.com/hashicorp/vault/helper/pgpkeys"
	"github.com/hashicorp/vault/shamir"
)

// InitResult is used to provide the key parts back after
// they are generated as part of the initialization.
type InitResult struct {
	SecretShares   [][]byte
	RecoveryShares [][]byte
	RootToken      string
}

// Initialized checks if the Vault is already initialized
func (c *Core) Initialized() (bool, error) {
	// Check the barrier first
	init, err := c.barrier.Initialized()
	if err != nil {
		c.logger.Error("barrier init check failed", "error", err)
		return false, err
	}
	if !init {
		c.logger.Info("security barrier not initialized")
		return false, nil
	}

	// Verify the seal configuration
	sealConf, err := c.seal.BarrierConfig()
	if err != nil {
		return false, err
	}
	if sealConf == nil {
		return false, fmt.Errorf("barrier reports initialized but no seal configuration found")
	}

	return true, nil
}

func (c *Core) generateShares(sc *SealConfig) ([]byte, [][]byte, error) {
	// Generate a master key
	masterKey, err := c.barrier.GenerateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("key generation failed: %v", err)
	}

	// Return the master key if only a single key part is used
	var unsealKeys [][]byte
	if sc.SecretShares == 1 {
		unsealKeys = append(unsealKeys, masterKey)
	} else {
		// Split the master key using the Shamir algorithm
		shares, err := shamir.Split(masterKey, sc.SecretShares, sc.SecretThreshold)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate barrier shares: %v", err)
		}
		unsealKeys = shares
	}

	// If we have PGP keys, perform the encryption
	if len(sc.PGPKeys) > 0 {
		hexEncodedShares := make([][]byte, len(unsealKeys))
		for i, _ := range unsealKeys {
			hexEncodedShares[i] = []byte(hex.EncodeToString(unsealKeys[i]))
		}
		_, encryptedShares, err := pgpkeys.EncryptShares(hexEncodedShares, sc.PGPKeys)
		if err != nil {
			return nil, nil, err
		}
		unsealKeys = encryptedShares
	}

	return masterKey, unsealKeys, nil
}

// Initialize is used to initialize the Vault with the given
// configurations.
func (c *Core) Initialize(barrierConfig, recoveryConfig *SealConfig) (*InitResult, error) {
	logger := logformat.DeriveModuleLogger(logger, "init")

	if c.seal.RecoveryKeySupported() {
		if recoveryConfig == nil {
			return nil, fmt.Errorf("recovery configuration must be supplied")
		}

		if recoveryConfig.SecretShares < 1 {
			return nil, fmt.Errorf("recovery configuration must specify a positive number of shares")
		}

		// Check if the seal configuration is valid
		if err := recoveryConfig.Validate(); err != nil {
			logger.Error("invalid recovery configuration", "error", err)
			return nil, fmt.Errorf("invalid recovery configuration: %v", err)
		}
	}

	// Check if the seal configuration is valid
	if err := barrierConfig.Validate(); err != nil {
		logger.Error("invalid seal configuration", "error", err)
		return nil, fmt.Errorf("invalid seal configuration: %v", err)
	}

	// Avoid an initialization race
	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	// Check if we are initialized
	init, err := c.Initialized()
	if err != nil {
		return nil, err
	}
	if init {
		return nil, ErrAlreadyInit
	}

	err = c.seal.Init()
	if err != nil {
		logger.Error("failed to initialize seal", "error", err)
		return nil, fmt.Errorf("error initializing seal: %v", err)
	}

	err = c.seal.SetBarrierConfig(barrierConfig)
	if err != nil {
		logger.Error("failed to save barrier configuration", "error", err)
		return nil, fmt.Errorf("barrier configuration saving failed: %v", err)
	}

	barrierKey, barrierUnsealKeys, err := c.generateShares(barrierConfig)
	if err != nil {
		logger.Error("error generating shares", "error", err)
		return nil, err
	}

	// If we are storing shares, pop them out of the returned results and push
	// them through the seal
	if barrierConfig.StoredShares > 0 {
		var keysToStore [][]byte
		for i := 0; i < barrierConfig.StoredShares; i++ {
			keysToStore = append(keysToStore, barrierUnsealKeys[0])
			barrierUnsealKeys = barrierUnsealKeys[1:]
		}
		if err := c.seal.SetStoredKeys(keysToStore); err != nil {
			logger.Error("failed to store keys", "error", err)
			return nil, fmt.Errorf("failed to store keys: %v", err)
		}
	}

	results := &InitResult{
		SecretShares: barrierUnsealKeys,
	}

	// Initialize the barrier
	if err := c.barrier.Initialize(barrierKey); err != nil {
		logger.Error("failed to initialize barrier", "error", err)
		return nil, fmt.Errorf("failed to initialize barrier: %v", err)
	}
	if logger.IsInfo() {
		logger.Info("security barrier initialized", "shares", barrierConfig.SecretShares, "threshold", barrierConfig.SecretThreshold)
	}

	// Unseal the barrier
	if err := c.barrier.Unseal(barrierKey); err != nil {
		logger.Error("failed to unseal barrier", "error", err)
		return nil, fmt.Errorf("failed to unseal barrier: %v", err)
	}

	// Ensure the barrier is re-sealed
	defer func() {
		if err := c.barrier.Seal(); err != nil {
			logger.Error("failed to seal barrier", "error", err)
		}
	}()

	// Perform initial setup
	if err := c.setupCluster(); err != nil {
		c.stateLock.Unlock()
		logger.Error("cluster setup failed during init", "error", err)
		return nil, err
	}
	if err := c.postUnseal(); err != nil {
		logger.Error("post-unseal setup failed during init", "error", err)
		return nil, err
	}

	// Save the configuration regardless, but only generate a key if it's not
	// disabled. When using recovery keys they are stored in the barrier, so
	// this must happen post-unseal.
	if c.seal.RecoveryKeySupported() {
		err = c.seal.SetRecoveryConfig(recoveryConfig)
		if err != nil {
			logger.Error("failed to save recovery configuration", "error", err)
			return nil, fmt.Errorf("recovery configuration saving failed: %v", err)
		}

		if recoveryConfig.SecretShares > 0 {
			recoveryKey, recoveryUnsealKeys, err := c.generateShares(recoveryConfig)
			if err != nil {
				logger.Error("failed to generate recovery shares", "error", err)
				return nil, err
			}

			err = c.seal.SetRecoveryKey(recoveryKey)
			if err != nil {
				return nil, err
			}

			results.RecoveryShares = recoveryUnsealKeys
		}
	}

	// Generate a new root token
	rootToken, err := c.tokenStore.rootToken()
	if err != nil {
		logger.Error("root token generation failed", "error", err)
		return nil, err
	}
	results.RootToken = rootToken.ID
	logger.Info("root token generated")

	// Prepare to re-seal
	if err := c.preSeal(); err != nil {
		logger.Error("pre-seal teardown failed", "error", err)
		return nil, err
	}

	return results, nil
}

func (c *Core) UnsealWithStoredKeys() error {
	if !c.seal.StoredKeysSupported() {
		return nil
	}

	sealed, err := c.Sealed()
	if err != nil {
		c.logger.Error("error checking sealed status in auto-unseal", "error", err)
		return fmt.Errorf("error checking sealed status in auto-unseal: %s", err)
	}
	if !sealed {
		return nil
	}

	c.logger.Info("stored unseal keys supported, attempting fetch")
	keys, err := c.seal.GetStoredKeys()
	if err != nil {
		c.logger.Error("fetching stored unseal keys failed", "error", err)
		return &NonFatalError{Err: fmt.Errorf("fetching stored unseal keys failed: %v", err)}
	}
	if len(keys) == 0 {
		c.logger.Warn("stored unseal key(s) supported but none found")
	} else {
		unsealed := false
		keysUsed := 0
		for _, key := range keys {
			unsealed, err = c.Unseal(key)
			if err != nil {
				c.logger.Error("unseal with stored unseal key failed", "error", err)
				return &NonFatalError{Err: fmt.Errorf("unseal with stored key failed: %v", err)}
			}
			keysUsed += 1
			if unsealed {
				break
			}
		}
		if !unsealed {
			if c.logger.IsWarn() {
				c.logger.Warn("stored unseal key(s) used but Vault not unsealed yet", "stored_keys_used", keysUsed)
			}
		} else {
			if c.logger.IsInfo() {
				c.logger.Info("successfully unsealed with stored key(s)", "stored_keys_used", keysUsed)
			}
		}
	}

	return nil
}
