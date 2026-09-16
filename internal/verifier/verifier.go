package verifier

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"airborne/internal/types"
)

const ManifestFile = "manifest.json"
const CryptoManifestFile = "crypto_manifest.json"
const SignatureFile = "signature.json"
const PublicKeyFile = "public_key.pem"

type Verifier struct {
	kernelDir string
}

func New(kernelDir string) *Verifier {
	return &Verifier{kernelDir: kernelDir}
}

func (v *Verifier) Verify() (*types.VerificationResult, error) {
	result := &types.VerificationResult{
		VerifiedAt: time.Now().UTC(),
	}

	manifestBytes, err := v.readFile(ManifestFile)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	result.ManifestSHA256 = sha256Hex(manifestBytes)

	manifest, err := parseManifest(manifestBytes)
	if err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	result.BuildID = manifest.BuildID
	result.KeyID = manifest.KeyID
	result.CryptoManifestDigest = manifest.CryptoManifestDigest

	sigBytes, err := v.readFile(SignatureFile)
	if err != nil {
		return nil, fmt.Errorf("read signature: %w", err)
	}
	sig, err := parseSignature(sigBytes)
	if err != nil {
		return nil, fmt.Errorf("parse signature: %w", err)
	}

	if !strings.EqualFold(result.ManifestSHA256, sig.ManifestSHA256) {
		return result, fmt.Errorf("manifest sha256 mismatch: manifest=%s signature=%s", result.ManifestSHA256, sig.ManifestSHA256)
	}
	result.ManifestVerified = true

	if !strings.EqualFold(result.KeyID, sig.KeyID) {
		return result, fmt.Errorf("key id mismatch: manifest=%s signature=%s", result.KeyID, sig.KeyID)
	}

	cryptoManifestBytes, err := v.readFile(CryptoManifestFile)
	if err != nil {
		return nil, fmt.Errorf("read crypto manifest: %w", err)
	}
	cmDigest := sha256Hex(canonicalize(cryptoManifestBytes))
	if !strings.EqualFold(cmDigest, stripPrefix(manifest.CryptoManifestDigest)) {
		return result, fmt.Errorf("crypto manifest digest mismatch: computed=%s manifest=%s", cmDigest, manifest.CryptoManifestDigest)
	}

	pubKeyBytes, err := v.readFile(PublicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	pubKey, err := parseEd25519PublicKey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	result.KeyVerified = true

	sigDecoded, err := hex.DecodeString(sig.Signature)
	if err != nil {
		sigDecoded, err = base64.StdEncoding.DecodeString(sig.Signature)
		if err != nil {
			return result, fmt.Errorf("decode signature: %w", err)
		}
	}

	digestBytes, err := hex.DecodeString(stripPrefix(manifest.CryptoManifestDigest))
	if err != nil {
		return result, fmt.Errorf("decode digest for signing: %w", err)
	}
	if !ed25519.Verify(pubKey, digestBytes, sigDecoded) {
		return result, fmt.Errorf("ed25519 signature verification failed")
	}
	result.SignatureValid = true

	for _, artifact := range manifest.Artifacts {
		ar := types.ArtifactResult{
			Name:   artifact.Name,
			SHA256: artifact.SHA256,
			Size:   artifact.Size,
		}
		artifactBytes, err := v.readFile(artifact.Name)
		if err != nil {
			return result, fmt.Errorf("read artifact %s: %w", artifact.Name, err)
		}
		actualHash := sha256Hex(artifactBytes)
		actualSize := int64(len(artifactBytes))
		ar.SHA256OK = strings.EqualFold(actualHash, stripPrefix(artifact.SHA256))
		ar.SizeOK = actualSize == artifact.Size
		result.Artifacts = append(result.Artifacts, ar)
		if !ar.SHA256OK || !ar.SizeOK {
			return result, fmt.Errorf("artifact %s verification failed: sha256_ok=%v size_ok=%v (expected_size=%d actual_size=%d)",
				artifact.Name, ar.SHA256OK, ar.SizeOK, artifact.Size, actualSize)
		}
	}

	result.AllPassed = true
	return result, nil
}

func (v *Verifier) readFile(name string) ([]byte, error) {
	path := filepath.Join(v.kernelDir, name)
	// #nosec G304 - path is constructed from trusted kernel directory
	return os.ReadFile(path)
}

func parseManifest(data []byte) (*types.Manifest, error) {
	var m types.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func parseSignature(data []byte) (*types.Signature, error) {
	var s types.Signature
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func parseEd25519PublicKey(data []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block != nil {
		data = block.Bytes
	}
	pubKey, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, err
	}
	ed25519Key, ok := pubKey.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ed25519 public key")
	}
	return ed25519Key, nil
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func canonicalize(data []byte) []byte {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return data
	}
	canonical, err := json.Marshal(v)
	if err != nil {
		return data
	}
	return canonical
}

func stripPrefix(s string) string {
	if strings.HasPrefix(s, "sha256:") {
		return s[7:]
	}
	return s
}
