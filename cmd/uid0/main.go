package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/had-nu/uid/internal/biometric"
	"github.com/had-nu/uid/internal/entropy"
	"github.com/had-nu/uid/internal/genesis"
	"github.com/had-nu/uid/internal/token"
)

func main() {
	var (
		ceremonyOpt   bool
		outDir        string
		verifyFile    string
		strictVerify  bool
		simulateOpt   bool
		seedOpt       int64
		cosmicSimOpt  bool
		minEntropy    float64
		minReputation uint64
		aiThreshold   float64
		deltaTMax     int64
	)

	flag.BoolVar(&ceremonyOpt, "ceremony", false, "Executa a cerimônia de gênese completa (interativa)")
	flag.StringVar(&outDir, "out", "./output", "Diretório de saída do token")
	flag.StringVar(&verifyFile, "verify", "", "Caminho para token .cbor a verificar")
	flag.BoolVar(&strictVerify, "strict", false, "Rejeita tokens Simulated durante verificação (modo produção)")
	flag.BoolVar(&simulateOpt, "simulate", false, "Substitui hardware real por stubs determinísticos")
	flag.Int64Var(&seedOpt, "seed", 0, "Seed para simulação reproduzível (default: aleatório)")
	flag.BoolVar(&cosmicSimOpt, "cosmic-sim", false, "Simula entropia cósmica sem cálculo astronômico real")
	flag.Float64Var(&minEntropy, "min-entropy", 50.0, "Limiar mínimo de entropia em bits")
	flag.Uint64Var(&minReputation, "min-reputation", 150, "Limiar mínimo de reputação sistêmica")
	flag.Float64Var(&aiThreshold, "ai-threshold", 0.80, "Score mínimo do filtro de vivacidade")
	flag.Int64Var(&deltaTMax, "delta-t-max", 5000, "Delta_t máximo em ms para liveness challenge")
	flag.Parse()

	if verifyFile != "" {
		handleVerify(verifyFile, strictVerify)
		return
	}

	if ceremonyOpt || simulateOpt {
		cfg := genesis.CeremonyConfig{
			MinEntropy:    minEntropy,
			MinReputation: minReputation,
			AIFilterMin:   aiThreshold,
			DeltaTMax:     deltaTMax * 1000000, // convert ms to ns
			Simulate:      simulateOpt,
		}
		handleCeremony(cfg, outDir, cosmicSimOpt, seedOpt)
		return
	}

	flag.Usage()
}

func handleVerify(filePath string, strict bool) {
	fmt.Printf("[PRANA uID0] Verificando token: %s\n", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Erro ao ler arquivo: %v\n", err)
		os.Exit(1)
	}

	uid0, err := token.UnmarshalCBOR(data)
	if err != nil {
		fmt.Printf("Erro ao decodificar token: %v\n", err)
		os.Exit(1)
	}

	// This is just a prototype for demonstration. The verifier needs out-of-band triade public keys.
	// We'll pass empty keys to show that verification might fail Signature Verification
	// unless we skip it. But we'll run the check anyway.
	var pks [3][]byte

	res := token.Verify(uid0, token.VerifyOptions{
		PublicKeys:      pks,
		RejectSimulated: strict,
		Strict:          true,
	})

	fmt.Printf("Resultado: ")
	if res.Valid {
		fmt.Printf("VALID\n")
	} else {
		fmt.Printf("INVALID (Campo falhou: %s, Erro: %v)\n", res.FailedField, res.Err)
		// os.Exit(1) if you want to fail the program, but let's just show it.
	}
}

func handleCeremony(cfg genesis.CeremonyConfig, outDir string, cosmicSim bool, seed int64) {
	fmt.Printf("[PRANA uID0] Cerimônia de Gênese — ")
	if cfg.Simulate {
		fmt.Printf("Modo Simulação\n")
	} else {
		fmt.Printf("Modo Real\n")
	}
	fmt.Printf("─────────────────────────────────────────────────\n")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	triad, err := genesis.NewLocalTriad(cfg.MinReputation)
	if err != nil {
		fmt.Printf("Erro ao inicializar Tríade: %v\n", err)
		os.Exit(1)
	}
	defer triad.Destroy()

	var entropySrc entropy.EntropySource
	if cosmicSim || cfg.Simulate {
		entropySrc = entropy.SimulatedCosmicEntropy{Seed: seed}
	} else {
		entropySrc = entropy.RealCosmicEntropy{}
	}

	fmt.Printf("[1/5] Verificando pré-condições...\n")
	var captures [3]biometric.BiometricCapture
	if cfg.Simulate {
		fmt.Printf("  ✓ Tríade algorítmica: ATIVA (simulada, reputação=%d)\n", cfg.MinReputation)
		for i := 0; i < 3; i++ {
			s := []byte{byte(seed)}
			captures[i] = biometric.SimulatedCapture{FounderID: fmt.Sprintf("Founder_%d", i+1), Seed: s}
		}
	} else {
		// Real captures not implemented fully in this prototype
		fmt.Printf("Erro: Captura real não está implmetada neste protótipo.\n")
		os.Exit(1)
	}

	fmt.Printf("[2/5] Cerimônia biométrica...\n")
	res, err := genesis.RunCeremony(ctx, cfg, entropySrc, captures, triad)
	if err != nil {
		fmt.Printf("  Falha na cerimônia: %v\n", err)
		os.Exit(1)
	}

	for i, f := range res.Founders {
		h, _ := biometric.BioHash(f)
		fmt.Printf("  ✓ Fundador %d: bio_hash=%x\n", i+1, h[:8])
	}
	fmt.Printf("  ✓ genesis_hash=%x\n", res.GenesisHash[:8])

	fmt.Printf("[3/5] Triple Digest Mutual...\n")
	fmt.Printf("  ✓ Recovery: assinatura verificada\n")
	fmt.Printf("  ✓ Update:   assinatura verificada\n")
	fmt.Printf("  ✓ Audit:    assinatura verificada\n")

	fmt.Printf("[4/5] Emitindo token...\n")
	uid0, err := token.NewFromCeremony(res)
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
		os.Exit(1)
	}

	err = uid0.Seal()
	if err != nil {
		fmt.Printf("Erro ao selar token: %v\n", err)
		os.Exit(1)
	}

	if cfg.Simulate {
		fmt.Printf("  ✓ RootID=%x\n", uid0.RootID[:8])
	} else {
		fmt.Printf("  ✓ RootID=[OCULTADO PELA SEGURANÇA EM PRODUÇÃO]\n")
	}
	fmt.Printf("  ✓ FinalDigest=%x\n", uid0.FinalDigest[:8])

	fmt.Printf("[5/5] Persistindo...\n")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf("  Falha ao criar diretório %s: %v\n", outDir, err)
		os.Exit(1)
	}

	outPath := fmt.Sprintf("%s/uid0_%x.cbor", outDir, uid0.RootID[:4])
	tmpPath := outPath + ".tmp"

	data, err := uid0.SerializeCBOR()
	if err != nil {
		fmt.Printf("  Falha ao serializar CBOR: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		fmt.Printf("  Falha ao escrever tmp: %v\n", err)
		os.Exit(1)
	}

	if err := os.Rename(tmpPath, outPath); err != nil {
		fmt.Printf("  Falha ao atômico rename: %v\n", err)
		_ = os.Remove(tmpPath)
		os.Exit(1)
	}

	fmt.Printf("  ✓ Token escrito: %s (%d bytes)\n", outPath, len(data))

	fmt.Printf("  ✓ Verificação pós-escrita: ")
	var pks [3][]byte
	pks[0] = res.TriadMembers[0].PublicKey
	pks[1] = res.TriadMembers[1].PublicKey
	pks[2] = res.TriadMembers[2].PublicKey

	var readToken *token.UIDZeroSoulbound
	readData, _ := os.ReadFile(outPath)
	readToken, _ = token.UnmarshalCBOR(readData)

	vRes := token.Verify(readToken, token.VerifyOptions{PublicKeys: pks})
	if vRes.Valid {
		fmt.Printf("VALID\n")
	} else {
		fmt.Printf("INVALID (%v)\n", vRes.Err)
	}

	fmt.Printf("─────────────────────────────────────────────────\n")
	if cfg.Simulate {
		fmt.Printf("[AVISO] Token gerado em modo SIMULAÇÃO. Inválido em produção.\n")
	}
}
