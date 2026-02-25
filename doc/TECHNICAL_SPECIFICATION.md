# uID₀ Technical Specification

## 1. Cryptographic Primitives & Justifications

The uID₀ system utilizes a defense-in-depth approach relying exclusively on post-quantum safe primitives for asymmetric operations and leading-edge hashing algorithms for symmetric operations.

### 1.1 Key Encapsulation Mechanism (KEM): ML-KEM-1024 (Kyber1024)
- **Dependency**: `github.com/cloudflare/circl/kem/kyber/kyber1024`
- **Justification**: Kyber1024 (now standardized as ML-KEM-1024 by NIST) offers NIST Security Level 5 (equivalent to AES-256 in classical security strength). In the context of uID₀, we utilize Kyber's deterministic key generation (`DeriveKeyPair`) and deterministic encapsulation (`EncapsulateDeterministically`) seeded by BLAKE3 hashes of biological and cosmic entropy. This acts as a highly resilient, post-quantum secure Pseudo-Random Function (PRF) / Key Derivation Function (KDF). Circl's implementation is chosen for its constant-time execution and rigorous auditing.
- **Complexity**: Breaking Kyber1024 requires solving the Module Learning With Errors (M-LWE) problem. The estimated cost for a known-plaintext attack using the best known algorithms (primal/dual lattice attacks) exceeds $2^{256}$ classical operations and $2^{128}$ quantum operations.

### 1.2 Digital Signatures: ML-DSA-87 (Dilithium3)
- **Dependency**: `github.com/cloudflare/circl/sign/eddilithium3`
- **Justification**: Dilithium3 (standardized as ML-DSA-87) is the primary algorithm for the Algorithmic Triad (Recovery, Update, Audit) mutual signatures. It provides NIST Security Level 3 (equivalent to AES-192). It is chosen over ECDSA/Ed25519 due to Shor's algorithm threat. The Triad generates ephemeral keys, signs the shared context, and immediately zeroes the memory (`WipeSecret`), ensuring forward secrecy.
- **Complexity**: Relies on the hardness of M-LWE and Module Short Integer Solution (M-SIS). Forging a signature without the private key requires exponential time in the lattice dimension, practically unfeasible with current or foreseeable quantum computers.

### 1.3 Cryptographic Hashing: BLAKE3
- **Dependency**: `lukechampine.com/blake3`
- **Justification**: BLAKE3 is highly parallelizable, offering significant performance advantages over SHA-256 and SHA-3 while maintaining a large security margin. It is used for all internal digests, Merkle tree construction, and intermediate key derivation (`DeriveKey` mode). The `lukechampine/blake3` implementation is pure Go, highly optimized, and audited.
- **Complexity**: Collision resistance, preimage resistance, and second preimage resistance are all $2^{128}$ (for the 256-bit output used).

## 2. Mathematical Models & Resolutions

### 2.1 The Biological Merkle Tree
To aggregate the biometric state of the founders, a deterministic, lexicographically sorted Merkle tree is constructed. 

Let the biometric captures be $B_1, B_2, B_3$.
The BioHash for founder $i$ is calculated as:

$$H_{bio,i} = \text{KyberDigest}(B_i \parallel Sig_{F_i} \parallel \Delta t \parallel Score_{AI})$$

The leaves $L$ are sorted chronologically: 
$$L = \text{Sort}(\{H_{bio,1}, H_{bio,2}, H_{bio,3}\})$$

For an odd number of leaves (3), the last leaf is duplicated.
$$Node_{0,0} = \text{BLAKE3}(L_0 \parallel L_1)$$
$$Node_{0,1} = \text{BLAKE3}(L_2 \parallel L_2)$$
$$Root_{Merkle} = \text{BLAKE3}(Node_{0,0} \parallel Node_{0,1})$$

### 2.2 Fundamental Entropy ($E_F$)
To ensure the system is not reliant purely on deterministic physical sensor data, cosmic mechanics and symbolic truth are injected.

$$E_{cosmic} = \text{SiderealTime}(JD_{now})$$
$$E_{symbolic} = \text{BLAKE3}(Text_{Genesis})$$
$$E_F = \text{KyberDigest}(E_{cosmic} \parallel E_{symbolic})$$

We compute the Shannon Entropy $H(X)$ of $E_F$ in bits:
$$H(X) = - \sum_{i=0}^{255} P(x_i) \log_2 P(x_i)$$

The ceremony aborts if $H(X) < 50.0$ bits.

### 2.3 RootID Derivation
The ultimate identifier is a fusion of the environmental context and the Algorithmic Triad's consensus.
Let $M$ be the `TriadContext` struct bytes (containing $E_F, Root_{Merkle}, Rep, t_{sidereal}, \text{cycle}=0$).
$Sig_{R}, Sig_{U}, Sig_{A}$ are the Dilithium signatures from the Recovery, Update, and Audit Triad members over $M$.

$$RootID = \text{KyberDigest}(M \parallel Sig_R \parallel Sig_U \parallel Sig_A)$$

### 2.4 Deterministic Serialization
- **Dependency**: `github.com/fxamacker/cbor/v2`
- **Justification**: CBOR (RFC 8949) with Core Deterministic Encoding Requirements (RFC 8949 Section 4.2.1) ensures that the bytes representing the `UIDZeroSoulbound` token are strictly identical across architectures and compilations. `fxamacker/cbor` provides `CanonicalEncOptions()`.
- **Complexity Validation**: Because the CBOR serialization is strictly canonical, the `FinalDigest` (BLAKE3 of the CBOR representation sans the digest field itself) acts as an inviolable seal over the token's exact state.

## 3. Threat Model & Complexity to Crack uID₀

The uID₀ token is designed to act as the Genesis Seed for a sovereign identity network. Cracking it implies generating a colliding `RootID`, creating a valid counterfeit token, or extracting sensitive biological material from the CBOR output.

1. **Information Theoretic Indistinguishability**: The biological inputs are irreversibly mixed using `KyberDigest`. Given a `RootID`, `GenesisHash`, or `BioHash`, extracting the original fingerprint, voice, or vascular vein patterns has a probability of exactly $0$ due to the one-way nature of the lattice-based KDF and BLAKE3 function.
2. **Pre-image & Collision Attacks**: Generating a counterfeit token with the exact same `RootID` requires finding a pre-image collision under `Kyber1024` determinism. Complexity: $> 2^{256}$.
3. **Triad Compromise**: A counterfeit token must hold valid Dilithium3 signatures from the Triad. Because the Triad's private keys ($sk$) are ephemeral and zeroized (`WipeSecret`) immediately after signing in RAM, attacking the keys retroactively via memory dumps is strictly mitigated. An attacker would need to subvert the ring-0 kernel memory of the specific air-gapped machine orchestrating the ceremony in real-time.
4. **Liveness & Hardware Spoofing**: The biological captures require a tight $\Delta t \le 5000\text{ms}$ between challenge and sensor response, alongside a local AI confidence score $\ge 0.8$. This bounds the attack surface strictly to the localized physical presence of the founders preventing replay attacks.

## 4. Potential to Contribute to Identity Sovereignty

The `uID₀` protocol fundamentally shifts the paradigm of digital identity from a **state-issued or corporation-hosted model** to a **self-certified, mathematically absolute, mathematically sovereign model**.

1. **Zero-Knowledge Origin**: The `UIDZeroSoulbound` token mathematically proves that three distinct humans, verified by vascular biometric entropy and localized by a specific cosmic sidewatch (sidereal time), consented to bootstrap a network. It proves *humanity* and *uniqueness* without ever revealing *who* those humans are. There are no names, social numbers, or plaintext biometric vectors stored.
2. **Post-Quantum Foundation**: By abandoning elliptic curves (ECC), secp256k1, and RSA from day one, identity anchored to uID₀ is resilient against the pending cryptanalytic breaks by quantum computers (Shor's Algorithm). This guarantees multi-generational longevity of the identity root.
3. **The Algorithmic Triad**: The concept of Recovery, Update, and Audit keys introduces a decentralized social recovery and state-transition model directly into the token payload format. It models human trust boundaries mathematically. This allows for sovereign key rotation without relying on a centralized PKI (Public Key Infrastructure) certificate authority.
4. **Soulbound Concept**: The initial token is strictly non-transferable (soulbound). It forms the Absolute Origin Point ($0,0$) for the Prana Network's directed acyclic graph (DAG) or blockchain, defining the anchor of truth for all subsequent verifiable credentials issued to global citizens.

This mathematical, deterministic, and biological synthesis creates the ultimate form of sovereign digital matter.
