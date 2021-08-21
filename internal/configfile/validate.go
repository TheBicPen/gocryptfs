package configfile

import (
	"fmt"

	"github.com/rfjakob/gocryptfs/v2/internal/contentenc"
)

// Validate that the combination of settings makes sense and is supported
func (cf *ConfFile) Validate() error {
	if cf.Version != contentenc.CurrentVersion {
		return fmt.Errorf("Unsupported on-disk format %d", cf.Version)
	}
	// scrypt params ok?
	if err := cf.ScryptObject.validateParams(); err != nil {
		return err
	}
	// All feature flags that are in the config file are known?
	for _, flag := range cf.FeatureFlags {
		if !isFeatureFlagKnown(flag) {
			return fmt.Errorf("Unknown feature flag %q", flag)
		}
	}
	// File content encryption
	{
		switch {
		case cf.IsFeatureFlagSet(FlagXChaCha20Poly1305) && cf.IsFeatureFlagSet(FlagAESSIV):
			return fmt.Errorf("Can't have both XChaCha20Poly1305 and AESSIV feature flags")
		case cf.IsFeatureFlagSet(FlagAESSIV):
			if !cf.IsFeatureFlagSet(FlagGCMIV128) {
				return fmt.Errorf("AESSIV requires GCMIV128 feature flag")
			}
		case cf.IsFeatureFlagSet(FlagXChaCha20Poly1305):
			if cf.IsFeatureFlagSet(FlagGCMIV128) {
				return fmt.Errorf("XChaCha20Poly1305 conflicts with GCMIV128 feature flag")
			}
			if !cf.IsFeatureFlagSet(FlagHKDF) {
				return fmt.Errorf("XChaCha20Poly1305 requires HKDF feature flag")
			}
		// The absence of other flags means AES-GCM (oldest algorithm)
		case !cf.IsFeatureFlagSet(FlagXChaCha20Poly1305) && !cf.IsFeatureFlagSet(FlagAESSIV):
			if !cf.IsFeatureFlagSet(FlagGCMIV128) {
				return fmt.Errorf("AES-GCM requires GCMIV128 feature flag")
			}
		}
	}
	// Filename encryption
	{
		switch {
		case cf.IsFeatureFlagSet(FlagPlaintextNames) && cf.IsFeatureFlagSet(FlagEMENames):
			return fmt.Errorf("Can't have both PlaintextNames and EMENames feature flags")
		case cf.IsFeatureFlagSet(FlagPlaintextNames):
			if cf.IsFeatureFlagSet(FlagDirIV) {
				return fmt.Errorf("PlaintextNames conflicts with DirIV feature flag")
			}
			if cf.IsFeatureFlagSet(FlagLongNames) {
				return fmt.Errorf("PlaintextNames conflicts with LongNames feature flag")
			}
			if cf.IsFeatureFlagSet(FlagRaw64) {
				return fmt.Errorf("PlaintextNames conflicts with Raw64 feature flag")
			}
		case cf.IsFeatureFlagSet(FlagEMENames):
			// All combinations of DirIV, LongNames, Raw64 allowed
		}
	}
	return nil
}
