# uID0 Genesis Prototype (prana-uid)

Unique Identity system for the Prana Network. Handles entropy sourcing, biometric integration, and device-level attestation heuristics.

## Context & Ethics

This repository is part of the modular architecture of the Prana Network, a regenerative, decentralised and self-sustaining socio-technical system. Each module in this ecosystem is designed to function independently, while conforming to the shared protocols of the network.

This repository is subject to ethical use constraints. Refer to the central `ETHICS.md` document in the main project repository for terms of responsible usage.
- **Ethical Clause:** Obligation to inform the user about how their entropy is used. Explicit prohibition of data mining.

## Genesis Requirements and Security

The system is built sequentially to support real hardware interactions (via Bluetooth vascular sensors, audio capture for voice recognition, and FIDO2 capable fingerprint sensors), but defaults to a fully deterministic simulation mode for local development and verification.

It employs post-quantum resistant cryptography:
- Kyber1024 for Deterministic Key Encapsulation (used as KDF)
- Dilithium3 for Mutual Algorithmic Triad Signatures
- BLAKE3 for robust cryptographic hashing 

The project structure adheres to idomatic Go standards:
- `cmd/uid0/` - Main CLI application
- `internal/` - Internal core modules (crypto, entropy, biometric, genesis, token)
- `doc/` - Technical specifications
- `test/` - Text references and test artifacts
- `bin/` - Output directory for the compiled binaries
- `output/` - Directory to hold successfully generated CBOR tokens

## Compilation
   
```bash
go mod tidy
go build -o bin/uid0 ./cmd/uid0
```

## Usage

### Simulation Mode

Developers can run the application in a mocked hardware context securely determinized by a seed. In this mode, the RootID hash is visible.

```bash
./bin/uid0 --simulate --seed 99 --out ./output
```

### Production Mode

In production mode, the real physical ceremony will trigger hardware bindings and interactive flows. By security default, the system will obscure the generated RootID.

```bash
./bin/uid0 --ceremony --out ./output
```

### Verification Mode

You can verify any generated tokens securely using offline mechanisms. Add `--strict` to explicitly reject artifacts produced in simulation mode.

```bash
./bin/uid0 --verify ./output/uid0_XX.cbor
```
