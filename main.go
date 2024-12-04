package main

import (
	"certutil/codec"
	"certutil/keypairgen"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"fmt"
	"path/filepath"
	"time"
)

type KeyPair struct {
	Type string
	Key  interface{}
}

func main() {
	startTime := time.Now()
	fmt.Printf("Main function started at: %s\n\n", startTime.Format(time.RFC3339))

	iterations := 5

	// Generate key pairs
	keyPairs := generateKeys(iterations)

	// Write key pairs to files
	writeKeys(keyPairs)

	// Read keys from files
	readKeys(keyPairs)

	fmt.Printf("\nTotal execution time: %d ms\n", time.Since(startTime).Milliseconds())
}

func generateKeys(iterations int) []KeyPair {
	fmt.Println("Generating key pairs:")

	rsaStart := time.Now()
	rsaPool := keypairgen.NewKeyPairPool(iterations, iterations, keypairgen.GenerateRSAKeyPair(2048))
	fmt.Printf("RSA KeyPairPool creation took: %d ms\n", time.Since(rsaStart).Milliseconds())
	defer rsaPool.Close()

	ecStart := time.Now()
	ecPool := keypairgen.NewKeyPairPool(iterations, 1, keypairgen.GenerateECKeyPair(elliptic.P256()))
	fmt.Printf("EC KeyPairPool creation took: %d ms\n", time.Since(ecStart).Milliseconds())
	defer ecPool.Close()

	edStart := time.Now()
	edPool := keypairgen.NewKeyPairPool(iterations, 1, keypairgen.GenerateEDKeyPair("Ed25519"))
	fmt.Printf("ED KeyPairPool creation took: %d ms\n\n", time.Since(edStart).Milliseconds())
	defer edPool.Close()

	keyPairs := make([]KeyPair, 0, iterations*3)

	for i := 0; i < iterations; i++ {
		fmt.Printf("Iteration %d:\n", i+1)

		start := time.Now()
		rsaKey := rsaPool.Get().(*rsa.PrivateKey)
		rsaDuration := time.Since(start).Milliseconds()
		fmt.Printf("Retrieved RSA Key in %d ms\n", rsaDuration)
		keyPairs = append(keyPairs, KeyPair{"rsa2048", rsaKey})

		start = time.Now()
		ecKey := ecPool.Get().(*ecdsa.PrivateKey)
		ecDuration := time.Since(start).Milliseconds()
		fmt.Printf("Retrieved EC Key in %d ms\n", ecDuration)
		keyPairs = append(keyPairs, KeyPair{"ecp256", ecKey})

		start = time.Now()
		edKey := edPool.Get().(ed25519.PrivateKey)
		edDuration := time.Since(start).Milliseconds()
		fmt.Printf("Retrieved ED Key in %d ms\n", edDuration)
		keyPairs = append(keyPairs, KeyPair{"ed25519", edKey})

		fmt.Println()
	}

	return keyPairs
}

func writeKeys(keyPairs []KeyPair) {
	fmt.Println("Writing key pairs to files:")
	start := time.Now()

	for i, kp := range keyPairs {
		baseFilename := filepath.Join("output", fmt.Sprintf("keypair_%d_%s", i+1, kp.Type))

		err := codec.WriteKeyToPEMFile(kp.Key, baseFilename+"_pri.pem", false)
		logFileOperation("write", baseFilename+"_pri.pem", err)

		err = codec.WriteKeyToDERFile(kp.Key, baseFilename+"_pri.der", false)
		logFileOperation("write", baseFilename+"_pri.der", err)

		var pubKey interface{}
		switch k := kp.Key.(type) {
		case *rsa.PrivateKey:
			pubKey = &k.PublicKey
		case *ecdsa.PrivateKey:
			pubKey = &k.PublicKey
		case ed25519.PrivateKey:
			pubKey = k.Public()
		}

		err = codec.WriteKeyToPEMFile(pubKey, baseFilename+"_pub.pem", true)
		logFileOperation("write", baseFilename+"_pub.pem", err)

		err = codec.WriteKeyToDERFile(pubKey, baseFilename+"_pub.der", true)
		logFileOperation("write", baseFilename+"_pub.der", err)
	}

	duration := time.Since(start).Milliseconds()
	fmt.Printf("Total time to write all key pairs: %d ms\n\n", duration)
}

func readKeys(keyPairs []KeyPair) {
	fmt.Println("Reading key pairs from files:")
	start := time.Now()

	for i, kp := range keyPairs {
		baseFilename := filepath.Join("output", fmt.Sprintf("keypair_%d_%s", i+1, kp.Type))

		_, err := codec.ReadKeyFromPEMFile(baseFilename + "_pri.pem")
		logFileOperation("read", baseFilename+"_pri.pem", err)

		_, err = codec.ReadKeyFromDERFile(baseFilename + "_pri.der")
		logFileOperation("read", baseFilename+"_pri.der", err)

		_, err = codec.ReadKeyFromPEMFile(baseFilename + "_pub.pem")
		logFileOperation("read", baseFilename+"_pub.pem", err)

		_, err = codec.ReadKeyFromDERFile(baseFilename + "_pub.der")
		logFileOperation("read", baseFilename+"_pub.der", err)
	}

	duration := time.Since(start).Milliseconds()
	fmt.Printf("Total time to read all key pairs: %d ms\n", duration)
}

func logFileOperation(operation, filename string, err error) {
	if err != nil {
		fmt.Printf("Error %sing %s: %v\n", operation, filename, err)
	} else {
		fmt.Printf("Successfully %s %s\n", operation, filename)
	}
}
